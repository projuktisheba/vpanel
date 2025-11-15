// utils/date.ts
import moment from "moment";

/**
 * Format a UTC date string or Date object to local timezone.
 * Always returns a string.
 */
export function formatUTCToLocal(
  date?: string | Date | null,
  formatStr = "YYYY-MM-DD HH:mm:ss"
): string {
  if (!date) return "-";
  return moment.utc(date).local().format("YYYY-MM-DD") || "-";
}
