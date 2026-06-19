import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useAdminCourses,
  useAdminCourse,
  useCreateCourse,
  useUpdateCourse,
  useAdminSections,
  useCreateSection,
  useUpdateSection,
  useDeleteSection,
  useReorderSections,
  useCreateLesson,
  useUpdateLesson,
  useDeleteLesson,
  useReorderLessons,
  adminCoursesKeys,
} from "./admin-courses";
import type { Course, AdminCourseDetail, CourseSection, Lesson } from "@/lib/types";

const mockAuthFetch = vi.fn();

vi.mock("@/lib/api", () => ({
  authFetch: (...args: Parameters<typeof mockAuthFetch>) => mockAuthFetch(...args),
  ApiError: class extends Error {
    code: string;
    status: number;
    constructor(code: string, message: string, status: number) {
      super(message);
      this.code = code;
      this.status = status;
    }
  },
}));

vi.mock("@/stores/auth", () => ({
  useAuthStore: {
    getState: () => ({ token: "test-token" }),
  },
}));

describe("admin-courses hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it("useAdminCourses fetches GET /admin/courses and returns data", async () => {
    const courses: Course[] = [{ id: "c1", title: "Course A" }];
    mockAuthFetch.mockResolvedValueOnce({ data: courses, next_cursor: "" });

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useAdminCourses(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses");
    expect(result.current.data).toEqual(courses);
  });

  it("useAdminCourse fetches GET /admin/courses/:id", async () => {
    const course: AdminCourseDetail = {
      id: "c1",
      title: "Course A",
      section_count: 2,
      lesson_count: 5,
    };
    mockAuthFetch.mockResolvedValueOnce(course);

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useAdminCourse("c1"), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1");
    expect(result.current.data).toEqual(course);
  });

  it("useCreateCourse posts to /admin/courses and invalidates list", async () => {
    const course: Course = { id: "c2", title: "Course B" };
    mockAuthFetch.mockResolvedValueOnce(course);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useCreateCourse(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ title: "Course B", level: "SMA" });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses", {
      method: "POST",
      body: JSON.stringify({ title: "Course B", level: "SMA" }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.list() });
  });

  it("useUpdateCourse patches /admin/courses/:id and invalidates list", async () => {
    const course: Course = { id: "c1", title: "Course A v2" };
    mockAuthFetch.mockResolvedValueOnce(course);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdateCourse(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ id: "c1", input: { title: "Course A v2" } });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1", {
      method: "PATCH",
      body: JSON.stringify({ title: "Course A v2" }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.list() });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.detail("c1") });
  });

  it("useAdminSections fetches GET /admin/courses/:id/sections", async () => {
    const sections: CourseSection[] = [
      { id: "s1", course_id: "c1", title: "Section 1", position: 1 },
    ];
    mockAuthFetch.mockResolvedValueOnce({ data: sections });

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useAdminSections("c1"), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1/sections");
    expect(result.current.data).toEqual(sections);
  });

  it("useCreateSection posts to /admin/courses/:id/sections and invalidates sections", async () => {
    const section: CourseSection = { id: "s1", course_id: "c1", title: "Section 1", position: 1 };
    mockAuthFetch.mockResolvedValueOnce(section);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useCreateSection("c1"), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ title: "Section 1" });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1/sections", {
      method: "POST",
      body: JSON.stringify({ title: "Section 1" }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.sections("c1") });
  });

  it("useUpdateSection puts /admin/courses/:id/sections/:sId and invalidates sections", async () => {
    const section: CourseSection = { id: "s1", course_id: "c1", title: "Renamed", position: 1 };
    mockAuthFetch.mockResolvedValueOnce(section);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdateSection("c1"), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ sectionId: "s1", input: { title: "Renamed" } });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1/sections/s1", {
      method: "PUT",
      body: JSON.stringify({ title: "Renamed" }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.sections("c1") });
  });

  it("useDeleteSection deletes /admin/courses/:id/sections/:sId and invalidates sections", async () => {
    mockAuthFetch.mockResolvedValueOnce(undefined);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useDeleteSection("c1"), { wrapper });

    await act(async () => {
      await result.current.mutateAsync("s1");
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1/sections/s1", { method: "DELETE" });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.sections("c1") });
  });

  it("useReorderSections patches reorder endpoint and invalidates sections", async () => {
    mockAuthFetch.mockResolvedValueOnce({ message: "sections reordered" });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useReorderSections("c1"), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ section_ids: ["s2", "s1"] });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1/sections/reorder", {
      method: "PATCH",
      body: JSON.stringify({ section_ids: ["s2", "s1"] }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.sections("c1") });
  });

  it("useCreateLesson posts to /admin/courses/:id/sections/:sId/lessons and invalidates sections", async () => {
    const lesson: Lesson = { id: "l1", section_id: "s1", title: "Lesson 1", duration_seconds: 120, position: 1 };
    mockAuthFetch.mockResolvedValueOnce(lesson);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useCreateLesson("c1", "s1"), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ title: "Lesson 1", duration: 120 });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1/sections/s1/lessons", {
      method: "POST",
      body: JSON.stringify({ title: "Lesson 1", duration: 120 }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.sections("c1") });
  });

  it("useUpdateLesson puts /admin/courses/:id/sections/:sId/lessons/:lId and invalidates sections", async () => {
    const lesson: Lesson = { id: "l1", section_id: "s1", title: "Renamed lesson", duration_seconds: 120, position: 1 };
    mockAuthFetch.mockResolvedValueOnce(lesson);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdateLesson("c1", "s1"), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ lessonId: "l1", input: { title: "Renamed lesson" } });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1/sections/s1/lessons/l1", {
      method: "PUT",
      body: JSON.stringify({ title: "Renamed lesson" }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.sections("c1") });
  });

  it("useDeleteLesson deletes /admin/courses/:id/sections/:sId/lessons/:lId and invalidates sections", async () => {
    mockAuthFetch.mockResolvedValueOnce(undefined);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useDeleteLesson("c1", "s1"), { wrapper });

    await act(async () => {
      await result.current.mutateAsync("l1");
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1/sections/s1/lessons/l1", {
      method: "DELETE",
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.sections("c1") });
  });

  it("useReorderLessons patches reorder endpoint and invalidates sections", async () => {
    mockAuthFetch.mockResolvedValueOnce({ message: "lessons reordered" });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useReorderLessons("c1", "s1"), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ lesson_ids: ["l2", "l1"] });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/courses/c1/sections/s1/lessons/reorder", {
      method: "PATCH",
      body: JSON.stringify({ lesson_ids: ["l2", "l1"] }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminCoursesKeys.sections("c1") });
  });
});

function wrapperFactory() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return {
    wrapper: ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    ),
    queryClient,
  };
}
