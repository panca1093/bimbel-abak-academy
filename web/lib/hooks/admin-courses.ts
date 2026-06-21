"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type {
  Course,
  AdminCourseDetail,
  AdminCreateCourseInput,
  AdminUpdateCourseInput,
  CourseSection,
  AdminCreateSectionInput,
  AdminUpdateSectionInput,
  AdminReorderSectionsInput,
  Lesson,
  AdminCreateLessonInput,
  AdminUpdateLessonInput,
  AdminReorderLessonsInput,
} from "@/lib/types";

export const adminCoursesKeys = {
  all: ["admin", "courses"] as const,
  list: () => [...adminCoursesKeys.all, "list"] as const,
  detail: (id: string) => [...adminCoursesKeys.all, "detail", id] as const,
  sections: (courseId: string) => [...adminCoursesKeys.detail(courseId), "sections"] as const,
};

export function useAdminCourses() {
  return useQuery({
    queryKey: adminCoursesKeys.list(),
    queryFn: async () => {
      const res = await authFetch<{ data: Course[]; next_cursor?: string }>("/admin/courses");
      return res.data ?? [];
    },
  });
}

export function useAdminCourse(id: string) {
  return useQuery({
    queryKey: adminCoursesKeys.detail(id),
    queryFn: () => authFetch<AdminCourseDetail>(`/admin/courses/${encodeURIComponent(id)}`),
    enabled: Boolean(id),
  });
}

export function useCreateCourse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminCreateCourseInput) =>
      authFetch<Course>("/admin/courses", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminCoursesKeys.list() });
    },
  });
}

export function useUpdateCourse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: AdminUpdateCourseInput }) =>
      authFetch<Course>(`/admin/courses/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: JSON.stringify(input),
      }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: adminCoursesKeys.list() });
      qc.invalidateQueries({ queryKey: adminCoursesKeys.detail(id) });
    },
  });
}

export function useAdminSections(courseId: string) {
  return useQuery({
    queryKey: adminCoursesKeys.sections(courseId),
    queryFn: async () => {
      const res = await authFetch<{ data: CourseSection[] }>(
        `/admin/courses/${encodeURIComponent(courseId)}/sections`
      );
      return res.data ?? [];
    },
    enabled: Boolean(courseId),
  });
}

export function useCreateSection(courseId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminCreateSectionInput) =>
      authFetch<CourseSection>(`/admin/courses/${encodeURIComponent(courseId)}/sections`, {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminCoursesKeys.sections(courseId) });
    },
  });
}

export function useUpdateSection(courseId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ sectionId, input }: { sectionId: string; input: AdminUpdateSectionInput }) =>
      authFetch<CourseSection>(
        `/admin/courses/${encodeURIComponent(courseId)}/sections/${encodeURIComponent(sectionId)}`,
        {
          method: "PUT",
          body: JSON.stringify(input),
        }
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminCoursesKeys.sections(courseId) });
    },
  });
}

export function useDeleteSection(courseId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (sectionId: string) =>
      authFetch<void>(
        `/admin/courses/${encodeURIComponent(courseId)}/sections/${encodeURIComponent(sectionId)}`,
        {
          method: "DELETE",
        }
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminCoursesKeys.sections(courseId) });
    },
  });
}

export function useReorderSections(courseId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminReorderSectionsInput) =>
      authFetch<{ message: string }>(
        `/admin/courses/${encodeURIComponent(courseId)}/sections/reorder`,
        {
          method: "PATCH",
          body: JSON.stringify(input),
        }
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminCoursesKeys.sections(courseId) });
    },
  });
}

export function useCreateLesson(courseId: string, sectionId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminCreateLessonInput) =>
      authFetch<Lesson>(
        `/admin/courses/${encodeURIComponent(courseId)}/sections/${encodeURIComponent(
          sectionId
        )}/lessons`,
        {
          method: "POST",
          body: JSON.stringify(input),
        }
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminCoursesKeys.sections(courseId) });
    },
  });
}

export function useUpdateLesson(courseId: string, sectionId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ lessonId, input }: { lessonId: string; input: AdminUpdateLessonInput }) =>
      authFetch<Lesson>(
        `/admin/courses/${encodeURIComponent(courseId)}/sections/${encodeURIComponent(
          sectionId
        )}/lessons/${encodeURIComponent(lessonId)}`,
        {
          method: "PUT",
          body: JSON.stringify(input),
        }
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminCoursesKeys.sections(courseId) });
    },
  });
}

export function useDeleteLesson(courseId: string, sectionId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (lessonId: string) =>
      authFetch<void>(
        `/admin/courses/${encodeURIComponent(courseId)}/sections/${encodeURIComponent(
          sectionId
        )}/lessons/${encodeURIComponent(lessonId)}`,
        {
          method: "DELETE",
        }
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminCoursesKeys.sections(courseId) });
    },
  });
}

export function useReorderLessons(courseId: string, sectionId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AdminReorderLessonsInput) =>
      authFetch<{ message: string }>(
        `/admin/courses/${encodeURIComponent(courseId)}/sections/${encodeURIComponent(
          sectionId
        )}/lessons/reorder`,
        {
          method: "PATCH",
          body: JSON.stringify(input),
        }
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminCoursesKeys.sections(courseId) });
    },
  });
}
