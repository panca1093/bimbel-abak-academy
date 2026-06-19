"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { toast } from "sonner";
import { useAdminCourse, useUpdateCourse } from "@/lib/hooks/admin-courses";
import { SectionEditor } from "@/components/admin/SectionEditor";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { ArrowLeft } from "lucide-react";
import type { AdminUpdateCourseInput } from "@/lib/types";

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  return "Terjadi kesalahan.";
}

export default function CourseBuilderPage() {
  const params = useParams();
  const router = useRouter();
  const courseId = typeof params.id === "string" ? params.id : "";

  const { data: course, isLoading, isError, error } = useAdminCourse(courseId);
  const update = useUpdateCourse();

  const [title, setTitle] = useState("");
  const [level, setLevel] = useState("");
  const [subject, setSubject] = useState("");
  const [instructorName, setInstructorName] = useState("");

  useEffect(() => {
    if (course) {
      setTitle(course.title ?? "");
      setLevel(course.level ?? "");
      setSubject(course.subject ?? "");
      setInstructorName(course.instructor_name ?? "");
    }
  }, [course]);

  const dirty = useMemo(() => {
    if (!course) return false;
    return (
      title !== (course.title ?? "") ||
      level !== (course.level ?? "") ||
      subject !== (course.subject ?? "") ||
      instructorName !== (course.instructor_name ?? "")
    );
  }, [course, title, level, subject, instructorName]);

  async function handleSaveMetadata(e: React.FormEvent) {
    e.preventDefault();
    if (!courseId || !dirty) return;

    const input: AdminUpdateCourseInput = {};
    if (title !== (course?.title ?? "")) input.title = title.trim();
    if (level !== (course?.level ?? "")) input.level = level.trim() || undefined;
    if (subject !== (course?.subject ?? "")) input.subject = subject.trim() || undefined;
    if (instructorName !== (course?.instructor_name ?? ""))
      input.instructor_name = instructorName.trim() || undefined;

    try {
      await update.mutateAsync({ id: courseId, input });
      toast.success("Metadata kursus disimpan.");
    } catch (err) {
      toast.error(errorMessage(err));
    }
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-48 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (isError || !course) {
    return (
      <div className="rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
        Gagal memuat kursus: {errorMessage(error)}
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" onClick={() => router.push("/admin/courses")}>
          <ArrowLeft className="mr-1 size-4" />
          Kembali
        </Button>
        <h1 className="text-2xl font-semibold">Edit kursus</h1>
      </div>

      <form onSubmit={handleSaveMetadata} className="rounded-lg border bg-card p-6">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-medium">Metadata kursus</h2>
          <Button type="submit" disabled={!dirty || !title.trim() || update.isPending}>
            {update.isPending ? "Menyimpan..." : "Simpan metadata"}
          </Button>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div className="grid gap-2">
            <Label htmlFor="course-title">Judul</Label>
            <Input
              id="course-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Judul kursus"
              disabled={update.isPending}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="course-level">Jenjang</Label>
            <Input
              id="course-level"
              value={level}
              onChange={(e) => setLevel(e.target.value)}
              placeholder="SMA / SMP"
              disabled={update.isPending}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="course-subject">Mapel</Label>
            <Input
              id="course-subject"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              placeholder="Matematika"
              disabled={update.isPending}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="course-instructor">Pengajar</Label>
            <Input
              id="course-instructor"
              value={instructorName}
              onChange={(e) => setInstructorName(e.target.value)}
              placeholder="Nama pengajar"
              disabled={update.isPending}
            />
          </div>
        </div>
      </form>

      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-4 text-lg font-medium">Kurikulum</h2>
        <SectionEditor courseId={courseId} />
      </div>
    </div>
  );
}
