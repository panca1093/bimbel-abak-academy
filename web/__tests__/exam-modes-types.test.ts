import { describe, it, expect } from "vitest";
import type {
  SectionType,
  SessionTest,
  SessionStartPayload,
  SessionState,
  Test,
  Exam,
  AdminCreateTestInput,
  AdminUpdateTestInput,
  SessionMonitorRow,
  ResultTopicRow,
} from "@/lib/types";
import { useAdvanceSection } from "@/lib/hooks/exam";

describe("exam-modes types", () => {
  it("SectionType is a string union", () => {
    const valid: SectionType[] = ["listening", "reading", "writing"];
    expect(valid).toHaveLength(3);
  });

  it("SessionTest accepts sectioned fields", () => {
    const st: SessionTest = {
      id: "t1",
      title: "Listening",
      subject: "English",
      questions: [],
      section_type: "listening",
      duration_minutes: 30,
      audio_url: "https://example.com/audio.mp3",
      audio_play_limit: 2,
      status: "active",
      remaining_seconds: 1800,
    };
    expect(st.section_type).toBe("listening");
  });

  it("SessionStartPayload accepts mode + active_test_id", () => {
    const payload: SessionStartPayload = {
      session_id: "s1",
      remaining_seconds: 1800,
      timer_mode: "per_test",
      duration_minutes: null,
      tests: [],
      mode: "utbk",
      active_test_id: "t1",
    };
    expect(payload.mode).toBe("utbk");
  });

  it("Test accepts section_type", () => {
    const t: Test = {
      id: "t1",
      title: "Listening",
      subject: "English",
      topic: "Section 1",
      duration_minutes: 30,
      section_type: "listening",
    };
    expect(t.section_type).toBe("listening");
  });

  it("Exam accepts mode", () => {
    const e: Exam = {
      id: "e1",
      title: "IELTS Mock",
      mode: "ielts",
    };
    expect(e.mode).toBe("ielts");
  });

  it("AdminCreateTestInput accepts section_type", () => {
    const input: AdminCreateTestInput = {
      title: "Listening",
      subject: "English",
      topic: "Section 1",
      duration_minutes: 30,
      section_type: "listening",
    };
    expect(input.section_type).toBe("listening");
  });

  it("AdminUpdateTestInput accepts section_type", () => {
    const input: AdminUpdateTestInput = {
      title: "Listening Updated",
      section_type: "reading",
    };
    expect(input.section_type).toBe("reading");
  });

  it("SessionMonitorRow accepts active-section fields", () => {
    const row: SessionMonitorRow = {
      registration_id: "r1",
      student_id: "s1",
      student_name: "Student One",
      school_name: null,
      status: "in_progress",
      answers_saved: 5,
      total_questions: 20,
      checked_in_at: null,
      last_saved_at: null,
      violation_count: 0,
      session_id: null,
      admin_submitted: false,
      extended_until: null,
      active_section_test_id: "t1",
      active_section_title: "Listening",
      active_section_started_at: "2026-07-08T00:00:00Z",
      active_section_duration_minutes: 30,
      active_section_extended_until: null,
      active_section_remaining_seconds: 1750,
    };
    expect(row.active_section_test_id).toBe("t1");
  });

  it("ResultTopicRow accepts section_type", () => {
    const row: ResultTopicRow = {
      test_id: "t1",
      title: "Listening",
      subject: "English",
      topic: "Section 1",
      earned: 15,
      max: 20,
      section_type: "listening",
    };
    expect(row.section_type).toBe("listening");
  });

  it("useAdvanceSection is exported and typed", () => {
    const hook = useAdvanceSection;
    expect(hook).toBeDefined();
    expect(typeof hook).toBe("function");
  });
});
