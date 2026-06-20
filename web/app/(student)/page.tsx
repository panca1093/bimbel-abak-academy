"use client";

import Link from "next/link";
import {
  AlertCircle,
  ChevronRight,
  Plus,
  Construction,
} from "lucide-react";
import { useDashboard } from "@/lib/hooks/students";
import { useTranslation } from "@/lib/i18n";
import { useAuthStore } from "@/stores/auth";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PendingBanner } from "@/components/dashboard/PendingBanner";
import { RankingCard } from "@/components/dashboard/RankingCard";
import { StudySummaryCard } from "@/components/dashboard/StudySummaryCard";
import { ExamProgressCard } from "@/components/dashboard/ExamProgressCard";
import { PopularLessonsSection } from "@/components/dashboard/PopularLessonsSection";

const COURSE_GRADIENTS = [
  "linear-gradient(135deg, #EEF0FE 0%, #F3E8FB 45%, #FDE8EE 100%)",
  "linear-gradient(135deg, #EAF0FE 0%, #E8F6FB 50%, #EEF0FE 100%)",
  "linear-gradient(135deg, #F0ECFC 0%, #EAF0FE 55%, #E8F7F2 100%)",
  "linear-gradient(135deg, #FDEEF0 0%, #EEF0FE 60%, #EAF6FB 100%)",
];

function firstName(name?: string): string {
  if (!name) return "";
  return name.trim().split(/\s+/)[0];
}

export default function DashboardPage() {
  const { t } = useTranslation();
  const user = useAuthStore((s) => s.user);
  const { data, isLoading, isError, error, refetch } = useDashboard();
  const name = firstName(user?.name ?? user?.username ?? undefined);

  const greeting = (() => {
    const h = new Date().getHours();
    if (h < 11) return t("greeting_morning");
    if (h < 16) return t("greeting_afternoon");
    return t("greeting_evening");
  })();

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10">
      <header className="mb-8 flex items-end justify-between gap-4">
        <div>
          <p className="text-sm font-medium text-ink-500">{greeting}</p>
          <h1 className="mt-1 font-serif text-3xl font-bold text-ink-900 md:text-4xl">
            {name ? `${t("hello")}, ${name}` : t("hello")}
          </h1>
        </div>
        <Button asChild variant="outline" size="sm" className="hidden md:inline-flex">
          <Link href="/catalog">
            <Plus className="size-4" /> {t("add_course")}
          </Link>
        </Button>
      </header>

      {isError && (
        <Card className="mb-8 border-danger/30 bg-danger-bg px-5 py-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="size-5 text-danger" />
            <div className="flex-1 text-sm text-ink-700">
              {t("dash_load_failed")}
              {error instanceof Error && error.message ? ` ${error.message}` : ""}
            </div>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              {t("retry")}
            </Button>
          </div>
        </Card>
      )}

      {isLoading ? (
        <DashboardSkeleton />
      ) : data ? (
        <>
          {data.pending_order && (
            <PendingBanner
              id={data.pending_order.id}
              product={data.pending_order.product}
              amount={data.pending_order.amount}
            />
          )}

          {/* (b) Ranking, (e) Study Summary, (f) Exam Progress */}
          <section className="mb-8">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <RankingCard ranking={data.ranking} />
              <StudySummaryCard study={data.study_summary} />
              <ExamProgressCard examProgress={data.exam_progress} />
            </div>
          </section>

          {/* (d) My Courses */}
          <section className="mb-8">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="font-serif text-xl font-semibold text-ink-900">
                {t("nav_courses")}
              </h2>
              <Link
                href="/courses"
                className="inline-flex items-center gap-1 text-sm font-medium text-brand-700 hover:text-brand-800"
              >
                {t("dash_view_all")} <ChevronRight className="size-4" />
              </Link>
            </div>

            {data.enrolled_courses.length === 0 ? (
              <EmptyCourses />
            ) : (
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {data.enrolled_courses.map((course, i) => (
                  <CourseCard
                    key={course.id}
                    course={course}
                    gradient={COURSE_GRADIENTS[i % COURSE_GRADIENTS.length]}
                  />
                ))}
              </div>
            )}
          </section>

          {/* (g) Popular Lessons */}
          <PopularLessonsSection lessons={data.popular_lessons} />

          {/* Explore Catalog CTA */}
          <section>
            <Card className="flex flex-col items-center justify-center gap-3 border-dashed border-line bg-surface-2 px-6 py-10 text-center">
              <div className="flex size-12 items-center justify-center rounded-full bg-brand-50 text-brand-600">
                <Plus className="size-6" />
              </div>
              <div>
                <p className="font-semibold text-ink-900">{t("dash_explore_catalog")}</p>
                <p className="mt-1 text-sm text-ink-500">
                  {t("dash_catalog_desc")}
                </p>
              </div>
              <Button asChild size="sm" className="mt-2">
                <Link href="/catalog">{t("dash_open_catalog")}</Link>
              </Button>
            </Card>
          </section>
        </>
      ) : null}
    </div>
  );
}

function CourseCard({
  course,
  gradient,
}: {
  course: {
    id: string;
    title: string;
    progress: number;
    total_lessons: number;
    done_lessons: number;
    cover?: string;
  };
  gradient: string;
}) {
  const pct = Math.round((course.progress ?? 0) * 100);
  return (
    <Link
      href={`/courses/${course.id}`}
      className="group flex flex-col rounded-lg border border-line p-5 transition-all hover:-translate-y-0.5 hover:shadow-md"
      style={{ background: gradient }}
    >
      <h3 className="text-lg font-bold text-ink-900">{course.title}</h3>
      <div className="mt-auto flex items-center justify-between pt-4">
        <span className="font-mono text-sm font-bold text-brand-700">
          {String(course.done_lessons).padStart(2, "0")}
          <span className="text-ink-400">/{course.total_lessons}</span>
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

function EmptyCourses() {
  const { t } = useTranslation();
  return (
    <Card className="flex flex-col items-center justify-center gap-3 border-dashed border-line bg-surface-2 px-6 py-10 text-center">
      <div className="flex size-12 items-center justify-center rounded-full bg-brand-50 text-brand-600">
        <Plus className="size-6" />
      </div>
      <div>
        <p className="font-semibold text-ink-900">{t("dash_no_courses")}</p>
        <p className="mt-1 text-sm text-ink-500">
          {t("dash_no_courses_desc")}
        </p>
      </div>
      <Button asChild size="sm" className="mt-2">
        <Link href="/catalog">{t("dash_open_catalog")}</Link>
      </Button>
    </Card>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-8" data-testid="dashboard-skeleton">
      <div className="flex items-end justify-between gap-4">
        <div className="space-y-2">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-9 w-56" />
        </div>
        <Skeleton className="hidden h-8 w-36 md:block" />
      </div>
      <Skeleton className="h-20 w-full rounded-lg" />
      <div>
        <Skeleton className="mb-4 h-6 w-32" />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-40 rounded-lg" />
          ))}
        </div>
      </div>
    </div>
  );
}
