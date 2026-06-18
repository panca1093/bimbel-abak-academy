"use client";

import { CheckCircle2, PlayCircle } from "lucide-react";
import { Progress } from "@/components/ui/progress";
import type { CourseSection, Lesson } from "@/lib/types";

export interface LessonListProps {
  sections: (CourseSection & { lessons?: Lesson[] })[];
  activeLessonId?: string;
  onSelectLesson: (lessonId: string) => void;
  doneCount: number;
  totalCount: number;
  progressPct: number;
}

export function LessonList({
  sections,
  activeLessonId,
  onSelectLesson,
  doneCount,
  totalCount,
  progressPct,
}: LessonListProps) {
  return (
    <div className="rounded-lg border border-line bg-surface p-4 md:sticky md:top-4">
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-sm font-semibold text-ink-900">Daftar Pelajaran</h3>
        <span className="font-mono text-xs text-ink-500">
          {doneCount}/{totalCount}
        </span>
      </div>
      <Progress
        value={progressPct}
        className="bg-ink-900/10 [&>[data-slot=progress-indicator]]:bg-brand-600"
      />
      <div className="mt-4 flex max-h-[60vh] flex-col gap-4 overflow-auto pr-1">
        {sections.map((section) => (
          <div key={section.id}>
            <div className="mb-1.5 text-[11px] font-semibold uppercase tracking-wide text-ink-500">
              {section.title}
            </div>
            {section.lessons?.map((lesson) => {
              const isActive = lesson.id === activeLessonId;
              const isDone = Boolean(lesson.completed);
              return (
                <button
                  key={lesson.id}
                  type="button"
                  onClick={() => onSelectLesson(lesson.id)}
                  className={[
                    "mb-0.5 flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left transition-colors",
                    isActive
                      ? "bg-brand-50 text-brand-700"
                      : "text-ink-700 hover:bg-ink-50",
                  ].join(" ")}
                >
                  {isDone ? (
                    <CheckCircle2 className="size-4 shrink-0 text-brand-600" />
                  ) : (
                    <PlayCircle
                      className={[
                        "size-4 shrink-0",
                        isActive ? "text-brand-600" : "text-ink-400",
                      ].join(" ")}
                    />
                  )}
                  <span
                    className={[
                      "flex-1 text-[13px]",
                      isActive ? "font-semibold" : "font-normal",
                    ].join(" ")}
                  >
                    {lesson.title}
                  </span>
                  {typeof lesson.duration_seconds === "number" &&
                    lesson.duration_seconds > 0 && (
                      <span className="shrink-0 font-mono text-[11px] text-ink-400">
                        {Math.round(lesson.duration_seconds / 60)}m
                      </span>
                    )}
                </button>
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
}