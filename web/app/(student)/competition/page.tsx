"use client";

import { useMemo, useState } from "react";
import {
  ClipboardList,
  Clock,
  Calendar,
  Download,
  Play,
  Trophy,
  Key,
  AlertCircle,
  XCircle,
  CheckCircle,
  ChevronRight,
  Award,
} from "lucide-react";
import { toast } from "sonner";
import { useTranslation } from "@/lib/i18n";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

type PackageState =
  | "free"
  | "locked"
  | "checkin"
  | "checkedin"
  | "inprogress"
  | "submitted"
  | "expired";

interface CompetitionPackage {
  id: string;
  title: string;
  type: "competition" | "exam";
  free: boolean;
  state: PackageState;
  attempts?: number;
  best?: number | null;
  duration: number;
  scheduled?: string;
  checkinOpens?: string;
  countdown?: string;
  token?: string;
  resultConfig?: "hidden" | "score_only" | "score_pembahasan";
  score?: number;
  rank?: number;
}

const INITIAL_PACKAGES: CompetitionPackage[] = [
  {
    id: "p1",
    title: "Try Out UTBK Gratis #12",
    type: "competition",
    free: true,
    state: "free",
    attempts: 2,
    best: 742,
    duration: 90,
  },
  {
    id: "p2",
    title: "Simulasi SNBT Nasional 2026",
    type: "competition",
    free: false,
    state: "checkin",
    token: "SMA1-7X9K2",
    scheduled: "29 Mei 2026, 09:00",
    duration: 120,
    checkinOpens: "08:30",
  },
  {
    id: "p3",
    title: "OSN Matematika — Babak Penyisihan",
    type: "competition",
    free: false,
    state: "locked",
    token: "SMA1-3M5P8",
    scheduled: "05 Jun 2026, 13:00",
    duration: 120,
    checkinOpens: "12:30",
    countdown: "6 hari 7 jam",
  },
  {
    id: "p4",
    title: "IELTS Full Mock Test — Academic",
    type: "exam",
    free: false,
    state: "submitted",
    token: "SMA1-9K2L4",
    scheduled: "22 Mei 2026, 10:00",
    duration: 165,
    resultConfig: "score_pembahasan",
    score: 7.5,
    rank: 18,
  },
  {
    id: "p5",
    title: "Try Out TKA SMA Semester Genap",
    type: "competition",
    free: false,
    state: "expired",
    token: "SMA1-2B8N6",
    scheduled: "18 Mei 2026, 09:00",
    duration: 90,
  },
  {
    id: "p6",
    title: "Latihan Soal HOTS Saintek",
    type: "competition",
    free: true,
    state: "free",
    attempts: 0,
    best: null,
    duration: 60,
  },
];

const STATE_META: Record<
  PackageState,
  { tone: "brand" | "warn" | "danger" | "success" | "info" | "secondary"; key?: string }
> = {
  free: { tone: "success" },
  locked: { tone: "secondary", key: "st_locked" },
  checkin: { tone: "info", key: "st_checkin" },
  checkedin: { tone: "warn", key: "st_checkedin" },
  inprogress: { tone: "success", key: "st_inprogress" },
  submitted: { tone: "brand", key: "st_submitted" },
  expired: { tone: "danger", key: "st_expired" },
};

function formatDuration(minutes: number, lang: "id" | "en") {
  return `${minutes} ${lang === "id" ? "menit" : "min"}`;
}

function replaceAttempts(text: string, n: number) {
  return text.replace("{n}", String(n));
}

