export function dateKey(date: Date): string {
  return [date.getFullYear(), String(date.getMonth() + 1).padStart(2, "0"), String(date.getDate()).padStart(2, "0")].join("-");
}
export function parseDay(value: string): Date {
  const [year, month, day] = value.slice(0, 10).split("-").map(Number);
  return new Date(year, month - 1, day, 12);
}
export function shiftDay(value: string, days: number): string {
  const date = parseDay(value);
  if (!Number.isFinite(date.getTime())) return value;
  date.setDate(date.getDate() + days);
  return dateKey(date) + value.slice(10);
}
export function monthDays(year: number, month: number): string[] {
  const first = new Date(year, month, 1, 12);
  first.setDate(1 - (first.getDay() + 6) % 7);
  return Array.from({ length: 42 }, (_, i) => dateKey(new Date(first.getFullYear(), first.getMonth(), first.getDate() + i, 12)));
}
export function calendarLabel(value: string): string {
  const date = parseDay(value);
  return Number.isFinite(date.getTime()) ? new Intl.DateTimeFormat("vi-VN", { weekday: "long", day: "2-digit", month: "2-digit", year: "numeric" }).format(date) : "Chọn ngày";
}
