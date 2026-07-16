"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { useAdminCourses, useCreateCourse } from "@/lib/hooks/admin-courses";
import { useTranslation } from "@/lib/i18n";
import { Library } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import type { Course, AdminCreateCourseInput } from "@/lib/types";
import type { DICT } from "@/lib/i18n";

type TFunc = (key: keyof (typeof DICT)["id"]) => string;

function CreateCourseModal({
  open,
  onOpenChange,
  onSubmit,
  isPending,
  t,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (input: AdminCreateCourseInput) => void;
  isPending: boolean;
  t: TFunc;
}) {
  const [title, setTitle] = useState("");
  const [level, setLevel] = useState("");
  const [subject, setSubject] = useState("");
  const [instructorName, setInstructorName] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim() || isPending) return;
    onSubmit({
      title: title.trim(),
      level: level.trim() || undefined,
      subject: subject.trim() || undefined,
      instructor_name: instructorName.trim() || undefined,
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{t("courses_create")}</DialogTitle>
            <DialogDescription>{t("courses_create_desc")}</DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="course-title">{t("th_title")}</Label>
              <Input
                id="course-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder={t("course_title_placeholder")}
                disabled={isPending}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="course-level">{t("th_level")}</Label>
                <Input
                  id="course-level"
                  value={level}
                  onChange={(e) => setLevel(e.target.value)}
                  placeholder={t("course_level_placeholder")}
                  disabled={isPending}
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="course-subject">{t("subject")}</Label>
                <Input
                  id="course-subject"
                  value={subject}
                  onChange={(e) => setSubject(e.target.value)}
                  placeholder={t("course_subject_placeholder")}
                  disabled={isPending}
                />
              </div>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="course-instructor">{t("th_instructor")}</Label>
              <Input
                id="course-instructor"
                value={instructorName}
                onChange={(e) => setInstructorName(e.target.value)}
                placeholder={t("course_instructor_placeholder")}
                disabled={isPending}
              />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isPending}>
              {t("cancel")}
            </Button>
            <Button type="submit" disabled={!title.trim() || isPending}>
              {isPending ? t("saving") : t("save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default function CoursesPage() {
  const { t } = useTranslation();
  const router = useRouter();
  const [modalOpen, setModalOpen] = useState(false);
  const { data: courses, isLoading, isError, error } = useAdminCourses();
  const create = useCreateCourse();

  function errorMessage(error: unknown): string {
    if (error instanceof Error) return error.message;
    return t("error_generic");
  }

  async function handleCreate(input: AdminCreateCourseInput) {
    try {
      await create.mutateAsync(input);
      toast.success(t("courses_created"));
      setModalOpen(false);
    } catch (e) {
      toast.error(errorMessage(e));
    }
  }

  function openCourse(course: Course) {
    router.push(`/admin/courses/${course.id}`);
  }

  return (
    <div className="space-y-6 fade-in">
      <AdminPageHeader
        icon={Library}
        title={t("courses_page_title")}
        description={t("courses_page_description")}
        actions={<Button onClick={() => setModalOpen(true)}>{t("courses_create")}</Button>}
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
          {t("courses_load_failed")}: {errorMessage(error)}
        </div>
      )}

      {!isLoading && !isError && (
        <div className="overflow-x-auto md-card-outlined">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-3 text-left font-medium">{t("th_title")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("th_level")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("subject")}</th>
                <th className="px-4 py-3 text-left font-medium">{t("th_instructor")}</th>
              </tr>
            </thead>
            <tbody>
              {(courses ?? []).map((course) => (
                <tr
                  key={course.id}
                  className="border-t cursor-pointer transition-colors hover:bg-muted/40"
                  onClick={() => openCourse(course)}
                >
                  <td className="px-4 py-3 font-medium">{course.title}</td>
                  <td className="px-4 py-3">{course.level ?? "-"}</td>
                  <td className="px-4 py-3">{course.subject ?? "-"}</td>
                  <td className="px-4 py-3">{course.instructor_name ?? "-"}</td>
                </tr>
              ))}
              {(courses ?? []).length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-muted-foreground">
                    {t("empty_courses")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <CreateCourseModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        onSubmit={handleCreate}
        isPending={create.isPending}
        t={t}
      />
    </div>
  );
}
