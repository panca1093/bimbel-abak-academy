"use client";

import Link from "next/link";
import { useParams } from "next/navigation";

import { useRegistration } from "@/lib/hooks/exam";
import { useProfile, useSchools } from "@/lib/hooks/students";
import { fileUrl } from "@/lib/api";
import { ExamCardPrintable } from "@/components/exam/ExamCardPrintable";
import styles from "@/components/exam/ExamCardPrintable.module.css";

const DASH = "—";

function fmtDate(iso: string | null | undefined): string {
  if (!iso) return DASH;
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return DASH;
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(d);
}

function fmtTime(iso: string | null | undefined): string | null {
  if (!iso) return null;
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return null;
  return new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(d);
}

export default function ExamCardPrintPage() {
  const params = useParams<{ id: string }>();
  const id = params?.id ?? "";

  const { data: reg, isLoading: regLoading, isError: regError } =
    useRegistration(id);
  const { data: student, isLoading: profileLoading } = useProfile();
  const { data: schools } = useSchools();

  if (regLoading || profileLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-ink-500">
        Memuat kartu…
      </div>
    );
  }

  if (regError || !reg) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-3 text-ink-500">
        <p>Kartu peserta tidak ditemukan.</p>
        <Link href="/exam" className="text-sm font-semibold text-primary">
          Kembali ke daftar ujian
        </Link>
      </div>
    );
  }

  const exam = reg.exam;

  const schoolName =
    student?.unlisted_school_name?.trim() ||
    schools?.find((s) => s.id === student?.school_id)?.name ||
    DASH;

  const participantNumber = reg.participant_no || DASH;

  const start = fmtTime(exam.scheduled_at);
  const end = fmtTime(exam.scheduled_end_at);
  const timeRange = start ? `${start}${end ? ` – ${end}` : ""} WIB` : DASH;

  return (
    <div className={styles.screen}>
      <div className={styles.toolbar}>
        <Link className={styles.backLink} href={`/exam/${id}`}>
          ← Kembali
        </Link>
        <p>
          <strong>Kartu Peserta Ujian</strong> — ukuran cetak A6 (148×105 mm)
        </p>
        <button
          type="button"
          className={styles.printBtn}
          onClick={() => window.print()}
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth={2}
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M6 9V2h12v7" />
            <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2" />
            <rect x="6" y="14" width="12" height="8" rx="1" />
          </svg>
          Cetak / Simpan PDF
        </button>
      </div>

      <ExamCardPrintable
        fullName={student?.name?.trim() || DASH}
        participantNumber={participantNumber}
        school={schoolName}
        grade={student?.grade?.trim() || DASH}
        dob={fmtDate(student?.dob)}
        photoUrl={fileUrl(student?.photo_url)}
        examName={exam.title || DASH}
        subject={reg.subject || DASH}
        date={fmtDate(exam.scheduled_at)}
        timeRange={timeRange}
        duration={
          exam.duration_minutes ? `${exam.duration_minutes} menit` : DASH
        }
        mode="Online (CBT)"
        platform={reg.platform || DASH}
        checkInCode={reg.token || DASH}
      />
    </div>
  );
}
