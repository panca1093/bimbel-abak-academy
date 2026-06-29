"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import {
  ClipboardList,
  ListChecks,
  Package,
  Pencil,
  Plus,
  Trash2,
  Trophy,
  Users,
} from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { ExamModal } from "@/components/admin/ExamModal";
import { UnderMaintenance } from "@/components/admin/UnderMaintenance";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useExam,
  usePublishExam,
  useReplaceExamTests,
  useUpdateExamPrice,
} from "@/lib/hooks/admin-exams";
import { useAdminTests } from "@/lib/hooks/admin-tests";
import { useTranslation } from "@/lib/i18n";
import { formatRupiah } from "@/lib/format";

type Tab =
  | "overview"
  | "tests"
  | "price"
  | "registrations"
  | "results"
  | "grading"
  | "leaderboard";

const TAB_ORDER: Tab[] = [
  "overview",
  "tests",
  "price",
  "registrations",
  "results",
  "grading",
  "leaderboard",
];

function errorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error) return err.message;
  return fallback;
}

function formatScheduled(iso?: string | null): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" });
}

function statusBadgeClass(status?: string): string {
  switch (status) {
    case "published":
      return "bg-green-100 text-green-800 border-green-200";
    case "draft":
      return "bg-line-2 text-ink-700 border-line";
    case "hidden":
      return "bg-amber-100 text-amber-800 border-amber-200";
    case "archived":
      return "bg-red-100 text-red-800 border-red-200";
    default:
      return "bg-line-2 text-ink-700 border-line";
  }
}

