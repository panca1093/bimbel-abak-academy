"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { useAdminCourses, useCreateCourse } from "@/lib/hooks/admin-courses";
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

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  return "Terjadi kesalahan.";
}

function CreateCourseModal({
  open,
  onOpenChange,
  onSubmit,
  isPending,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (input: AdminCreateCourseInput) => void;
  isPending: boolean;
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
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create course</DialogTitle>
            <DialogDescription>Add a new course to the catalog.</DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="course-title">Title</Label>
              <Input
                id="course-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Course title"
                disabled={isPending}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="course-level">Level</Label>
                <Input
                  id="course-level"
                  value={level}
                  onChange={(e) => setLevel(e.target.value)}
                  placeholder="SMA / SMP"
                  disabled={isPending}
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="course-subject">Subject</Label>
                <Input
                  id="course-subject"
                  value={subject}
                  onChange={(e) => setSubject(e.target.value)}
                  placeholder="Matematika"
                  disabled={isPending}
                />
              </div>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="course-instructor">Instructor</Label>
              <Input
                id="course-instructor"
                value={instructorName}
                onChange={(e) => setInstructorName(e.target.value)}
                placeholder="Instructor name"
                disabled={isPending}
              />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={!title.trim() || isPending}>
              {isPending ? "Saving..." : "Save"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default function CoursesPage() {
  const router = useRouter();
  const [modalOpen, setModalOpen] = useState(false);
  const { data: courses, isLoading, isError, error } = useAdminCourses();
  const create = useCreateCourse();

  async function handleCreate(input: AdminCreateCourseInput) {
    try {
      await create.mutateAsync(input);
      toast.success("Kursus dibuat.");
      setModalOpen(false);
    } catch (e) {
      toast.error(errorMessage(e));
    }
  }

  function openCourse(course: Course) {
    router.push(`/admin/courses/${course.id}`);
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Course Builder</h1>
        <Button onClick={() => setModalOpen(true)}>Create course</Button>
      </div>

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      )}

      {isError && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
          Gagal memuat kursus: {errorMessage(error)}
        </div>
      )}

      {!isLoading && !isError && (
        <div className="overflow-x-auto rounded-lg border">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-3 text-left font-medium">Title</th>
                <th className="px-4 py-3 text-left font-medium">Level</th>
                <th className="px-4 py-3 text-left font-medium">Subject</th>
                <th className="px-4 py-3 text-left font-medium">Instructor</th>
              </tr>
            </thead>
            <tbody>
              {(courses ?? []).map((course) => (
                <tr
                  key={course.id}
                  className="border-t cursor-pointer hover:bg-muted/50"
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
                    Tidak ada kursus.
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
      />
    </div>
  );
}
