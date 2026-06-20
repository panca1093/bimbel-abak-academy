"use client";

import { useMemo, useState } from "react";
import {
  Plus,
  Upload,
  Tag,
  Edit,
  MoreHorizontal,
} from "lucide-react";
import { useTranslation } from "@/lib/i18n";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";

type QuestionFormat = "mcq" | "multi_answer" | "short" | "fill_blank" | "essay";
type Difficulty = "easy" | "medium" | "hard";

interface Question {
  id: string;
  test: string;
  topic: string;
  format: QuestionFormat;
  difficulty: Difficulty;
  body: string;
  pointCorrect?: number;
  pointWrong?: number;
}

interface Topic {
  id: string;
  name: string;
  subject: string;
  count: number;
}

const INITIAL_TOPICS: Topic[] = [
  { id: "TP-01", name: "Aljabar", subject: "Matematika", count: 8 },
  { id: "TP-02", name: "Geometri", subject: "Matematika", count: 6 },
  { id: "TP-03", name: "Statistika", subject: "Matematika", count: 5 },
  { id: "TP-04", name: "Kosakata", subject: "B. Indonesia", count: 4 },
  { id: "TP-05", name: "Pemahaman Bacaan", subject: "B. Indonesia", count: 7 },
  { id: "TP-06", name: "Listening", subject: "English", count: 9 },
  { id: "TP-07", name: "Writing", subject: "English", count: 3 },
];

const INITIAL_QUESTIONS: Question[] = [
  {
    id: "AQ-1041",
    test: "Penalaran Matematika TO#12",
    topic: "Aljabar",
    format: "mcq",
    difficulty: "medium",
    body: "Jika 2x+3=11 maka x²−1 = …",
  },
  {
    id: "AQ-1042",
    test: "Penalaran Matematika TO#12",
    topic: "Geometri",
    format: "mcq",
    difficulty: "easy",
    body: "Keliling persegi 36 cm, luasnya …",
  },
  {
    id: "AQ-1043",
    test: "Penalaran Matematika TO#12",
    topic: "Statistika",
    format: "multi_answer",
    difficulty: "hard",
    body: "Pernyataan benar tentang median …",
  },
  {
    id: "AQ-1101",
    test: "Literasi B. Indonesia TO#12",
    topic: "Kosakata",
    format: "fill_blank",
    difficulty: "easy",
    body: "Kereta berangkat _____ pukul tujuh.",
  },
  {
    id: "AQ-1102",
    test: "Literasi B. Indonesia TO#12",
    topic: "Pemahaman Bacaan",
    format: "short",
    difficulty: "medium",
    body: "Sinonim dari \"cerdas\" …",
  },
  {
    id: "AQ-1203",
    test: "Listening Section A — IELTS",
    topic: "Listening",
    format: "mcq",
    difficulty: "medium",
    body: "What time does the lecture begin?",
  },
  {
    id: "AQ-1204",
    test: "Writing Task — IELTS",
    topic: "Writing",
    format: "essay",
    difficulty: "hard",
    body: "Describe advantages of timed practice…",
  },
];

const FORMATS: QuestionFormat[] = [
  "mcq",
  "multi_answer",
  "short",
  "fill_blank",
  "essay",
];

const DIFFICULTY_TONE: Record<Difficulty, string> = {
  easy: "bg-success-bg text-success border-success",
  medium: "bg-warn-bg text-warn border-warn",
  hard: "bg-danger-bg text-danger border-danger",
};

const FORMAT_TONE: Record<QuestionFormat, string> = {
  mcq: "bg-info-bg text-info border-info",
  multi_answer: "bg-violet-bg text-violet border-violet",
  short: "bg-ink-100 text-ink-600 border-line",
  fill_blank: "bg-ink-100 text-ink-600 border-line",
  essay: "bg-warn-bg text-warn border-warn",
};

const DEFAULT_POINTS: Record<Difficulty, number> = {
  easy: 1,
  medium: 2,
  hard: 3,
};

type DictKey = keyof (typeof import("@/lib/i18n").DICT)["id"];

function fmtFormat(format: QuestionFormat, t: (key: DictKey) => string) {
  return t(`fmt_${format}` as DictKey);
}

function fmtDifficulty(difficulty: Difficulty, t: (key: DictKey) => string) {
  return t(`diff_${difficulty}` as DictKey);
}

function questionPoints(q: Question) {
  return {
    correct: q.pointCorrect ?? DEFAULT_POINTS[q.difficulty],
    wrong: q.pointWrong ?? 0,
  };
}

function plainText(html: string) {
  if (typeof window === "undefined") return html;
  const d = document.createElement("div");
  d.innerHTML = html;
  return (d.textContent || "").replace(/\s+/g, " ").trim();
}

