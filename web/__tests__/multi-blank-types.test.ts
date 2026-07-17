import { describe, it, expect } from "vitest";
import type {
  QuestionFormat,
  Question,
  SessionQuestion,
  QuestionWithOptions,
  BankQuestionListItem,
  AdminQuestionInput,
} from "@/lib/types";

describe("multi-blank types", () => {
  it("QuestionFormat includes multi_blank", () => {
    const formats: QuestionFormat[] = ["mcq", "multi_answer", "short", "fill_blank", "essay", "multi_blank"];
    expect(formats).toContain("multi_blank");
  });

  it("Question accepts audio_url", () => {
    const q: Question = {
      id: "q1",
      format: "multi_blank",
      body: "{{1}} and {{2}}",
      sort_order: 1,
      point_correct: 2,
      point_wrong: 1,
      audio_url: "https://example.com/audio.mp3",
    };
    expect(q.audio_url).toBe("https://example.com/audio.mp3");
  });

  it("Question audio_url can be null", () => {
    const q: Question = {
      id: "q1",
      format: "mcq",
      body: "What is...?",
      sort_order: 1,
      point_correct: 1,
      point_wrong: 0,
      audio_url: null,
    };
    expect(q.audio_url).toBeNull();
  });

  it("SessionQuestion accepts audio_url and blanks", () => {
    const sq: SessionQuestion = {
      id: "q1",
      test_id: "t1",
      format: "multi_blank",
      body: "{{1}} and {{2}}",
      sort_order: 1,
      options: [],
      audio_url: "https://example.com/audio.mp3",
      blanks: [1, 2],
    };
    expect(sq.audio_url).toBe("https://example.com/audio.mp3");
    expect(sq.blanks).toEqual([1, 2]);
  });

  it("QuestionWithOptions accepts blanks field", () => {
    const qwo: QuestionWithOptions = {
      question: {
        id: "q1",
        format: "multi_blank",
        body: "{{1}} and {{2}}",
        sort_order: 1,
        point_correct: 2,
        point_wrong: 1,
      },
      options: [],
      blanks: [
        { index: 1, correct_answer: "jakarta" },
        { index: 2, correct_answer: "1945" },
      ],
    };
    expect(qwo.blanks).toHaveLength(2);
    expect(qwo.blanks?.[0].index).toBe(1);
  });

  it("BankQuestionListItem accepts blanks field", () => {
    const bqli: BankQuestionListItem = {
      question: {
        id: "q1",
        format: "multi_blank",
        body: "{{1}} and {{2}}",
        sort_order: 1,
        point_correct: 2,
        point_wrong: 1,
      },
      options: [],
      attached_count: 3,
      blanks: [
        { index: 1, correct_answer: "jakarta" },
        { index: 2, correct_answer: "1945" },
      ],
    };
    expect(bqli.blanks).toHaveLength(2);
    expect(bqli.blanks?.[1].correct_answer).toBe("1945");
  });

  it("AdminQuestionInput accepts audio_url and blanks", () => {
    const input: AdminQuestionInput = {
      format: "multi_blank",
      body: "{{1}} and {{2}}",
      point_correct: 2,
      point_wrong: 1,
      audio_url: "https://example.com/audio.mp3",
      blanks: [
        { index: 1, correct_answer: "jakarta" },
        { index: 2, correct_answer: "1945" },
      ],
    };
    expect(input.audio_url).toBe("https://example.com/audio.mp3");
    expect(input.blanks).toHaveLength(2);
  });
});
