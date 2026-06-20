"use client";

import { Trophy, Play, Layers, Clock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Ring } from "@/components/dashboard/Ring";
import { useTranslation } from "@/lib/i18n";
import type { PopularLessonEntry } from "@/lib/types";

export function PopularLessonsSection({
  lessons,
}: {
  lessons: PopularLessonEntry[];
}) {
  const { t } = useTranslation();

  return (
    <section className="mb-8">
      <div className="mb-4 flex items-center gap-2">
        <Trophy className="size-4 text-brand-600" />
        <h2 className="font-serif text-xl font-semibold text-ink-900">
          {t("dash_popular_lessons")}
        </h2>
      </div>

      {lessons.length === 0 ? (
        <Card className="flex flex-col items-center justify-center border-dashed border-line bg-surface-2 px-6 py-10 text-center">
          <p className="font-semibold text-ink-900">{t("dash_popular_empty")}</p>
          <p className="mt-1 text-sm text-ink-500">{t("dash_popular_empty_desc")}</p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {lessons.map((lesson, i) => (
            <Card
              key={i}
              className="flex items-center gap-4 border-line px-5 py-4"
            >
              <Ring value={lesson.progress} size={52} thickness={6} />
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-semibold text-ink-900">
                  {lesson.title}
                </p>
                <div className="mt-1 flex flex-wrap gap-3">
                  <span className="inline-flex items-center gap-1 text-xs text-ink-500">
                    <Layers className="size-3" />
                    {lesson.topics} {t("dash_topics")}
                  </span>
                  <span className="inline-flex items-center gap-1 text-xs text-ink-500">
                    <Clock className="size-3" />
                    {lesson.duration}
                  </span>
                </div>
              </div>
              <Button size="icon" className="size-9 shrink-0 rounded-full">
                <Play className="size-4" />
              </Button>
            </Card>
          ))}
        </div>
      )}
    </section>
  );
}