export default function ExamBanksPage() {
  const { t } = useTranslation();
  const [format, setFormat] = useState<QuestionFormat | "all">("all");
  const [topic, setTopic] = useState<string>("all");
  const [topics, setTopics] = useState(INITIAL_TOPICS);
  const [questions] = useState(INITIAL_QUESTIONS);
  const [topicsOpen, setTopicsOpen] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);

  const subjects = useMemo(
    () => [...new Set(topics.map((tp) => tp.subject))],
    [topics]
  );

  const rows = useMemo(() => {
    return questions.filter(
      (q) =>
        (format === "all" || q.format === format) &&
        (topic === "all" || q.topic === topic)
    );
  }, [questions, format, topic]);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 md:px-6 md:py-10 fade-in"
    >
      <header className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
      >
        <div>
          <h1 className="font-serif text-3xl font-bold text-ink-900 md:text-4xl"
          >
            {t("question_bank")}
          </h1>
          <p className="mt-2 text-sm text-ink-500"
          >
            Kelola soal — terikat ke satu Tes, dikelompokkan per topik.
          </p>
        </div>
        <div className="flex flex-wrap gap-2"
        >
          <Button variant="outline" size="sm" onClick={() => setTopicsOpen(true)}
          >
            <Tag className="mr-1 size-4" />
            {t("manage_topics")}
          </Button>
          <Button variant="outline" size="sm"
          >
            <Upload className="mr-1 size-4" />
            CSV
          </Button>
          <Button size="sm" onClick={() => setCreateOpen(true)}
          >
            <Plus className="mr-1 size-4" />
            {t("create")}
          </Button>
        </div>
      </header>

      <div className="mb-4 flex flex-wrap items-center gap-2"
      >
        <FilterChip
          active={format === "all"}
          onClick={() => setFormat("all")}
        >
          {t("tab_all")}
        </FilterChip>
        {FORMATS.map((f) => (
          <FilterChip
            key={f}
            active={format === f}
            onClick={() => setFormat(f)}
          >
            {fmtFormat(f, t)}
          </FilterChip>
        ))}
        <div className="ml-auto flex items-center gap-2"
        >
          <Tag className="size-4 text-ink-400" />
          <Select value={topic} onValueChange={setTopic}
          >
            <SelectTrigger className="h-9 w-[180px] text-xs"
            >
              <SelectValue placeholder={t("all_topics")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t("all_topics")}</SelectItem>
              {topics.map((tp) => (
                <SelectItem key={tp.id} value={tp.name}
                >
                  {tp.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <Card className="overflow-hidden"
      >
        <div className="overflow-x-auto"
        >
          <table className="w-full text-sm"
          >
            <thead className="bg-surface-2 text-left text-xs font-semibold text-ink-600"
            >
              <tr>
                <th className="px-4 py-3">ID</th>
                <th className="px-4 py-3">{t("question")}</th>
                <th className="px-4 py-3">{t("test")}</th>
                <th className="px-4 py-3">{t("topic")}</th>
                <th className="px-4 py-3">{t("format")}</th>
                <th className="px-4 py-3">{t("difficulty")}</th>
                <th className="px-4 py-3">{t("points")}</th>
                <th className="px-4 py-3 text-right"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line"
            >
              {rows.map((q) => {
                const pts = questionPoints(q);
                return (
                  <tr key={q.id} className="group hover:bg-surface-2"
                  >
                    <td className="px-4 py-3 font-mono text-xs text-ink-500"
                    >
                      {q.id}
                    </td>
                    <td className="px-4 py-3">
                      <div className="max-w-[320px] truncate font-medium text-ink-900"
                      >
                        {plainText(q.body)}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-xs text-ink-600"
                    >
                      {q.test}
                    </td>
                    <td className="px-4 py-3"
                    >
                      <span className="inline-flex items-center rounded-md bg-surface-2 px-2 py-1 text-xs font-medium text-ink-600"
                      >
                        {q.topic}
                      </span>
                    </td>
                    <td className="px-4 py-3"
                    >
                      <Badge
                        variant="outline"
                        className={cn(
                          "text-[11px] font-semibold",
                          FORMAT_TONE[q.format]
                        )}
                      >
                        {fmtFormat(q.format, t)}
                      </Badge>
                    </td>
                    <td className="px-4 py-3"
                    >
                      <Badge
                        variant="outline"
                        className={cn(
                          "text-[11px] font-semibold",
                          DIFFICULTY_TONE[q.difficulty]
                        )}
                      >
                        {fmtDifficulty(q.difficulty, t)}
                      </Badge>
                    </td>
                    <td className="px-4 py-3"
                    >
                      <span className="font-mono text-xs font-bold text-brand-700"
                      >
                        +{pts.correct}
                      </span>{" "}
                      <span className="font-mono text-xs text-ink-400"
                      >
                        / {pts.wrong}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right"
                    >
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon-xs"
                          >
                            <MoreHorizontal className="size-4 text-ink-500" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end"
                        >
                          <DropdownMenuItem onClick={() => setCreateOpen(true)}
                          >
                            <Edit className="mr-2 size-4" />
                            {t("update")}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </Card>

      <TopicsDialog
        open={topicsOpen}
        onClose={() => setTopicsOpen(false)}
        topics={topics}
        subjects={subjects}
        onChange={setTopics}
      />

      <Dialog open={createOpen} onOpenChange={setCreateOpen}
      >
        <DialogContent className="sm:max-w-lg"
        >
          <DialogHeader>
            <DialogTitle className="font-serif"
            >
              {t("create_question")}
            </DialogTitle>
            <DialogDescription>
              Form pembuat soal akan tersedia di iterasi berikutnya.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4"
          >
            <div>
              <Label>{t("question")}</Label>
              <Input placeholder="Tulis soal…" />
            </div>
            <div>
              <Label>{t("topic")}</Label>
              <Select>
                <SelectTrigger>
                  <SelectValue placeholder={t("select_school")} />
                </SelectTrigger>
                <SelectContent>
                  {topics.map((tp) => (
                    <SelectItem key={tp.id} value={tp.name}
                    >
                      {tp.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex justify-end gap-2"
            >
              <Button variant="outline" onClick={() => setCreateOpen(false)}
              >
                {t("cancel")}
              </Button>
              <Button onClick={() => setCreateOpen(false)}
              >
                {t("create")}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function FilterChip({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "rounded-lg border px-3 py-[7px] text-xs font-semibold transition-colors",
        active
          ? "border-brand-600 bg-brand-600 text-white"
          : "border-line bg-surface text-ink-600 hover:text-ink-900"
      )}
    >
      {children}
    </button>
  );
}

function TopicsDialog({
  open,
  onClose,
  topics,
  subjects,
  onChange,
}: {
  open: boolean;
  onClose: () => void;
  topics: Topic[];
  subjects: string[];
  onChange: (topics: Topic[]) => void;
}) {
  const { t } = useTranslation();
  const [name, setName] = useState("");
  const [subject, setSubject] = useState(subjects[0] ?? "Matematika");

  function add() {
    if (!name.trim()) return;
    onChange([
      ...topics,
      {
        id: `TP-${Math.random().toString(36).slice(2, 5).toUpperCase()}`,
        name: name.trim(),
        subject,
        count: 0,
      },
    ]);
    setName("");
  }

  function remove(id: string) {
    onChange(topics.filter((tp) => tp.id !== id));
  }

  return (
    <Dialog open={open} onOpenChange={onClose}
    >
      <DialogContent className="sm:max-w-md"
      >
        <DialogHeader>
          <DialogTitle className="font-serif"
          >
            {t("manage_topics")}
          </DialogTitle>
          <DialogDescription>Kelola topik soal ujian.</DialogDescription>
        </DialogHeader>
        <div className="flex items-end gap-2"
        >
          <div className="flex-1"
          >
            <Label className="text-xs"
            >
              {t("topic_name")}
            </Label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="mis. Trigonometri"
            />
          </div>
          <div className="w-[140px]"
          >
            <Label className="text-xs"
            >
              {t("subject")}
            </Label>
            <Select value={subject} onValueChange={setSubject}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {subjects.map((s) => (
                  <SelectItem key={s} value={s}
                  >
                    {s}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <Button onClick={add} className="shrink-0"
          >
            <Plus className="mr-1 size-4" />
            {t("add_topic")}
          </Button>
        </div>
        <div className="max-h-[46vh] overflow-y-auto rounded-lg border border-line"
        >
          {topics.map((tp, i) => (
            <div
              key={tp.id}
              className={cn(
                "flex items-center gap-3 px-4 py-3",
                i < topics.length - 1 && "border-b border-line"
              )}
            >
              <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-brand-50 text-brand-600"
              >
                <Tag className="size-4" />
              </div>
              <div className="min-w-0 flex-1"
              >
                <div className="text-sm font-semibold text-ink-900"
                >
                  {tp.name}
                </div>
                <div className="text-xs text-ink-500"
                >
                  {tp.subject} · {tp.count} {t("questions_in_topic")}
                </div>
              </div>
              <Button
                variant="ghost"
                size="icon-xs"
                onClick={() => remove(tp.id)}
                className="text-danger"
              >
                <span className="sr-only">Remove</span>×
              </Button>
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
