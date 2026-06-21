"use client";

import { Trophy, Award } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { useTranslation } from "@/lib/i18n";
import type { DashboardRanking } from "@/lib/types";

function medalColor(rank: number): string | null {
  if (rank === 1) return "var(--color-gold)";
  if (rank === 2) return "#9AA3B5";
  if (rank === 3) return "#C08A4A";
  return null;
}

function initial(name: string): string {
  return name.trim().charAt(0).toUpperCase() || "?";
}

export function RankingCard({ ranking }: { ranking: DashboardRanking }) {
  const { t } = useTranslation();
  const { leaderboard } = ranking;

  return (
    <Card className="flex h-full flex-col border-line px-5 py-5">
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex size-8 items-center justify-center rounded-lg bg-brand-50 text-brand-600">
            <Trophy className="size-4" />
          </div>
          <h3 className="font-serif text-base font-semibold text-ink-900">{t("dash_my_ranking")}</h3>
        </div>
      </div>

      {leaderboard.length === 0 ? (
        <div className="flex flex-1 items-center justify-center py-8 text-center">
          <p className="text-sm text-ink-500">{t("dash_ranking_empty")}</p>
        </div>
      ) : (
        <>
          <div className="flex flex-col">
            {leaderboard.map((entry, i) => {
              const medalClr = medalColor(entry.rank);
              return (
                <div
                  key={i}
                  className={`flex items-center gap-3 py-2 ${i < leaderboard.length - 1 ? "border-b border-line" : ""}`}
                >
                  <div className="flex w-6 shrink-0 items-center justify-center">
                    {medalClr ? (
                      <Award className="size-4" style={{ color: medalClr }} />
                    ) : (
                      <span className="font-mono text-xs font-bold text-ink-400">{entry.rank}</span>
                    )}
                  </div>
                  <Avatar className="size-7">
                    <AvatarFallback>{initial(entry.name)}</AvatarFallback>
                  </Avatar>
                  <div className="min-w-0 flex-1">
                    <span className="truncate text-sm font-semibold text-ink-900">
                      {entry.name}
                      {entry.is_me && (
                        <span className="text-brand-700"> &middot; {t("dash_ranking_you")}</span>
                      )}
                    </span>
                  </div>
                  <span className="font-mono text-sm font-bold text-brand-700">{entry.points}</span>
                </div>
              );
            })}
          </div>
          <Button asChild variant="ghost" size="sm" className="mt-2 self-center">
            <Link href="/competition">{t("dash_view_all")}</Link>
          </Button>
        </>
      )}
    </Card>
  );
}
