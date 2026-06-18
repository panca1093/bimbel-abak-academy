"use client";

import Link from "next/link";
import { AlertCircle, Plus } from "lucide-react";
import { useMyCourses } from "@/lib/hooks/courses";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { CourseCard } from "@/components/courses/CourseCard";

const COURSE_GRADIENTS = [
  "linear-gradient(135deg, #EEF0FE 0%, #F3E8FB 45%, #FDE8EE 100%)",
  "linear-gradient(135deg, #EAF0FE 0%, #E8F6FB 50%, #EEF0FE 100%)",
  "linear-gradient(135deg, #F0ECFC 0%, #EAF0FE 55%, #E8F7F2 100%)",
  "linear-gradient(135deg, #FDEEF0 0%, #EEF0FE 60%, #EAF6FB 100%)",
];

export default function CoursesPage() {
  const { data: sessions, isLoading, isError, error, refetch } = useMyCourses();

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10">
      <header className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
          Kursus saya
        </h1>
        <p className="mt-2 text-sm text-ink-500">
          Lanjutkan belajar dari kursus yang sudah terdaftar.
        </p>
      </header>

      {isError && (
        <Card className="mb-8 border-danger/30 bg-danger-bg px-5 py-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="size-5 text-danger" />
            <div className="flex-1 text-sm text-ink-700">
              Gagal memuat kursus.
              {error instanceof Error && error.message ? ` ${error.message}` : ""}
            </div>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              Coba lagi
            </Button>
          </div>
        </Card>
      )}

      {isLoading ? (
        <CoursesSkeleton />
      ) : sessions && sessions.length > 0 ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {sessions.map((session, i) => (
            <CourseCard
              key={session.id}
              session={session}
              gradient={COURSE_GRADIENTS[i % COURSE_GRADIENTS.length]}
            />
          ))}
        </div>
      ) : (
        <EmptyCourses />
      )}
    </div>
  );
}

function EmptyCourses() {
  return (
    <Card className="flex flex-col items-center justify-center gap-3 border-dashed border-line bg-surface-2 px-6 py-10 text-center">
      <div className="flex size-12 items-center justify-center rounded-full bg-brand-50 text-brand-600">
        <Plus className="size-6" />
      </div>
      <div>
        <p className="font-semibold text-ink-900">Belum ada kursus</p>
        <p className="mt-1 text-sm text-ink-500">
          Mulai belajar dengan menjelajahi katalog kami.
        </p>
      </div>
      <Button asChild size="sm" className="mt-2">
        <Link href="/catalog">Buka katalog</Link>
      </Button>
    </Card>
  );
}

function CoursesSkeleton() {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <Skeleton className="h-9 w-48" />
        <Skeleton className="h-4 w-72" />
      </div>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-40 rounded-lg" />
        ))}
      </div>
    </div>
  );
}