"use client";

import { useState, useMemo } from "react";
import { GraduationCap, Loader2, CheckCircle } from "lucide-react";
import { toast } from "sonner";
import { useTranslation } from "@/lib/i18n";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { ParticipantPicker } from "@/components/admin/ParticipantPicker";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useOrderableExams } from "@/lib/hooks/admin-bulk-exam-orders";
import { useGrantExamAccess } from "@/lib/hooks/admin-exam-grants";
import type { ExamListItem } from "@/lib/types";

export default function ExamGrantPage() {
  const { t } = useTranslation();

  // Step state
  const [selectedExamId, setSelectedExamId] = useState<string>("");
  const [selectedStudentIds, setSelectedStudentIds] = useState<string[]>([]);
  const [grantResult, setGrantResult] = useState<{
    granted_count: number;
    granted_students: Array<{ id: string; name: string; username: string }>;
  } | null>(null);

  // Hooks
  const { data: examsData, isLoading: examsLoading } = useOrderableExams();
  const grantMutation = useGrantExamAccess();

  const exams: ExamListItem[] = useMemo(
    () => examsData?.data ?? [],
    [examsData],
  );

  const selectedExam = useMemo(
    () => exams.find((e) => e.id === selectedExamId),
    [exams, selectedExamId],
  );

  const handleGrant = () => {
    if (!selectedExamId) {
      toast.error(t("exam_grant_empty_exam"));
      return;
    }
    if (selectedStudentIds.length === 0) {
      toast.error(t("exam_grant_empty_students"));
      return;
    }
    grantMutation.mutate(
      { exam_id: selectedExamId, student_ids: selectedStudentIds },
      {
        onSuccess: (result) => {
          setGrantResult(result);
          toast.success(t("exam_grant_success"));
        },
        onError: (err) => {
          const msg =
            err instanceof Error
              ? err.message
              : t("error_generic");
          toast.error(msg);
        },
      },
    );
  };

  const handleReset = () => {
    setSelectedExamId("");
    setSelectedStudentIds([]);
    setGrantResult(null);
    grantMutation.reset();
  };

  // ── Success state ──────────────────────────────────────────────────────

  if (grantResult) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-8 md:px-6 md:py-10 fade-in">
        <AdminPageHeader
          icon={CheckCircle}
          title={t("exam_grant_success_title")}
          description={`${t("exam_grant_success_desc")} ${selectedExam?.title}`}
        />

        <div className="md-card-outlined p-6 text-center">
          <CheckCircle className="mx-auto mb-4 size-12 text-success" />
          <h2 className="font-serif text-xl font-bold text-ink-900">
            {t("exam_grant_success_title")}
          </h2>
          <p className="mt-2 text-sm text-ink-500">
            {t("exam_grant_success_desc_count").replace("{n}", String(grantResult.granted_count))} &middot; {selectedExam?.title}
          </p>

          {grantResult.granted_students.length > 0 && (
            <div className="mx-auto mt-6 max-h-[200px] max-w-sm overflow-y-auto rounded-lg border border-line p-2 text-left">
              {grantResult.granted_students.map((s) => (
                <div
                  key={s.id}
                  className="flex items-center gap-2 px-2 py-1.5 text-sm"
                >
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
      </div>
    );
  }

  // ── Main form ──────────────────────────────────────────────────────────

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={GraduationCap}
        title={t("exam_grant_title")}
        description={t("exam_grant_subtitle")}
      />

      {/* Step 1: Pick exam */}
      <section className="mb-8">
        <h3 className="font-serif text-base font-semibold text-ink-900">
          {t("exam_grant_select_exam")}
        </h3>
        <div className="mt-2">
          {examsLoading ? (
            <div className="h-9 w-[280px] animate-pulse rounded-md bg-surface-2" />
          ) : (
            <Select value={selectedExamId} onValueChange={setSelectedExamId}>
              <SelectTrigger className="h-9 w-[280px] text-xs">
                <SelectValue
                  placeholder={t("exam_grant_select_exam")}
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

      {/* Step 2: Pick participants (no schoolId — cross-school picker) */}
      {selectedExamId && (
        <section className="mb-8">
          <h3 className="font-serif text-base font-semibold text-ink-900">
            {t("bulk_exam_order_pick_participants")}
          </h3>
          <div className="mt-3">
            <ParticipantPicker
              schoolId={undefined}
              selected={selectedStudentIds}
              onChange={setSelectedStudentIds}
            />
          </div>
        </section>
      )}

      {/* Grant action */}
      {selectedStudentIds.length > 0 && (
        <div className="space-y-4">
          <Button
            size="lg"
            onClick={handleGrant}
            disabled={grantMutation.isPending}
          >
            {grantMutation.isPending ? (
              <Loader2 className="mr-2 size-4 animate-spin" />
            ) : null}
            {grantMutation.isPending
              ? t("exam_grant_granting")
              : t("exam_grant_grant")}
          </Button>

          {grantMutation.isError && (
            <p className="text-sm text-danger">
              {t("error_generic")}
            </p>
          )}
        </div>
      )}

      {!selectedExamId && (
        <div className="py-12 text-center text-sm text-ink-500">
          {t("exam_grant_no_exam")}
        </div>
      )}
    </div>
  );
}
