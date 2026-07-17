"use client";

import { useRef, type ChangeEvent } from "react";
import { toast } from "sonner";
import { Upload } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { fileUrl } from "@/lib/api";
import { usePresignAdminAudioUpload } from "@/lib/hooks/admin-uploads";

interface AudioUploadInputProps {
  value: string;
  onChange: (url: string) => void;
  disabled?: boolean;
  placeholder?: string;
  id?: string;
  "aria-label"?: string;
}

export function AudioUploadInput({
  value,
  onChange,
  disabled,
  placeholder,
  id,
  "aria-label": ariaLabel,
}: AudioUploadInputProps) {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const presign = usePresignAdminAudioUpload();

  async function handleFileSelected(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      const presigned = await presign.mutateAsync({
        filename: file.name,
        content_type: file.type,
      });

      const uploadRes = await fetch(presigned.url, {
        method: "PUT",
        body: file,
        headers: { "Content-Type": file.type },
      });

      if (!uploadRes.ok) {
        throw new Error(`Upload failed: ${uploadRes.status}`);
      }

      const url = fileUrl(presigned.key) ?? presigned.key;
      onChange(url);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Upload failed");
    } finally {
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  const uploading = presign.isPending;

  return (
    <div className="flex gap-2">
      <Input
        id={id}
        aria-label={ariaLabel}
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled || uploading}
      />
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={() => fileInputRef.current?.click()}
        disabled={disabled || uploading}
        aria-label="Upload audio file"
      >
        {uploading ? (
          <Upload className="size-4 animate-pulse" />
        ) : (
          <Upload className="size-4" />
        )}
      </Button>
      <input
        ref={fileInputRef}
        type="file"
        accept="audio/*"
        hidden
        data-testid={`audio-upload-input-${id}`}
        onChange={handleFileSelected}
      />
    </div>
  );
}