export default function CompetitionPage() {
  const { t, lang } = useTranslation();
  const [packages, setPackages] = useState(INITIAL_PACKAGES);
  const [activePkg, setActivePkg] = useState<CompetitionPackage | null>(null);
  const [dialog, setDialog] = useState<
    "exam" | "result" | "leaderboard" | null
  >(null);

  const free = useMemo(() => packages.filter((p) => p.free), [packages]);
  const mine = useMemo(() => packages.filter((p) => !p.free), [packages]);

  function handleCheckIn(id: string, tokenValue: string) {
    const pkg = packages.find((p) => p.id === id);
    if (!pkg) return;
    if (tokenValue.trim().toUpperCase() !== (pkg.token ?? "")) {
      toast.error(t("invalid_token"));
      return;
    }
    setPackages((prev) =>
      prev.map((p) => (p.id === id ? { ...p, state: "checkedin" } : p))
    );
    toast.success(t("checkin_success"));
  }

  function handleStart(pkg: CompetitionPackage) {
    setActivePkg(pkg);
    setDialog("exam");
  }

  function handleResult(pkg: CompetitionPackage) {
    setActivePkg(pkg);
    setDialog("result");
  }

  function handleLeaderboard(pkg: CompetitionPackage) {
    setActivePkg(pkg);
    setDialog("leaderboard");
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <header className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl">
          {t("comp_title")}
        </h1>
        <p className="mt-2 text-sm text-ink-500">{t("comp_sub")}</p>
      </header>

      <Section
        title={t("free_packages")}
        packages={free}
        onStart={handleStart}
      />

      <Section
        title={t("my_packages")}
        packages={mine}
        onStart={handleStart}
        onCheckIn={handleCheckIn}
        onResult={handleResult}
        onLeaderboard={handleLeaderboard}
      />

      <Dialog open={dialog === "exam"} onOpenChange={() => setDialog(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="font-serif">
              {activePkg?.title}
            </DialogTitle>
            <DialogDescription>{t("exam_coming_soon")}</DialogDescription>
          </DialogHeader>
          <div className="flex items-center gap-3 rounded-lg bg-brand-50 px-4 py-3 text-sm text-brand-700">
            <Play className="size-4" />
            <span>
              {activePkg?.scheduled
                ? activePkg.scheduled
                : formatDuration(activePkg?.duration ?? 0, lang)}
            </span>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={dialog === "result"} onOpenChange={() => setDialog(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="font-serif">
              {t("view_result")}
            </DialogTitle>
            <DialogDescription>{t("result_coming_soon")}</DialogDescription>
          </DialogHeader>
          {activePkg?.score != null && (
            <div className="rounded-lg bg-brand-600 px-4 py-6 text-center text-white">
              <div className="text-xs font-semibold uppercase tracking-wider opacity-80">
                {t("score")}
              </div>
              <div className="font-serif text-5xl font-bold">
                {activePkg.score}
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>

      <Dialog
        open={dialog === "leaderboard"}
        onOpenChange={() => setDialog(null)}
      >
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="font-serif">
              {t("view_leaderboard")}
            </DialogTitle>
            <DialogDescription>{t("result_coming_soon")}</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            {[
              { rank: 1, name: "Kevin Wijaya", school: "SMAN 8 Jakarta", score: 892 },
              { rank: 2, name: "Nadia Salsabila", school: "SMA Pelita", score: 871 },
              { rank: 3, name: "Reza Pratama", school: "SMAN 3 Bandung", score: 858 },
              {
                rank: activePkg?.rank ?? 18,
                name: "Anda",
                school: "SMAN 1 Jakarta",
                score: (activePkg?.score ?? 0) * 100 || 758,
                me: true,
              },
            ].map((row) => (
              <div
                key={row.rank}
                className={cn(
                  "flex items-center gap-3 rounded-lg border border-line px-3 py-2",
                  row.me && "bg-brand-50"
                )}
              >
                <div className="w-8 text-center">
                  {row.rank <= 3 ? (
                    <Award
                      className={cn(
                        "mx-auto size-5",
                        row.rank === 1 && "text-gold",
                        row.rank === 2 && "text-ink-400",
                        row.rank === 3 && "text-[#B07B3E]"
                      )}
                    />
                  ) : (
                    <span className="font-serif text-sm font-semibold text-ink-400">
                      {row.rank}
                    </span>
                  )}
                </div>
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-semibold text-ink-900">
                    {row.name}
                  </div>
                  <div className="truncate text-xs text-ink-500">
                    {row.school}
                  </div>
                </div>
                <div className="font-serif text-base font-bold text-brand-700">
                  {row.score}
                </div>
              </div>
            ))}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function Section({
  title,
  packages,
  onStart,
  onCheckIn,
  onResult,
  onLeaderboard,
}: {
  title: string;
  packages: CompetitionPackage[];
  onStart: (pkg: CompetitionPackage) => void;
  onCheckIn?: (id: string, token: string) => void;
  onResult?: (pkg: CompetitionPackage) => void;
  onLeaderboard?: (pkg: CompetitionPackage) => void;
}) {
  if (packages.length === 0) return null;
  return (
    <section className="mb-10">
      <h2 className="mb-4 text-xs font-bold uppercase tracking-wider text-ink-500">
        {title}
      </h2>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        {packages.map((pkg) => (
          <PackageCard
            key={pkg.id}
            pkg={pkg}
            onStart={() => onStart(pkg)}
            onCheckIn={onCheckIn}
            onResult={() => onResult?.(pkg)}
            onLeaderboard={() => onLeaderboard?.(pkg)}
          />
        ))}
      </div>
    </section>
  );
}

function PackageCard({
  pkg,
  onStart,
  onCheckIn,
  onResult,
  onLeaderboard,
}: {
  pkg: CompetitionPackage;
  onStart: () => void;
  onCheckIn?: (id: string, token: string) => void;
  onResult?: () => void;
  onLeaderboard?: () => void;
}) {
  const { t, lang } = useTranslation();
  const [token, setToken] = useState("");
  const meta = STATE_META[pkg.state];
  const Icon = pkg.type === "exam" ? ClipboardList : Trophy;

  return (
    <Card className="flex flex-col p-5">
      <div className="flex items-start justify-between gap-3">
        <div
          className={cn(
            "flex size-10 shrink-0 items-center justify-center rounded-[10px]",
            meta.tone === "secondary" && "bg-info-bg text-info",
            meta.tone === "info" && "bg-info-bg text-info",
            meta.tone === "warn" && "bg-warn-bg text-warn",
            meta.tone === "danger" && "bg-danger-bg text-danger",
            meta.tone === "success" && "bg-success-bg text-success",
            meta.tone === "brand" && "bg-brand-50 text-brand-600"
          )}
        >
          <Icon className="size-5" />
        </div>
        {meta.key ? (
          <Badge
            variant="outline"
            className={cn(
              "text-[11px] font-semibold",
              meta.tone === "info" && "border-info text-info",
              meta.tone === "warn" && "border-warn text-warn",
              meta.tone === "danger" && "border-danger text-danger",
              meta.tone === "success" && "border-success text-success",
              meta.tone === "brand" && "border-brand-600 text-brand-700",
              meta.tone === "secondary" && "border-line text-ink-500"
            )}
          >
            {t(meta.key as keyof (typeof import("@/lib/i18n").DICT)["id"])}
          </Badge>
        ) : (
          <Badge
            variant="outline"
            className="border-success text-success text-[11px] font-semibold"
          >
            {t("free")}
          </Badge>
        )}
      </div>

      <h3 className="mt-3 text-base font-semibold leading-snug text-ink-900">
        {pkg.title}
      </h3>
      <div className="mt-2 flex flex-wrap gap-3 text-xs text-ink-500">
        <span className="inline-flex items-center gap-1">
          <Clock className="size-3" />
          {formatDuration(pkg.duration, lang)}
        </span>
        {pkg.scheduled && (
          <span className="inline-flex items-center gap-1">
            <Calendar className="size-3" />
            {pkg.scheduled}
          </span>
        )}
      </div>

      <div className="mt-4 border-t border-line-2 pt-4">
        {pkg.state === "free" && (
          <div className="flex items-center justify-between gap-3">
            <div className="text-xs text-ink-500">
              {pkg.attempts != null && (
                <span>{replaceAttempts(t("attempts"), pkg.attempts)}</span>
              )}
              {pkg.best != null && (
                <span className="ml-2">
                  · {t("best")} {pkg.best}
                </span>
              )}
            </div>
            <Button size="sm" onClick={onStart}>
              <Play className="mr-1 size-4" />
              {t("start_exam")}
            </Button>
          </div>
        )}

        {pkg.state === "locked" && (
          <div className="flex items-center justify-between gap-3">
            <div className="text-xs text-ink-500">
              <div>
                {t("checkin_opens")} {pkg.checkinOpens}
              </div>
              {pkg.countdown && (
                <div className="mt-1 font-semibold text-info">
                  {t("countdown")}: {pkg.countdown}
                </div>
              )}
            </div>
            <Button size="sm" variant="outline" onClick={() => toast(t("exam_coming_soon"))}>
              <Download className="mr-1 size-4" />
              {t("download_card")}
            </Button>
          </div>
        )}

        {pkg.state === "checkin" && onCheckIn && (
          <div>
            <Label className="text-xs text-ink-600">{t("enter_token")}</Label>
            <div className="mt-1 flex gap-2">
              <Input
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder={t("token_placeholder")}
                className="h-9 text-sm uppercase"
              />
              <Button
                size="sm"
                onClick={() => onCheckIn(pkg.id, token)}
                className="shrink-0"
              >
                {t("check_in")}
              </Button>
            </div>
            {pkg.token && (
              <p className="mt-2 flex items-center gap-1 text-[11px] text-ink-400">
                <Key className="size-3" />
                {t("token_hint")} ({pkg.token})
              </p>
            )}
          </div>
        )}

        {pkg.state === "checkedin" && (
          <div className="flex items-center justify-between gap-3">
            <div className="text-xs font-semibold text-warn">
              {t("exam_starts").replace("{t}", "4:12")}
            </div>
            <Button size="sm" onClick={onStart}>
              <Play className="mr-1 size-4" />
              {t("start_exam")}
            </Button>
          </div>
        )}

        {pkg.state === "submitted" && (
          <div>
            {pkg.resultConfig === "hidden" ? (
              <p className="flex items-center gap-1 text-xs text-ink-500">
                <AlertCircle className="size-4 text-warn" />
                {t("result_hidden")}
              </p>
            ) : (
              <div className="flex items-center justify-between gap-3">
                <div>
                  <div className="text-xs text-ink-500">{t("score")}</div>
                  <div className="font-serif text-2xl font-bold text-brand-700">
                    {pkg.score ?? "—"}
                  </div>
                </div>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={onLeaderboard}
                  >
                    <Trophy className="mr-1 size-4" />
                    {t("view_leaderboard")}
                  </Button>
                  <Button size="sm" onClick={onResult}>
                    {t("view_result")}
                    <ChevronRight className="ml-1 size-4" />
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}

        {pkg.state === "expired" && (
          <p className="flex items-center gap-1 text-xs text-ink-500">
            <XCircle className="size-4 text-danger" />
            {lang === "id"
              ? "Anda tidak melakukan check-in. Jendela telah ditutup."
              : "You never checked in. The window has closed."}
          </p>
        )}

        {pkg.state === "inprogress" && (
          <div className="flex items-center justify-between gap-3">
            <p className="flex items-center gap-1 text-xs text-success">
              <CheckCircle className="size-4" />
              {t("st_inprogress")}
            </p>
            <Button size="sm" onClick={onStart}>
              <Play className="mr-1 size-4" />
              {t("start_exam")}
            </Button>
          </div>
        )}
      </div>
    </Card>
  );
}
