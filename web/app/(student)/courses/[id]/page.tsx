"use client";

import { useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { AlertCircle, ArrowLeft, Check, Loader2 } from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";

import { useCourse, useCompleteLesson } from "@/lib/hooks/courses";
import type { Lesson } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { LessonList } from "@/components/courses/LessonList";
import { VideoPlayer } from "@/components/courses/VideoPlayer";

type SectionWithLessons = NonNullable<
  ReturnType<typeof useCourse>["data"]
>["sections"] extends infer S
  ? S extends Array<infer U>
    ? U
    : never
  : never;

function flattenLessons(
  sections: SectionWithLessons[] | undefined
): Lesson[] {
  if (!sections) return [];
  return sections.flatMap((s) => s.lessons ?? []);
}

export default function CourseDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const courseId = params?.id ?? "";

  const { data: course, isLoading, isError, error, refetch } = useCourse(courseId);
  const completeLesson = useCompleteLesson();

  const flatLessons = useMemo(
    () => flattenLessons(course?.sections),
    [course?.sections]
  );

  const firstIncompleteId = useMemo(
    () => flatLessons.find((l) => !l.completed)?.id ?? flatLessons[0]?.id,
    [flatLessons]
  );

  const [activeLessonId, setActiveLessonId] = useState<string | undefined>(
    undefined
  );

  const activeId = activeLessonId ?? firstIncompleteId;
  const activeLesson = flatLessons.find((l) => l.id === activeId);

  const totalCount = flatLessons.length;
  const doneCount = flatLessons.filter((l) => l.completed).length;
  const progressPct =
    totalCount > 0 ? Math.round((doneCount / totalCount) * 100) : 0;

  const handleToggleComplete = () => {
    if (!activeLesson) return;
    const wasDone = Boolean(activeLesson.completed);
    completeLesson.mutate(
      { courseId, lessonId: activeLesson.id },
      {
        onSuccess: () => {
          toast.success(
            wasDone ? "Pelajaran dibatalkan." : "Pelajaran selesai."
          );
        },
        onError: (err: unknown) => {
          const message =
            err instanceof Error ? err.message : "Gagal memperbarui status.";
          toast.error(message);
        },
      }
    );
  };

  if (isLoading) {
    return <CourseDetailSkeleton />;
  }

  if (isError) {
    return (
      <div className="mx-auto max-w-6xl px-4 py-8 md:px-6">
        <Card className="border-danger/30 bg-danger-bg px-5 py-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="size-5 text-danger" />
            <div className="flex-1 text-sm text-ink-700">
              Gagal memuat kursus.
              {error instanceof Error && error.message
                ? ` ${error.message}`
                : ""}
            </div>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              Coba lagi
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  if (!course) {
    return (
      <div className="mx-auto max-w-6xl px-4 py-8 md:px-6">
        <p className="text-sm text-ink-500">Kursus tidak ditemukan.</p>
        <Button
          variant="ghost"
          size="sm"
          className="mt-3"
          onClick={() => router.push("/courses")}
        >
          Kembali ke kursus
        </Button>
      </div>
    );
  }

  const sections = course.sections ?? [];
  const activeSection = sections.find((s) =>
    s.lessons?.some((l) => l.id === activeId)
  );

  return (
    <div className="mx-auto max-w-6xl px-4 py-6 md:px-6 md:py-8">
      <Link
        href="/courses"
        className="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-ink-600 hover:text-ink-900"
      >
        <ArrowLeft className="size-4" />
        Kembali ke kursus
      </Link>

      <header className="mb-6">
        <h1 className="font-serif text-2xl font-bold text-ink-900 md:text-3xl">
          {course.title}
        </h1>
        {course.instructor_name && (
          <p className="mt-1 text-sm text-ink-500">
            Pengajar: {course.instructor_name}
          </p>
        )}
      </header>

      <div className="grid grid-cols-1 gap-5 lg:grid-cols-[1fr_320px]">
        <div className="min-w-0">
          <VideoPlayer
            videoRef={activeLesson?.video_url}
            title={activeLesson?.title}
          />

          <Card className="mt-4 p-5">
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0">
                {activeSection && (
                  <div className="text-[11px] font-semibold uppercase tracking-wide text-brand-700">
                    {activeSection.title}
                  </div>
                )}
                <h2 className="mt-1 font-serif text-xl font-semibold text-ink-900">
                  {activeLesson?.title ?? "Pilih pelajaran"}
                </h2>
                {typeof activeLesson?.duration_seconds === "number" &&
                  activeLesson.duration_seconds > 0 && (
                    <p className="mt-1 text-xs text-ink-500">
                      Durasi {Math.round(activeLesson.duration_seconds / 60)} menit
                    </p>
                  )}
              </div>
              <Button
                type="button"
                size="sm"
                variant={
                  activeLesson?.completed ? "secondary" : "default"
                }
                disabled={!activeLesson || completeLesson.isPending}
                onClick={handleToggleComplete}
                className="shrink-0"
              >
                {completeLesson.isPending ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : activeLesson?.completed ? (
                  <Check className="size-4" />
                ) : null}
                {activeLesson?.completed ? "Selesai" : "Tandai selesai"}
              </Button>
            </div>
          </Card>
        </div>

        <LessonList
          sections={sections}
          activeLessonId={activeId}
          onSelectLesson={setActiveLessonId}
          doneCount={doneCount}
          totalCount={totalCount}
          progressPct={progressPct}
        />
      </div>
    </div>
  );
}

function CourseDetailSkeleton() {
  return (
    <div className="mx-auto max-w-6xl px-4 py-6 md:px-6 md:py-8">
      <Skeleton className="mb-4 h-4 w-32" />
      <Skeleton className="mb-6 h-8 w-2/3" />
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-[1fr_320px]">
        <div>
          <Skeleton className="h-0 w-full pb-[56.25%]" />
          <Skeleton className="mt-4 h-24 w-full" />
        </div>
        <Skeleton className="h-80 w-full" />
      </div>
    </div>
  );
}