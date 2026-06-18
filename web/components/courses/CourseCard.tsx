"use client";

import Link from "next/link";
import { ChevronRight } from "lucide-react";
import type { CourseSession } from "@/lib/types";
import { useCourse } from "@/lib/hooks/courses";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";

export interface CourseCardProps {
  session: CourseSession;
  gradient: string;
}

function countLessons(
  sections?: { lessons?: { id: string }[] }[]
): number {
  if (!sections) return 0;
  return sections.reduce(
    (acc, s) => acc + (s.lessons?.length ?? 0),
    0
  );
}

export function CourseCard({ session, gradient }: CourseCardProps) {
  const { data: course, isLoading } = useCourse(session.course_id);
  const totalLessons = countLessons(course?.sections);
  const doneLessons = Object.keys(session.completed_lessons ?? {}).length;
  const pct = totalLessons > 0 ? Math.round((doneLessons / totalLessons) * 100) : 0;

  return (
    <Link
      href={`/courses/${session.course_id}`}
      className="group flex min-h-40 flex-col rounded-lg border border-line p-5 transition-all hover:-translate-y-0.5 hover:shadow-md"
      style={{ background: gradient }}
    >
      {isLoading || !course ? (
        <Skeleton className="h-5 w-3/4 bg-white/60" />
      ) : (
        <h3 className="text-lg font-bold text-ink-900">{course.title}</h3>
      )}
      <div className="mt-auto flex items-center justify-between pt-4">
        <span className="font-mono text-sm font-bold text-brand-700">
          {String(doneLessons).padStart(2, "0")}
          <span className="text-ink-400">/{totalLessons}</span>
        </span>
        <span className="flex size-7 items-center justify-center rounded-full bg-white shadow-sm text-brand-600 transition-transform group-hover:translate-x-0.5">
          <ChevronRight className="size-4" />
        </span>
      </div>
      <Progress
        value={pct}
        className="mt-3 bg-ink-900/10 [&>[data-slot=progress-indicator]]:bg-brand-600"
      />
    </Link>
  );
}