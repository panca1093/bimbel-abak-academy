import type { User } from "./types";

/** A profile is complete ⟺ school_id is non-null AND grade is non-null/non-empty.
 *  Grade may arrive as a string or number; both are valid non-empty states. */
export function isProfileComplete(user: User | null | undefined): boolean {
  if (!user) return false;
  return (
    !!user.school_id &&
    user.grade != null &&
    String(user.grade) !== ""
  );
}
