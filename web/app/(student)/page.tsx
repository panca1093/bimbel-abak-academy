"use client";

import Link from "next/link";
import {
  AlertCircle,
  ChevronRight,
  Plus,
  Trophy,
  Clock,
  Target,
  Construction,
} from "lucide-react";
import { useDashboard } from "@/lib/hooks/students";
import { useAuthStore } from "@/stores/auth";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PendingBanner } from "@/components/dashboard/PendingBanner";

const COURSE_GRADIENTS = [
  "linear-gradient(135deg, #EEF0FE 0%, #F3E8FB 45%, #FDE8EE 100%)",
  "linear-gradient(135deg, #EAF0FE 0%, #E8F6FB 50%, #EEF0FE 100%)",
  "linear-gradient(135deg, #F0ECFC 0%, #EAF0FE 55%, #E8F7F2 100%)",
  "linear-gradient(135deg, #FDEEF0 0%, #EEF0FE 60%, #EAF6FB 100%)",
];

function greet(): string {
  const h = new Date().getHours();
  if (h < 11) return "Selamat pagi";
  if (h < 16) return "Selamat siang";
  return "Selamat malam";
}

function firstName(name?: string): string {
  if (!name) return "";
  return name.trim().split(/\s+/)[0];
}

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user);
  const { data, isLoading, isError, error, refetch } = useDashboard();
  const greeting = greet();
  const name = firstName(user?.name ?? user?.username ?? undefined);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10">
      <header className="mb-8 flex items-end justify-between gap-4">
        <div>
          <p className="text-sm font-medium text-ink-500">{greeting}</p>
          <h1 className="mt-1 font-serif text-3xl font-bold text-ink-900 md:text-4xl">
            {name ? `Halo, ${name}` : "Halo"}
          </h1>
        </div>
        <Button asChild variant="outline" size="sm" className="hidden md:inline-flex">
          <Link href="/catalog">
            <Plus className="size-4" /> Tambah Kursus
          </Link>
        </Button>
      </header>

      {isError && (
        <Card className="mb-8 border-danger/30 bg-danger-bg px-5 py-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="size-5 text-danger" />
            <div className="flex-1 text-sm text-ink-700">
              Gagal memuat dashboard.
              {error instanceof Error && error.message ? ` ${error.message}` : ""}
            </div>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              Coba lagi
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

          <section className="mb-8">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <DashboardStubCard
                title="Peringkat saya"
                icon={Trophy}
                placeholder="Total peringkat akan tersedia setelah ujian pertama."
              />
              <DashboardStubCard
                title="Total jam belajar"
                icon={Clock}
                placeholder="Ringkasan waktu belajar akan hadir di sini."
              >
                <div className="mx-auto my-2 size-28 rounded-full border-8 border-brand-100 border-t-brand-600" />
              </DashboardStubCard>
              <DashboardStubCard
                title="Progress ujian"
                icon={Target}
                placeholder="Statistik progress ujian sedang disiapkan."
              >
                <div className="my-2 space-y-3">
                  <div className="space-y-1">
                    <div className="flex justify-between text-xs text-ink-500">
                      <span>Tryout 1</span>
                      <span>—</span>
                    </div>
                    <div className="h-2 w-full rounded-full bg-ink-100" />
                  </div>
                  <div className="space-y-1">
                    <div className="flex justify-between text-xs text-ink-500">
                      <span>Tryout 2</span>
                      <span>—</span>
                    </div>
                    <div className="h-2 w-full rounded-full bg-ink-100" />
                  </div>
                  <div className="space-y-1">
                    <div className="flex justify-between text-xs text-ink-500">
                      <span>Tryout 3</span>
                      <span>—</span>
                    </div>
                    <div className="h-2 w-full rounded-full bg-ink-100" />
                  </div>
                </div>
              </DashboardStubCard>
            </div>
          </section>

          <section className="mb-8">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="font-serif text-xl font-semibold text-ink-900">
                Kursus saya
              </h2>
              <Link
                href="/courses"
                className="inline-flex items-center gap-1 text-sm font-medium text-brand-700 hover:text-brand-800"
              >
                Lihat semua <ChevronRight className="size-4" />
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

          <section>
            <Card className="flex flex-col items-center justify-center gap-3 border-dashed border-line bg-surface-2 px-6 py-10 text-center">
              <div className="flex size-12 items-center justify-center rounded-full bg-brand-50 text-brand-600">
                <Plus className="size-6" />
              </div>
              <div>
                <p className="font-semibold text-ink-900">Jelajahi katalog</p>
                <p className="mt-1 text-sm text-ink-500">
                  Temukan buku, kursus, dan paket kompetisi baru.
                </p>
              </div>
              <Button asChild size="sm" className="mt-2">
                <Link href="/catalog">Buka katalog</Link>
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

function DashboardStubCard({
  title,
  icon: Icon,
  placeholder,
  children,
}: {
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  placeholder: string;
  children?: React.ReactNode;
}) {
  return (
    <Card className="flex flex-col border-line px-5 py-5">
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex size-8 items-center justify-center rounded-lg bg-brand-50 text-brand-600">
            <Icon className="size-4" />
          </div>
          <h3 className="font-serif text-base font-semibold text-ink-900">{title}</h3>
        </div>
        <Badge variant="outline" className="text-[10px] font-medium">
          <Construction className="mr-1 size-3" />
          Akan datang
        </Badge>
      </div>
      <div className="flex flex-1 flex-col justify-center">
        {children}
      </div>
      <p className="mt-3 text-xs text-ink-500">{placeholder}</p>
    </Card>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-8">
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