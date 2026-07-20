"use client";

import { useRef, useState, type CSSProperties, type PointerEvent as ReactPointerEvent } from "react";
import { useTranslation } from "@/lib/i18n";
import type { DICT } from "@/lib/i18n";
import type { CertificateLayout, CertificateLayoutField } from "@/lib/types";

interface CertificateFieldEditorProps {
  layout: CertificateLayout;
  onChange: (fields: CertificateLayoutField[]) => void;
}

interface DragState {
  fieldId: string;
  offsetPx: { x: number; y: number };
  currentPx: { x: number; y: number };
}

const FIELD_LABEL_KEY: Record<string, keyof (typeof DICT)["id"]> = {
  title: "certificate_field_title",
  subtitle: "certificate_field_subtitle",
  student_name: "certificate_field_student_name",
  exam_title: "certificate_field_exam_title",
  completion_text: "certificate_field_completion_text",
  date: "certificate_field_date",
  certificate_number: "certificate_field_certificate_number",
  logo: "certificate_field_logo",
};

// Coordinate contract (FR-1): x_mm,y_mm is the field box's top-left corner in
// millimetres, origin top-left, Y down - identical to the renderer. The only
// conversion here is the uniform scale mm = px * (page_width_mm /
// preview_width_px); there is no Y-axis flip.
function pxToMm(px: number, widthPx: number, pageWidthMm: number): number {
  return px * (pageWidthMm / widthPx);
}

// clampFieldPosition mirrors the backend's ValidateLayout bounds (Task 3) so
// a drop the editor accepts never comes back as a 422 on save: x_mm,y_mm is
// the box's top-left corner, so x clamps against the box's own width, and a
// logo's y clamps against its height too.
export function clampFieldPosition(
  field: CertificateLayoutField,
  page: CertificateLayout["page"],
): CertificateLayoutField {
  const maxX = Math.max(0, page.width_mm - field.w_mm);
  const x_mm = Math.min(Math.max(field.x_mm, 0), maxX);
  const maxY =
    field.id === "logo" ? Math.max(0, page.height_mm - (field.h_mm ?? 0)) : page.height_mm;
  const y_mm = Math.min(Math.max(field.y_mm, 0), maxY);
  return { ...field, x_mm, y_mm };
}

export function CertificateFieldEditor({ layout, onChange }: CertificateFieldEditorProps) {
  const { t } = useTranslation();
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [drag, setDrag] = useState<DragState | null>(null);
  const { page, fields } = layout;
  const visibleFields = fields.filter((f) => f.visible);

  function commitField(fieldId: string, patch: Partial<CertificateLayoutField>) {
    const field = fields.find((f) => f.id === fieldId);
    if (!field) return;
    const updated = clampFieldPosition({ ...field, ...patch }, page);
    onChange(fields.map((f) => (f.id === fieldId ? updated : f)));
  }

  function handlePointerDown(e: ReactPointerEvent<HTMLDivElement>, field: CertificateLayoutField) {
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) return;
    const boxLeftPx = (field.x_mm / page.width_mm) * rect.width;
    const boxTopPx = (field.y_mm / page.height_mm) * rect.height;
    const target = e.currentTarget;
    if (typeof target.setPointerCapture === "function") {
      try {
        target.setPointerCapture(e.pointerId);
      } catch {
        // Not every environment (e.g. jsdom) implements pointer capture; the
        // drag still works from move/up events bubbling normally.
      }
    }
    setDrag({
      fieldId: field.id,
      offsetPx: { x: e.clientX - rect.left - boxLeftPx, y: e.clientY - rect.top - boxTopPx },
      currentPx: { x: boxLeftPx, y: boxTopPx },
    });
  }

  function handlePointerMove(e: ReactPointerEvent<HTMLDivElement>) {
    if (!drag) return;
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) return;
    setDrag({
      ...drag,
      currentPx: {
        x: e.clientX - rect.left - drag.offsetPx.x,
        y: e.clientY - rect.top - drag.offsetPx.y,
      },
    });
  }

  function handlePointerUp() {
    if (!drag) return;
    const rect = containerRef.current?.getBoundingClientRect();
    if (rect) {
      const x_mm = pxToMm(drag.currentPx.x, rect.width, page.width_mm);
      const y_mm = pxToMm(drag.currentPx.y, rect.width, page.width_mm);
      commitField(drag.fieldId, { x_mm, y_mm });
    }
    setDrag(null);
  }

  return (
    <div className="space-y-3">
      <div>
        <span className="text-xs font-semibold uppercase tracking-wide text-ink-500">
          {t("certificate_field_editor_label")}
        </span>
        <p className="text-xs text-muted-foreground">{t("certificate_field_editor_hint")}</p>
      </div>

      <div
        ref={containerRef}
        data-testid="certificate-field-editor-canvas"
        className="relative w-full select-none rounded-md border border-line bg-ink-50"
        style={{ aspectRatio: `${page.width_mm} / ${page.height_mm}` }}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
      >
        {visibleFields.map((field) => {
          const isDragging = drag?.fieldId === field.id;
          const widthPct = (field.w_mm / page.width_mm) * 100;
          const style: CSSProperties = isDragging
            ? { position: "absolute", left: drag!.currentPx.x, top: drag!.currentPx.y, width: `${widthPct}%` }
            : {
                position: "absolute",
                left: `${(field.x_mm / page.width_mm) * 100}%`,
                top: `${(field.y_mm / page.height_mm) * 100}%`,
                width: `${widthPct}%`,
              };
          return (
            <div
              key={field.id}
              data-testid={`certificate-field-box-${field.id}`}
              className="cursor-grab touch-none rounded border border-dashed border-brand-400 bg-brand-50/70 px-1 py-0.5 text-[10px] leading-tight text-brand-800 active:cursor-grabbing"
              style={style}
              onPointerDown={(e) => handlePointerDown(e, field)}
            >
              {t(FIELD_LABEL_KEY[field.id] ?? "certificate_field_title")}
            </div>
          );
        })}
      </div>

      <div className="space-y-1">
        {visibleFields.map((field) => (
          <div key={field.id} className="flex items-center gap-2 text-xs">
            <span className="w-28 truncate text-ink-600">
              {t(FIELD_LABEL_KEY[field.id] ?? "certificate_field_title")}
            </span>
            <label className="flex items-center gap-1">
              x(mm)
              <input
                type="number"
                aria-label={`x_mm ${field.id}`}
                value={Math.round(field.x_mm * 10) / 10}
                onChange={(e) => commitField(field.id, { x_mm: Number(e.target.value) })}
                className="w-16 rounded border border-line px-1 py-0.5"
              />
            </label>
            <label className="flex items-center gap-1">
              y(mm)
              <input
                type="number"
                aria-label={`y_mm ${field.id}`}
                value={Math.round(field.y_mm * 10) / 10}
                onChange={(e) => commitField(field.id, { y_mm: Number(e.target.value) })}
                className="w-16 rounded border border-line px-1 py-0.5"
              />
            </label>
          </div>
        ))}
      </div>
    </div>
  );
}
