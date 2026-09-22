import { accountMonthKey } from "./accountTime.ts";

export function isMonthKey(value: string | undefined): value is string {
  if (!value || !/^\d{4}-(0[1-9]|1[0-2])$/.test(value)) return false;
  const [year, month] = value.split("-").map(Number);
  const date = new Date(Date.UTC(year, month - 1, 1));
  return date.getUTCFullYear() === year && date.getUTCMonth() + 1 === month;
}

export function resolveMonthKey(requested: string | undefined, timezone: string, now = new Date()): string {
  return isMonthKey(requested) ? requested : accountMonthKey(now, timezone);
}

export function monthKeyForInstant(instant: string, timezone: string, now = new Date()): string {
  const date = new Date(instant);
  return Number.isFinite(date.getTime()) ? accountMonthKey(date, timezone) : accountMonthKey(now, timezone);
}

export function shiftMonthKey(value: string, offset: number): string {
  if (!isMonthKey(value) || !Number.isInteger(offset)) throw new Error("Invalid calendar month shift.");
  const [year, month] = value.split("-").map(Number);
  const shifted = new Date(Date.UTC(year, month - 1 + offset, 1));
  return `${shifted.getUTCFullYear()}-${String(shifted.getUTCMonth() + 1).padStart(2, "0")}`;
}

export function monthOptions(value: string, monthsBefore = 24, monthsAfter = 0): string[] {
  if (!isMonthKey(value) || !Number.isInteger(monthsBefore) || monthsBefore < 0 || !Number.isInteger(monthsAfter) || monthsAfter < 0) {
    throw new Error("Invalid month picker range.");
  }
  return Array.from({ length: monthsBefore + monthsAfter + 1 }, (_, index) => shiftMonthKey(value, index - monthsBefore));
}

export function monthLabel(value: string, locale = "vi-VN"): string {
  if (!isMonthKey(value)) throw new Error("Invalid calendar month.");
  const [year, month] = value.split("-").map(Number);
  const date = new Date(Date.UTC(year, month - 1, 1, 12));
  return new Intl.DateTimeFormat(locale, { month: "long", year: "numeric", timeZone: "UTC" }).format(date);
}

export function jarUsagePercent(spent: number, allocation: number | null | undefined): number | null {
  if (allocation == null || !Number.isFinite(allocation) || allocation <= 0 || !Number.isFinite(spent)) return null;
  return Math.max(0, spent / allocation * 100);
}

export function jarAssignmentNeedsSelection(
  jarID: string | null | undefined,
  options: { jar_id: string; active: boolean }[],
  originalInstant: string | undefined,
  currentInstant: string,
): boolean {
  if (!jarID || originalInstant === currentInstant) return false;
  return options.some(option => option.jar_id === jarID && !option.active);
}
