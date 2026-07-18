"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";
import type { ExamListItem, Order, CheckoutResult } from "@/lib/types";

// ── Query keys ────────────────────────────────────────────────────────────

export const bulkExamOrderKeys = {
  all: ["admin", "bulkExamOrders"] as const,
  orderableExams: () => [...bulkExamOrderKeys.all, "orderableExams"] as const,
  preview: (examId: string) => [...bulkExamOrderKeys.all, "preview", examId] as const,
};

// ── Types ─────────────────────────────────────────────────────────────────

export interface BulkExamOrderStudent {
  id: string;
  name: string;
  username: string;
  jenjang: string;
  grade?: number;
}

export interface BulkExamOrderPreview {
  exam: ExamListItem;
  students: BulkExamOrderStudent[];
  total_price: number;
}

export interface CreateBulkOrderInput {
  exam_id: string;
  student_ids: string[];
}

// ── Hooks ─────────────────────────────────────────────────────────────────

/**
 * Fetch exams that have a published product (can be ordered via bulk).
 * GET /admin/exams?orderable=true  (or equivalent backend endpoint)
 */
export function useOrderableExams() {
  return useQuery({
    queryKey: bulkExamOrderKeys.orderableExams(),
    queryFn: () =>
      authFetch<{ data: ExamListItem[] }>("/admin/exams?orderable=true"),
  });
}

/**
 * Preview a bulk exam order: given an exam + student list, what will the
 * order look like.
 * POST /admin/bulk-exam-orders/preview
 */
export function usePreviewBulkExamOrder() {
  return useMutation({
    mutationFn: (input: CreateBulkOrderInput) =>
      authFetch<BulkExamOrderPreview>("/admin/bulk-exam-orders/preview", {
        method: "POST",
        body: JSON.stringify(input),
      }),
  });
}

/**
 * Create a bulk exam order.
 * POST /admin/bulk-exam-orders
 */
export function useCreateBulkExamOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateBulkOrderInput) =>
      authFetch<Order>("/admin/bulk-exam-orders", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: bulkExamOrderKeys.all });
    },
  });
}
