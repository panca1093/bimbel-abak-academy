"use client";

import { useMemo, useState } from "react";
import { CheckCircle, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { useTranslation } from "@/lib/i18n";
import { ParticipantPicker } from "@/components/admin/ParticipantPicker";
import { SnapCheckout } from "@/components/cart/SnapCheckout";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { formatRupiah } from "@/lib/format";
import {
  usePreviewBulkExamOrder,
  useCreateBulkExamOrder,
} from "@/lib/hooks/admin-bulk-exam-orders";
import { useGrantExamAccess } from "@/lib/hooks/admin-exam-grants";
import { useAuthStore } from "@/stores/auth";

interface ExamRegistrationsTabProps {
  examId: string;
  examName: string;
}

// admin_school orders (Midtrans-paid); super_admin grants directly (no order,
// no payment). Same participant pool, submit action branches by role.
export function ExamRegistrationsTab({ examId, examName }: ExamRegistrationsTabProps) {
  const { t } = useTranslation();
  const role = useAuthStore((s) => s.user?.role);
  const schoolId = useAuthStore((s) => s.user?.school_id);
  const isSuperAdmin = role === "super_admin";

  const [selectedStudentIds, setSelectedStudentIds] = useState<string[]>([]);
  const [createdOrderId, setCreatedOrderId] = useState<string | null>(null);
  const [grantResult, setGrantResult] = useState<{
    granted_count: number;
    granted_students: Array<{ id: string; name: string; username: string }>;
  } | null>(null);

  const previewMutation = usePreviewBulkExamOrder();
  const createMutation = useCreateBulkExamOrder();
  const grantMutation = useGrantExamAccess();

  const previewInput = useMemo(() => {
    if (selectedStudentIds.length === 0) return null;
    return { exam_id: examId, student_ids: selectedStudentIds };
  }, [examId, selectedStudentIds]);

  const handlePreview = () => {
    if (selectedStudentIds.length === 0) {
      toast.error(t("bulk_exam_order_empty_students"));
      return;
    }
    previewMutation.mutate(
      { exam_id: examId, student_ids: selectedStudentIds },
      { onError: () => toast.error(t("bulk_exam_order_preview_failed")) },
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
          err instanceof Error ? err.message : t("bulk_exam_order_creating_failed");
        toast.error(msg);
      },
    });
  };

  const handleGrant = () => {
    if (selectedStudentIds.length === 0) {
      toast.error(t("exam_grant_empty_students"));
      return;
    }
    grantMutation.mutate(
      { exam_id: examId, student_ids: selectedStudentIds },
      {
        onSuccess: (result) => {
          setGrantResult(result);
          toast.success(t("exam_grant_success"));
        },
        onError: (err) => {
          const msg = err instanceof Error ? err.message : t("error_generic");
          toast.error(msg);
        },
      },
    );
  };

  const handleReset = () => {
    setSelectedStudentIds([]);
    setCreatedOrderId(null);
    setGrantResult(null);
    previewMutation.reset();
    createMutation.reset();
    grantMutation.reset();
  };

  // ── Success states ──────────────────────────────────────────────────────

  if (createdOrderId) {
    return (
      <div className="md-card-outlined p-6 text-center">
        <CheckCircle className="mx-auto mb-4 size-12 text-success" />
        <h2 className="font-serif text-xl font-bold text-ink-900">
          {t("bulk_exam_order_created")}
        </h2>
        <p className="mt-2 text-sm text-ink-500">{t("bulk_exam_order_created_desc")}</p>
        <p className="mt-1 text-sm text-ink-500">
          {examName} &middot;{" "}
          {t("bulk_exam_order_students_count").replace(
            "{n}",
            String(selectedStudentIds.length),
          )}
        </p>

        <div className="mt-8 flex flex-col items-center gap-4 sm:flex-row sm:justify-center">
          <SnapCheckout orderId={createdOrderId} basePath="/admin/bulk-exam-orders" />
          <Button variant="outline" onClick={handleReset}>
            {t("bulk_exam_order_reset")}
          </Button>
        </div>
      </div>
    );
  }

  if (grantResult) {
    return (
      <div className="md-card-outlined p-6 text-center">
        <CheckCircle className="mx-auto mb-4 size-12 text-success" />
        <h2 className="font-serif text-xl font-bold text-ink-900">
          {t("exam_grant_success_title")}
        </h2>
        <p className="mt-2 text-sm text-ink-500">
          {t("exam_grant_success_desc_count").replace(
            "{n}",
            String(grantResult.granted_count),
          )}{" "}
          &middot; {examName}
        </p>

        {grantResult.granted_students.length > 0 && (
          <div className="mx-auto mt-6 max-h-[200px] max-w-sm overflow-y-auto rounded-lg border border-line p-2 text-left">
            {grantResult.granted_students.map((s) => (
              <div key={s.id} className="flex items-center gap-2 px-2 py-1.5 text-sm">
                <span className="font-medium text-ink-900">{s.name}</span>
                <span className="text-ink-500">@{s.username}</span>
              </div>
            ))}
          </div>
        )}

        <div className="mt-8 flex justify-center">
          <Button variant="outline" onClick={handleReset}>
            {t("exam_grant_grant_again")}
          </Button>
        </div>
      </div>
    );
  }

  // ── Main picker + action ────────────────────────────────────────────────

  return (
    <div className="space-y-6">
      <section>
        <h3 className="font-serif text-base font-semibold text-ink-900">
          {t("bulk_exam_order_pick_participants")}
        </h3>
        <div className="mt-3">
          <ParticipantPicker
            schoolId={isSuperAdmin ? undefined : schoolId}
            selected={selectedStudentIds}
            onChange={setSelectedStudentIds}
          />
        </div>
      </section>

      {selectedStudentIds.length > 0 && isSuperAdmin && (
        <div className="space-y-4">
          <Button size="lg" onClick={handleGrant} disabled={grantMutation.isPending}>
            {grantMutation.isPending ? (
              <Loader2 className="mr-2 size-4 animate-spin" />
            ) : null}
            {grantMutation.isPending ? t("exam_grant_granting") : t("exam_grant_grant")}
          </Button>

          {grantMutation.isError && (
            <p className="text-sm text-danger">{t("error_generic")}</p>
          )}
        </div>
      )}

      {selectedStudentIds.length > 0 && !isSuperAdmin && (
        <div className="space-y-4">
          <Button size="lg" onClick={handlePreview} disabled={previewMutation.isPending}>
            {previewMutation.isPending ? (
              <Loader2 className="mr-2 size-4 animate-spin" />
            ) : null}
            {t("bulk_exam_order_preview")}
          </Button>

          {previewMutation.data && (
            <div className="md-card-outlined space-y-4 p-5">
              <h4 className="font-serif text-base font-semibold text-ink-900">
                {t("bulk_exam_order_preview_title")}
              </h4>
              <div className="flex items-center gap-2">
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
                      <span className="font-medium text-ink-900">{s.name}</span>
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

          {previewMutation.isError && (
            <p className="text-sm text-danger">{t("bulk_exam_order_preview_failed")}</p>
          )}
        </div>
      )}
    </div>
  );
}
