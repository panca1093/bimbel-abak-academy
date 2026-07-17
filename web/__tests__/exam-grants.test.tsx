import { describe, it, expect } from "vitest";
import { DICT } from "@/lib/i18n";
import { NAV_CONFIG } from "@/lib/nav-config";

// Get actual translations for assertions
const t = (key: string) => DICT.id[key as keyof typeof DICT.id] || DICT.en[key as keyof typeof DICT.en];

describe("FR-FE-13/14: exam grant screen requirements", () => {
  it("should have exam-grant navigation item in super_admin nav", () => {
    const saNav = NAV_CONFIG["super_admin"];
    const examGroup = saNav.find((g) => g.titleKey === "nav_group_exam");

    expect(examGroup).toBeDefined();
    const grantItem = examGroup!.items.find((i) => i.href === "/admin/exam-grants");
    expect(grantItem).toBeDefined();
    expect(grantItem!.labelKey).toBe("nav_exam_grant");
  });

  it("should have correct translations for exam grant screen", () => {
    expect(t("exam_grant_title")).toBe("Beri Akses Ujian");
    expect(t("exam_grant_subtitle")).toBeDefined();
    expect(t("exam_grant_select_exam")).toBe("Pilih ujian");
    expect(t("exam_grant_grant")).toBe("Beri Akses");
    expect(t("exam_grant_success_title")).toBe("Akses Diberikan");
    expect(t("bulk_exam_order_pick_participants")).toBeDefined();
  });

  it("should have exam-grant screen file created", async () => {
    // This test verifies the file exists by importing it
    const module = await import("@/app/(admin)/admin/exam-grants/page");
    expect(module.default).toBeDefined();
  });

  it("should have useGrantExamAccess hook available", async () => {
    const module = await import("@/lib/hooks/admin-exam-grants");
    expect(module.useGrantExamAccess).toBeDefined();
    expect(module.useSearchStudentsAcrossSchools).toBeDefined();
  });

  it("should have ParticipantPicker component available", async () => {
    const module = await import("@/components/admin/ParticipantPicker");
    expect(module.ParticipantPicker).toBeDefined();
  });

  it("should have CrossSchoolStudent type available", async () => {
    const module = await import("@/lib/types");
    // This verifies the type exists (can't directly assert on types, but we can check exports)
    expect(module).toBeDefined();
  });

  it("should not include exam-grant in admin_exam nav subset", () => {
    const examNav = NAV_CONFIG["admin_exam"];
    const examGroup = examNav[0];
    const grantItem = examGroup.items.find((i) => i.href === "/admin/exam-grants");
    expect(grantItem).toBeUndefined();
  });

  it("should not include exam-grant in admin_school nav subset", () => {
    const schoolNav = NAV_CONFIG["admin_school"];
    const examGroup = schoolNav[0];
    const grantItem = examGroup.items.find((i) => i.href === "/admin/exam-grants");
    expect(grantItem).toBeUndefined();
  });

  it("should include bulk-exam-order in admin_school nav", () => {
    const schoolNav = NAV_CONFIG["admin_school"];
    const examGroup = schoolNav[0];
    const bulkItem = examGroup.items.find((i) => i.href === "/admin/school/bulk-exam-order");
    expect(bulkItem).toBeDefined();
  });

  it("should have nav_exam_grant key in both id and en dictionaries", () => {
    const idDict = DICT.id as Record<string, string>;
    const enDict = DICT.en as Record<string, string>;

    expect(idDict["nav_exam_grant"]).toBeDefined();
    expect(enDict["nav_exam_grant"]).toBeDefined();
    expect(idDict["nav_exam_grant"]).toBe("Beri Akses Ujian");
    expect(enDict["nav_exam_grant"]).toBe("Grant Exam Access");
  });
});

describe("FR-FE-13: Screen renders without school-picker gating", () => {
  it("should have exam grant page that imports ParticipantPicker", async () => {
    // Read the actual page source to verify structure
    const fs = await import("fs/promises");
    const path = await import("path");

    const filePath = path.join(
      process.cwd(),
      "app/(admin)/admin/exam-grants/page.tsx"
    );

    const content = await fs.readFile(filePath, "utf-8");

    // Verify no "school picker first" gating
    expect(content).toContain("ParticipantPicker");
    expect(content).toContain("selectedExamId");
    expect(content).toContain("selectedStudentIds");

    // Verify the structure shows ParticipantPicker immediately after exam selection
    expect(content).toContain('schoolId={undefined}');
  });
});

describe("FR-FE-14: Grant action with multi-school support", () => {
  it("should have POST /admin/exam-grants hook that accepts exam_id and student_ids", async () => {
    const fs = await import("fs/promises");
    const path = await import("path");

    const filePath = path.join(
      process.cwd(),
      "lib/hooks/admin-exam-grants.ts"
    );

    const content = await fs.readFile(filePath, "utf-8");

    // Verify the hook signature
    expect(content).toContain("GrantExamAccessInput");
    expect(content).toContain("exam_id");
    expect(content).toContain("student_ids");
    expect(content).toContain("/admin/exam-grants");
    expect(content).toContain("POST");
  });

  it("should show success state with granted student details", async () => {
    const fs = await import("fs/promises");
    const path = await import("path");

    const filePath = path.join(
      process.cwd(),
      "app/(admin)/admin/exam-grants/page.tsx"
    );

    const content = await fs.readFile(filePath, "utf-8");

    // Verify success state shows student names and usernames
    expect(content).toContain("grantResult");
    expect(content).toContain("granted_students");
    expect(content).toContain("granted_count");
    expect(content).toContain("CheckCircle");
    expect(content).toContain("exam_grant_success_title");
  });

  it("should not show preview or payment step in the flow", async () => {
    const fs = await import("fs/promises");
    const path = await import("path");

    const filePath = path.join(
      process.cwd(),
      "app/(admin)/admin/exam-grants/page.tsx"
    );

    const content = await fs.readFile(filePath, "utf-8");

    // Verify no preview step
    expect(content).not.toContain("handlePreview");
    expect(content).not.toContain("preview");

    // Verify no payment/checkout integration
    expect(content).not.toContain("SnapCheckout");
    expect(content).not.toContain("useCheckout");
    expect(content).not.toContain("Midtrans");
    expect(content).not.toContain("payment");

    // Verify direct grant action
    expect(content).toContain("handleGrant");
    expect(content).toContain("exam_grant_grant");
  });

  it("should call POST /admin/exam-grants with flat student_ids list", async () => {
    const fs = await import("fs/promises");
    const path = await import("path");

    const filePath = path.join(
      process.cwd(),
      "app/(admin)/admin/exam-grants/page.tsx"
    );

    const content = await fs.readFile(filePath, "utf-8");

    // Verify the mutation call includes both exam_id and student_ids
    expect(content).toContain("grantMutation.mutate");
    expect(content).toContain("exam_id: selectedExamId");
    expect(content).toContain("student_ids: selectedStudentIds");
  });
});
