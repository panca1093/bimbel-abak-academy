"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { Course, CourseSection, CourseSession, Lesson } from "@/lib/types";

export interface CourseDetail extends Course {
  sections?: (CourseSection & { lessons?: Lesson[] })[];
}

export const coursesKeys = {
  all: ["courses"] as const,
  list: () => [...coursesKeys.all, "list"] as const,
  detail: (id: string) => [...coursesKeys.all, "detail", id] as const,
};

export function useMyCourses() {
  return useQuery({
    queryKey: coursesKeys.list(),
    queryFn: async () => {
      const res = await authFetch<{ data: CourseSession[] }>(`/courses`);
      return res.data ?? [];
    },
  });
}

export function useCourse(id: string) {
  return useQuery({
    queryKey: coursesKeys.detail(id),
    queryFn: () => authFetch<CourseDetail>(`/courses/${encodeURIComponent(id)}`),
    enabled: Boolean(id),
  });
}

interface CompleteLessonInput {
  courseId: string;
  lessonId: string;
}

export function useCompleteLesson() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ courseId, lessonId }: CompleteLessonInput) =>
      authFetch<void>(
        `/courses/${encodeURIComponent(courseId)}/lessons/${encodeURIComponent(lessonId)}/complete`,
        { method: "POST" }
      ),
    onSuccess: (_data, { courseId }) => {
      qc.invalidateQueries({ queryKey: coursesKeys.detail(courseId) });
      qc.invalidateQueries({ queryKey: coursesKeys.list() });
    },
  });
}