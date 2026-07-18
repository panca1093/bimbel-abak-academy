"use client";

import { useState, useMemo } from "react";
import { ShoppingCart, Loader2, CheckCircle } from "lucide-react";
import { toast } from "sonner";
import { useTranslation } from "@/lib/i18n";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { ParticipantPicker } from "@/components/admin/ParticipantPicker";
import { SnapCheckout } from "@/components/cart/SnapCheckout";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { formatRupiah } from "@/lib/format";
import {
  useOrderableExams,
  usePreviewBulkExamOrder,
  useCreateBulkExamOrder,
} from "@/lib/hooks/admin-bulk-exam-orders";
import { useAuthStore } from "@/stores/auth";
import type { ExamListItem } from "@/lib/types";

export default function BulkExamOrderPage() {
  const { t } = useTranslation();

  // Current school (admin_school's JWT carries school_id)
  const currentUser = useAuthStore((s) => s.user);
  const schoolId = currentUser?.school_id;

  // Step state
  const [selectedExamId, setSelectedExamId] = useState<string>("");
  const [selectedStudentIds, setSelectedStudentIds] = useState<string[]>([]);
  const [createdOrderId, setCreatedOrderId] = useState<string | null>(null);

  // Hooks
  const { data: examsData, isLoading: examsLoading } = useOrderableExams();
  const previewMutation = usePreviewBulkExamOrder();
  const createMutation = useCreateBulkExamOrder();

  const exams: ExamListItem[] = useMemo(
    () => examsData?.data ?? [],
    [examsData],
  );

  const selectedExam = useMemo(
    () => exams.find((e) => e.id === selectedExamId),
    [exams, selectedExamId],
  );

  // Build preview request
  const previewInput = useMemo(() => {
    if (!selectedExamId || selectedStudentIds.length === 0) return null;
    return { exam_id: selectedExamId, student_ids: selectedStudentIds };
  }, [selectedExamId, selectedStudentIds]);

  const hasPreview =
    previewMutation.data && previewInput !== null &&
    previewInput.exam_id === selectedExamId;

  const handlePreview = () => {
    if (!selectedExamId) {
      toast.error(t("bulk_exam_order_empty_exam"));
      return;
    }
    if (selectedStudentIds.length === 0) {
      toast.error(t("bulk_exam_order_empty_students"));
      return;
    }
    previewMutation.mutate(
      { exam_id: selectedExamId, student_ids: selectedStudentIds },
      {
        onError: () => toast.error(t("bulk_exam_order_preview_failed")),
      },
    );
  };

  const handleCreateOrder = () => {
    if (!previewInput) return;
    createMutation.mutate(previewInput, {
      onSuccess: (order) => {
        setCreatedOrderId(order.id);
        toast.success(t("bulk_exam_order_created"));
      },
      onError: (err) => {
        const msg =
          err instanceof Error
            ? err.message
            : t("bulk_exam_order_creating_failed");
        toast.error(msg);
      },
    });
  };

  const handleReset = () => {
    setSelectedExamId("");
    setSelectedStudentIds([]);
    setCreatedOrderId(null);
    previewMutation.reset();
    createMutation.reset();
  };

  // ── Success state ──────────────────────────────────────────────────────

  if (createdOrderId) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader
          icon={CheckCircle}
          title={t("bulk_exam_order_title")}
          description={t("bulk_exam_order_created_desc")}
        />

        <div className="md-card-outlined p-6 text-center">
          <CheckCircle className="mx-auto mb-4 size-12 text-success" />
          <h2 className="font-serif text-xl font-bold text-ink-900">
            {t("bulk_exam_order_created")}
          </h2>
          <p className="mt-2 text-sm text-ink-500">
            {selectedExam?.title} &middot; {selectedStudentIds.length} {t("bulk_exam_order_students_count").replace("{n}", String(selectedStudentIds.length))}
          </p>

          <div className="mt-8 flex flex-col items-center gap-4 sm:flex-row sm:justify-center">
            <SnapCheckout
              orderId={createdOrderId}
              basePath="/admin/bulk-exam-orders"
            />
            <Button variant="outline" onClick={handleReset}>
              {t("bulk_exam_order_select_exam")}
            </Button>
          </div>
        </div>
      </div>
    );
  }

  // ── Main form ──────────────────────────────────────────────────────────

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={ShoppingCart}
        title={t("bulk_exam_order_title")}
        description={t("bulk_exam_order_subtitle")}
      />

      {/* Step 1: Pick exam */}
      <section className="mb-8">
        <h3 className="font-serif text-base font-semibold text-ink-900">
          1. {t("bulk_exam_order_select_exam")}
        </h3>
        <div className="mt-2">
          {examsLoading ? (
            <div className="h-9 w-[280px] animate-pulse rounded-md bg-surface-2" />
          ) : (
            <Select value={selectedExamId} onValueChange={setSelectedExamId}>
              <SelectTrigger className="h-9 w-[280px] text-xs">
                <SelectValue
                  placeholder={t("bulk_exam_order_select_exam")}
                />
              </SelectTrigger>
              <SelectContent>
                {exams.map((e) => (
                  <SelectItem key={e.id} value={e.id}>
                    {e.title}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
          {!examsLoading && exams.length === 0 && (
            <p className="mt-1 text-xs text-ink-500">
              {t("empty_products")}
            </p>
          )}
        </div>
      </section>

      {/* Step 2: Pick participants */}
      {selectedExamId && (
        <section className="mb-8">
          <h3 className="font-serif text-base font-semibold text-ink-900">
            2. {t("bulk_exam_order_pick_participants")}
          </h3>
          <div className="mt-3">
            <ParticipantPicker
              schoolId={schoolId}
              selected={selectedStudentIds}
              onChange={setSelectedStudentIds}
            />
          </div>
        </section>
      )}

      {/* Preview & create */}
      {selectedStudentIds.length > 0 && (
        <div className="space-y-4">
          <Button
            size="lg"
            onClick={handlePreview}
            disabled={previewMutation.isPending}
          >
            {previewMutation.isPending ? (
              <Loader2 className="mr-2 size-4 animate-spin" />
            ) : null}
            {t("bulk_exam_order_preview")}
          </Button>

          {/* Preview panel */}
          {hasPreview && (
            <div className="md-card-outlined space-y-4 p-5">
              <h4 className="font-serif text-base font-semibold text-ink-900">
                {t("bulk_exam_order_preview_title")}
              </h4>
              <div className="flex items-center gap-2">
                <span className="text-sm text-ink-600">
                  {selectedExam?.title}
                </span>
                <Badge variant="outline">
                  {previewMutation.data.net_new_count}{" "}
                  {t("bulk_exam_order_students_count").replace(
                    "{n}",
                    String(previewMutation.data.net_new_count),
                  )}
                </Badge>
              </div>

              {previewMutation.data.excluded.length > 0 && (
                <div className="max-h-[160px] overflow-y-auto rounded-lg border border-line p-2">
                  {previewMutation.data.excluded.map((s) => (
                    <div
                      key={s.student_id}
                      className="flex items-center gap-2 px-2 py-1.5 text-sm"
                    >
                      <span className="font-medium text-ink-900">
                        {s.name}
                      </span>
                      <span className="text-ink-500">({s.reason})</span>
                    </div>
                  ))}
                </div>
              )}

              <div className="border-t border-line pt-3">
                <div className="flex items-center justify-between text-sm">
                  <span className="font-semibold text-ink-900">
                    {t("bulk_exam_order_total")}
                  </span>
                  <span className="font-serif text-lg font-bold text-success">
                    {formatRupiah(previewMutation.data.total)}
                  </span>
                </div>
              </div>

              <Button
                size="lg"
                className="w-full"
                onClick={handleCreateOrder}
                disabled={createMutation.isPending}
              >
                {createMutation.isPending ? (
                  <Loader2 className="mr-2 size-4 animate-spin" />
                ) : null}
                {createMutation.isPending
                  ? t("bulk_exam_order_confirming")
                  : t("bulk_exam_order_confirm")}
              </Button>
            </div>
          )}

          {/* Preview error */}
          {previewMutation.isError && (
            <p className="text-sm text-danger">
              {t("bulk_exam_order_preview_failed")}
            </p>
          )}
        </div>
      )}

      {!selectedExamId && (
        <div className="py-12 text-center text-sm text-ink-500">
          {t("bulk_exam_order_no_exam")}
        </div>
      )}
    </div>
  );
}