export default function ExamPackageDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params?.id ?? "";
  const { t } = useTranslation();

  const [tab, setTab] = useState<Tab>("overview");
  const [editOpen, setEditOpen] = useState(false);

  const { data, isLoading, isError, error, refetch } = useExam(id);
  const replaceTests = useReplaceExamTests(id);
  const updatePrice = useUpdateExamPrice(id);
  const publish = usePublishExam(id);
  const { data: availableResp, isLoading: availableLoading } = useAdminTests();
  const availableTests = availableResp?.data ?? [];

  const [attachedIds, setAttachedIds] = useState<string[]>([]);
  const [priceInput, setPriceInput] = useState("");

  useEffect(() => {
    if (!data) return;
    setAttachedIds(data.tests.map((entry) => entry.test_id));
  }, [data]);

  useEffect(() => {
    if (!data) return;
    setPriceInput(String(data.product_price ?? 0));
  }, [data]);

  const availableToAdd = useMemo(() => {
    const attached = new Set(attachedIds);
    return availableTests.filter((test) => !attached.has(test.id));
  }, [availableTests, attachedIds]);

  function handleAddTest(testId: string) {
    setAttachedIds((prev) => [...prev, testId]);
  }

  function handleRemoveTest(testId: string) {
    setAttachedIds((prev) => prev.filter((entry) => entry !== testId));
  }

  async function handleSaveTests() {
    if (!id) return;
    try {
      await replaceTests.mutateAsync(attachedIds);
      toast.success(t("changes_saved"));
      refetch();
    } catch (e) {
      toast.error(errorMessage(e, t("error_generic")));
    }
  }

  async function handleSavePrice() {
    const next = Number(priceInput);
    if (!Number.isFinite(next) || next < 0) {
      toast.error(t("error_generic"));
      return;
    }
    try {
      await updatePrice.mutateAsync(next);
      toast.success(t("changes_saved"));
      refetch();
    } catch (e) {
      toast.error(errorMessage(e, t("error_generic")));
    }
  }

  async function handlePublish() {
    if (!confirm(t("admin_exam_detail_publish_confirm"))) return;
    try {
      await publish.mutateAsync();
      toast.success(t("changes_saved"));
      refetch();
    } catch (e) {
      toast.error(errorMessage(e, t("error_generic")));
    }
  }

  const title = data?.title ?? t("exam_packages_page_title");
  const description = data
    ? `${formatScheduled(data.scheduled_at)} · ${data.product_status ?? "draft"}`
    : undefined;

  return (
    <div className="space-y-6 fade-in">
      <AdminPageHeader
        icon={Package}
        title={title}
        description={description}
      />

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </div>
      )}

      {isError && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
          {errorMessage(error, t("error_generic"))}
        </div>
      )}

      {!isLoading && !isError && data && (
        <>
          <div className="flex flex-wrap gap-1 border-b">
            {TAB_ORDER.map((key) => (
              <button
                key={key}
                type="button"
                onClick={() => setTab(key)}
                className={
                  tab === key
                    ? "border-b-2 border-primary px-3 py-2 text-sm font-medium text-primary"
                    : "px-3 py-2 text-sm font-medium text-muted-foreground hover:text-foreground"
                }
              >
                {t(`admin_exam_detail_tab_${key}` as const)}
              </button>
            ))}
          </div>

          {tab === "overview" && (
            <div className="md-card-outlined space-y-4 p-6">
              <div className="flex items-center justify-between">
                <h2 className="text-title-large font-semibold">
                  {t("admin_exam_detail_tab_overview")}
                </h2>
                <Button variant="outline" size="sm" onClick={() => setEditOpen(true)}>
                  <Pencil className="mr-1 size-4" />
                  {t("admin_exam_detail_edit")}
                </Button>
              </div>
              <dl className="grid grid-cols-1 gap-4 text-sm sm:grid-cols-2">
                <OverviewRow label="Title" value={data.title} />
                <OverviewRow label="Scheduled" value={formatScheduled(data.scheduled_at)} />
                <OverviewRow label="Timer mode" value={data.timer_mode ?? "—"} />
                <OverviewRow
                  label="Duration"
                  value={
                    data.duration_minutes != null
                      ? `${data.duration_minutes} ${t("minutes")}`
                      : "—"
                  }
                />
                <OverviewRow
                  label="Free"
                  value={data.is_free ? t("status_label_active") : t("status_label_inactive")}
                />
                <OverviewRow
                  label="Requires check-in"
                  value={
                    data.requires_checkin
                      ? t("status_label_active")
                      : t("status_label_inactive")
                  }
                />
                <OverviewRow
                  label="Leaderboard"
                  value={
                    data.allow_leaderboard
                      ? t("status_label_active")
                      : t("status_label_inactive")
                  }
                />
                <OverviewRow
                  label="Randomize"
                  value={
                    data.randomize
                      ? t("status_label_active")
                      : t("status_label_inactive")
                  }
                />
                <OverviewRow label="Status" value={data.status ?? "—"} />
                <OverviewRow
                  label="Product status"
                  value={
                    <Badge className={statusBadgeClass(data.product_status)}>
                      {data.product_status ?? "draft"}
                    </Badge>
                  }
                />
                <OverviewRow label="Price" value={formatRupiah(data.product_price ?? 0)} />
              </dl>
            </div>
          )}

          {tab === "tests" && (
            <div className="grid gap-4 lg:grid-cols-2">
              <div className="md-card-outlined p-4">
                <div className="mb-3 flex items-center justify-between">
                  <h3 className="text-title-medium font-semibold">
                    {t("admin_exam_detail_tests_attached")}
                  </h3>
                  <span className="text-label text-muted-foreground">
                    {attachedIds.length}
                  </span>
                </div>
                {attachedIds.length === 0 ? (
                  <div className="rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground">
                    —
                  </div>
                ) : (
                  <ul className="space-y-2">
                    {attachedIds.map((testId, idx) => {
                      const meta = data.tests.find((e) => e.test_id === testId)?.test;
                      return (
                        <li
                          key={`${testId}-${idx}`}
                          className="flex items-center justify-between gap-2 rounded-md border p-3 text-sm"
                        >
                          <div className="min-w-0">
                            <div className="truncate font-medium">
                              #{idx + 1} · {meta?.title ?? testId}
                            </div>
                            {meta && (
                              <div className="text-label text-muted-foreground">
                                {meta.subject} · {meta.topic ?? "—"} ·{" "}
                                {meta.question_count ?? 0} soal ·{" "}
                                {meta.duration_minutes ?? 0} {t("minutes")}
                              </div>
                            )}
                          </div>
                          <Button
                            type="button"
                            size="icon-xs"
                            variant="ghost"
                            onClick={() => handleRemoveTest(testId)}
                            aria-label={t("admin_exam_detail_tests_remove")}
                          >
                            <Trash2 className="size-3" />
                          </Button>
                        </li>
                      );
                    })}
                  </ul>
                )}
                <div className="mt-4 flex justify-end">
                  <Button
                    type="button"
                    onClick={handleSaveTests}
                    disabled={replaceTests.isPending}
                  >
                    {replaceTests.isPending
                      ? t("saving")
                      : t("admin_exam_detail_tests_save")}
                  </Button>
                </div>
              </div>

              <div className="md-card-outlined p-4">
                <h3 className="text-title-medium mb-3 font-semibold">
                  {t("admin_exam_detail_tests_available")}
                </h3>
                {availableLoading ? (
                  <div className="space-y-2">
                    {Array.from({ length: 3 }).map((_, i) => (
                      <Skeleton key={i} className="h-10 w-full" />
                    ))}
                  </div>
                ) : availableToAdd.length === 0 ? (
                  <div className="rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground">
                    —
                  </div>
                ) : (
                  <ul className="space-y-2">
                    {availableToAdd.map((test) => (
                      <li
                        key={test.id}
                        className="flex items-center justify-between gap-2 rounded-md border p-3 text-sm"
                      >
                        <div className="min-w-0">
                          <div className="truncate font-medium">{test.title}</div>
                          <div className="text-label text-muted-foreground">
                            {test.subject} · {test.topic} · {test.duration_minutes}{" "}
                            {t("minutes")}
                          </div>
                        </div>
                        <Button
                          type="button"
                          size="icon-xs"
                          variant="outline"
                          onClick={() => handleAddTest(test.id)}
                          aria-label={t("admin_exam_detail_tests_add")}
                        >
                          <Plus className="size-3" />
                        </Button>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          )}

          {tab === "price" && (
            <div className="md-card-outlined space-y-6 p-6">
              <div className="space-y-3">
                <h2 className="text-title-large font-semibold">
                  {t("admin_exam_detail_price_label")}
                </h2>
                <div className="flex items-end gap-3">
                  <div className="grid flex-1 gap-2">
                    <Label htmlFor="exam-price">
                      {t("admin_exam_detail_price_label")}
                    </Label>
                    <Input
                      id="exam-price"
                      type="number"
                      min={0}
                      value={priceInput}
                      onChange={(e) => setPriceInput(e.target.value)}
                      disabled={updatePrice.isPending}
                    />
                  </div>
                  <Button
                    type="button"
                    onClick={handleSavePrice}
                    disabled={updatePrice.isPending}
                  >
                    {updatePrice.isPending
                      ? t("saving")
                      : t("admin_exam_detail_price_save")}
                  </Button>
                </div>
              </div>

              <div className="border-t pt-6">
                <h2 className="text-title-large font-semibold">
                  {t("admin_exam_detail_publish")}
                </h2>
                <p className="text-body-medium mt-1 text-muted-foreground">
                  {data.product_status ?? "draft"}
                </p>
                <Button
                  type="button"
                  className="mt-3"
                  onClick={handlePublish}
                  disabled={publish.isPending || data.product_status === "published"}
                >
                  {t("admin_exam_detail_publish")}
                </Button>
              </div>
            </div>
          )}

          {tab === "registrations" && (
            <UnderMaintenance icon={Users} title={t("admin_exam_detail_tab_registrations")} />
          )}
          {tab === "results" && (
            <UnderMaintenance icon={ListChecks} title={t("admin_exam_detail_tab_results")} />
          )}
          {tab === "grading" && (
            <UnderMaintenance icon={ClipboardList} title={t("admin_exam_detail_tab_grading")} />
          )}
          {tab === "leaderboard" && (
            <UnderMaintenance icon={Trophy} title={t("admin_exam_detail_tab_leaderboard")} />
          )}
        </>
      )}

      <ExamModal
        open={editOpen}
        exam={data ?? null}
        onClose={() => setEditOpen(false)}
        onSaved={() => {
          setEditOpen(false);
          refetch();
        }}
      />
    </div>
  );
}

function OverviewRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <dt className="text-label text-muted-foreground">{label}</dt>
      <dd className="text-sm">{value}</dd>
    </div>
  );
}