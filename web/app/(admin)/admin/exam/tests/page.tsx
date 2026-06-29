"use client";

import { useState } from "react";
import { toast } from "sonner";
import { ClipboardList } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { TestModal } from "@/components/admin/TestModal";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useAdminTests,
  useCreateTest,
  useUpdateTest,
  useDeleteTest,
} from "@/lib/hooks/admin-tests";
import { useTranslation } from "@/lib/i18n";
import type { Test, AdminCreateTestInput, AdminUpdateTestInput } from "@/lib/types";

export default function TestsPage() {
  const { t } = useTranslation();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingTest, setEditingTest] = useState<Test | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const { data: testsResp, isLoading, isError, error } = useAdminTests();
  const tests = testsResp?.data ?? [];
  const create = useCreateTest();
  const update = useUpdateTest(editingTest?.id ?? "");
  const remove = useDeleteTest(deletingId ?? "");

  function openCreate() {
    setEditingTest(null);
    setModalOpen(true);
  }

  function openEdit(test: Test) {
    setEditingTest(test);
    setModalOpen(true);
  }

  function errorMessage(err: unknown): string {
    if (err instanceof Error) return err.message;
    return t("error_generic");
  }

  async function handleSubmit(input: AdminCreateTestInput | AdminUpdateTestInput) {
    try {
      if (editingTest) {
        await update.mutateAsync(input as AdminUpdateTestInput);
        toast.success(t("tests_update_success"));
      } else {
        await create.mutateAsync(input as AdminCreateTestInput);
        toast.success(t("tests_create_success"));
      }
      setModalOpen(false);
      setEditingTest(null);
    } catch (e) {
      toast.error(errorMessage(e));
    }
  }

  async function handleDelete(id: string) {
    if (!confirm(t("tests_confirm_delete"))) return;
    setDeletingId(id);
    try {
      await remove.mutateAsync();
      toast.success(t("tests_delete_success"));
    } catch (e) {
      toast.error(errorMessage(e));
    } finally {
      setDeletingId(null);
    }
  }

  return (
    <div className="space-y-6 fade-in">
      <AdminPageHeader
        icon={ClipboardList}
        title={t("tests_page_title")}
        actions={<Button onClick={openCreate}>{t("tests_new")}</Button>}
      />

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      )}

      {isError && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
          {errorMessage(error)}
        </div>
      )}

      {!isLoading && !isError && (
        <div className="overflow-x-auto md-card-outlined">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-3 text-left font-medium">{t("tests_field_title")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("tests_field_subject")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("tests_field_topic")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("tests_field_duration")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("tests_question_count")}</th>
                <th className="px-4 py-3 text-right font-medium">{t("th_actions")}</th>
              </tr>
            </thead>
            <tbody>
              {tests.map((test) => (
                <tr
                  key={test.id}
                  className="border-t transition-colors hover:bg-muted/40"
                >
                  <td className="px-4 py-3 font-medium">{test.title}</td>
                  <td className="px-4 py-3">{test.subject}</td>
                  <td className="px-4 py-3">{test.topic}</td>
                  <td className="px-4 py-3">{test.duration_minutes}</td>
                  <td className="px-4 py-3">{test.question_count ?? 0}</td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <Button size="sm" variant="outline" onClick={() => openEdit(test)}>
                        {t("action_edit")}
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => handleDelete(test.id)}
                        disabled={remove.isPending && deletingId === test.id}
                      >
                        {t("action_delete")}
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
              {tests.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-muted-foreground">
                    {t("tests_empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <TestModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        test={editingTest}
        onSubmit={handleSubmit}
        isPending={create.isPending || update.isPending}
      />
    </div>
  );
}