"use client";

import { useRef, useState } from "react";
import { FileUp, Download, Loader2, CheckCircle } from "lucide-react";
import { toast } from "sonner";
import { useTranslation } from "@/lib/i18n";
import { AdminPageHeader } from "@/components/admin/AdminPageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  usePresignStudentBulkUpload,
  putFileToPresignedURL,
  useEnqueueStudentBulkImport,
} from "@/lib/hooks/admin-students-bulk";
import { useJobStatus } from "@/lib/hooks/jobs";

const TEMPLATE_HEADER = "name,school,jenjang,provinsi,kota,kecamatan,kode_pos,email";
const TEMPLATE_EXAMPLE_ROW = "Budi Santoso,SMAN 1 Jakarta,sma,Jawa Barat,Bandung,Coblong,40132,budi@example.com";

function buildTemplateCSV(): string {
  // One illustrative example row, per architecture decision 27.
  return `${TEMPLATE_HEADER}\n${TEMPLATE_EXAMPLE_ROW}\n`;
}

function downloadTemplate(): void {
  const csv = buildTemplateCSV();
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "bulk_register_template.csv";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

export default function BulkRegisterPage() {
  const { t } = useTranslation();

  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [jobId, setJobId] = useState<string | null>(null);

  const presign = usePresignStudentBulkUpload();
  const enqueue = useEnqueueStudentBulkImport();
  const job = useJobStatus(jobId);

  const isUploading = presign.isPending || enqueue.isPending;

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    setFile(e.target.files?.[0] ?? null);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!file) {
      toast.error(t("bulk_register_no_file"));
      return;
    }
    try {
      const presignResp = await presign.mutateAsync({
        filename: file.name,
        contentType: file.type || "text/csv",
      });
      try {
        await putFileToPresignedURL(
          presignResp.url,
          file,
          file.type || "text/csv",
        );
      } catch (err) {
        toast.error(
          err instanceof Error ? err.message : t("bulk_register_put_failed"),
        );
        return;
      }
      const enqueueResp = await enqueue.mutateAsync({
        fileKey: presignResp.key,
      });
      setJobId(enqueueResp.job_id);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : t("bulk_register_enqueue_failed"),
      );
    }
  }

  const jobData = job.data;
  const isTerminalSuccess = jobData?.status === "succeeded";
  const isTerminalFailed = jobData?.status === "failed";

  return (
    <div className="mx-auto max-w-3xl px-4 py-8 md:px-6 md:py-10 fade-in">
      <AdminPageHeader
        icon={FileUp}
        title={t("bulk_register_title")}
        description={t("bulk_register_subtitle")}
      />

      {/* Step 1: Download template */}
      <section className="mb-6">
        <h3 className="font-serif text-base font-semibold text-ink-900">
          1. {t("bulk_register_download_template")}
        </h3>
        <div className="mt-3">
          <Button
            type="button"
            variant="outline"
            onClick={downloadTemplate}
          >
            <Download className="mr-2 size-4" />
            {t("bulk_register_download_template")}
          </Button>
        </div>
      </section>

      {/* Step 2: Upload */}
      <section className="mb-6">
        <h3 className="font-serif text-base font-semibold text-ink-900">
          2. {t("bulk_register_upload")}
        </h3>
        <form onSubmit={handleSubmit} className="mt-3 space-y-4">
          <div className="grid gap-2">
            <Label htmlFor="bulk-register-file">
              {t("bulk_register_choose_file")}
            </Label>
            <Input
              ref={fileInputRef}
              id="bulk-register-file"
              type="file"
              accept=".csv,text/csv"
              onChange={handleFileChange}
              disabled={isUploading}
            />
            {file && (
              <p className="text-sm text-muted-foreground">{file.name}</p>
            )}
          </div>

          <Button type="submit" disabled={isUploading || !file}>
            {isUploading ? (
              <Loader2 className="mr-2 size-4 animate-spin" />
            ) : null}
            {isUploading
              ? t("bulk_register_uploading")
              : t("bulk_register_upload")}
          </Button>
        </form>
      </section>

      {/* Step 3: Progress + result */}
      {jobData && (
        <section className="md-card-outlined space-y-3 p-5">
          {isTerminalSuccess && (
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <CheckCircle className="size-5 text-success" />
                <h4 className="font-serif text-base font-semibold text-ink-900">
                  {t("bulk_register_success")}
                </h4>
              </div>
              {jobData.result_url && (
                <a
                  href={jobData.result_url}
                  className="inline-flex items-center gap-2 text-sm font-medium text-primary underline-offset-4 hover:underline"
                  download="bulk_register_result.csv"
                >
                  <Download className="size-4" />
                  {t("bulk_register_download_result")}
                </a>
              )}
            </div>
          )}

          {!isTerminalSuccess && !isTerminalFailed && (
            <div className="space-y-2">
              <p className="text-sm font-medium text-ink-900">
                {t("bulk_register_progress").replace(
                  "{pct}",
                  String(Math.round(jobData.progress ?? 0)),
                )}
              </p>
              <div className="h-2 w-full overflow-hidden rounded-full bg-surface-2">
                <div
                  className="h-full bg-primary transition-all"
                  style={{ width: `${Math.max(0, Math.min(100, jobData.progress ?? 0))}%` }}
                />
              </div>
            </div>
          )}

          {isTerminalFailed && (
            <div className="space-y-2">
              <h4 className="font-serif text-base font-semibold text-danger">
                {t("bulk_register_failed")}
              </h4>
              {jobData.error && (
                <p className="text-sm text-danger">{jobData.error}</p>
              )}
            </div>
          )}
        </section>
      )}
    </div>
  );
}
