"use client";

import { useMutation } from "@tanstack/react-query";
import { authFetch } from "@/lib/api";

export interface PresignUploadInput {
  filename: string;
  content_type: string;
}

export interface PresignUploadResponse {
  url: string;
  method: "PUT";
  key: string;
}

export function usePresignAdminAudioUpload() {
  return useMutation({
    mutationFn: ({ filename, content_type }: PresignUploadInput) =>
      authFetch<PresignUploadResponse>(
        `/admin/uploads/audio?filename=${encodeURIComponent(filename)}&content_type=${encodeURIComponent(content_type)}`,
        { method: "POST" }
      ),
  });
}

export function usePresignAdminImageUpload() {
  return useMutation({
    mutationFn: ({ filename, content_type }: PresignUploadInput) =>
      authFetch<PresignUploadResponse>(
        `/admin/uploads/image?filename=${encodeURIComponent(filename)}&content_type=${encodeURIComponent(content_type)}`,
        { method: "POST" }
      ),
  });
}
