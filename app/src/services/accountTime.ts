type DateTimeParts = { year: number; month: number; day: number; hour: number; minute: number; second: number };

export const DEFAULT_ACCOUNT_TIMEZONE = "Asia/Ho_Chi_Minh";
const formatters = new Map<string, Intl.DateTimeFormat>();

export function isIanaTimezone(timezone: string): boolean {
  const value = timezone.trim();
  if (!value || value === "Local") return false;
  try {
    new Intl.DateTimeFormat("en-CA", { timeZone: value }).format(new Date());
    return true;
  } catch {
    return false;
  }
}

export function resolveClientTimezone(...candidates: Array<string | null | undefined>): string {
  for (const candidate of candidates) {
    const value = candidate?.trim() ?? "";
    if (isIanaTimezone(value)) return value;
  }
  try {
    const browserTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    if (isIanaTimezone(browserTimezone)) return browserTimezone;
  } catch {
    // Ignore browsers or embedded webviews that cannot expose a canonical IANA zone.
  }
  return DEFAULT_ACCOUNT_TIMEZONE;
}

function formatter(timezone: string): Intl.DateTimeFormat {
  const cached = formatters.get(timezone);
  if (cached) return cached;
  const created = new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric", month: "2-digit", day: "2-digit",
    hour: "2-digit", minute: "2-digit", second: "2-digit", hourCycle: "h23",
  });
  formatters.set(timezone, created);
  return created;
}

function partsAt(epochMs: number, timezone: string): DateTimeParts {
  const parts = formatter(timezone).formatToParts(new Date(epochMs));
  const value = Object.fromEntries(parts.map(part => [part.type, part.value]));
  return {
    year: Number(value.year), month: Number(value.month), day: Number(value.day),
    hour: Number(value.hour), minute: Number(value.minute), second: Number(value.second),
  };
}

function dateTimeKey(parts: DateTimeParts): string {
  return `${parts.year}-${String(parts.month).padStart(2, "0")}-${String(parts.day).padStart(2, "0")}T${String(parts.hour).padStart(2, "0")}:${String(parts.minute).padStart(2, "0")}:${String(parts.second).padStart(2, "0")}`;
}

function parseLocal(value: string): DateTimeParts {
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?$/.exec(value);
  if (!match) throw new Error("Invalid local date and time.");
  const parts = { year: Number(match[1]), month: Number(match[2]), day: Number(match[3]), hour: Number(match[4]), minute: Number(match[5]), second: Number(match[6] ?? 0) };
  const roundTrip = new Date(Date.UTC(parts.year, parts.month - 1, parts.day, parts.hour, parts.minute, parts.second));
  if (roundTrip.getUTCFullYear() !== parts.year || roundTrip.getUTCMonth() + 1 !== parts.month || roundTrip.getUTCDate() !== parts.day || parts.hour > 23 || parts.minute > 59 || parts.second > 59) {
    throw new Error("Invalid local date and time.");
  }
  return parts;
}

export function dateKeyAt(instant: string | Date, timezone: string): string {
  const epochMs = instant instanceof Date ? instant.getTime() : Date.parse(instant);
  if (!Number.isFinite(epochMs)) return "";
  const { year, month, day } = partsAt(epochMs, timezone);
  return `${year}-${String(month).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
}

export function accountMonthKey(instant: string | Date, timezone: string): string {
  const epochMs = instant instanceof Date ? instant.getTime() : Date.parse(instant);
  if (!Number.isFinite(epochMs)) return "";
  const { year, month } = partsAt(epochMs, timezone);
  return `${year}-${String(month).padStart(2, "0")}`;
}

export function todayDateKey(timezone: string, now = new Date()): string {
  return dateKeyAt(now, timezone);
}

export function weekDateRange(today: string, offset: number): { start: string; end: string } {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(today) || !Number.isInteger(offset)) throw new Error("Invalid week anchor.");
  const anchor = new Date(`${today}T12:00:00Z`);
  if (Number.isNaN(anchor.getTime()) || anchor.toISOString().slice(0, 10) !== today) throw new Error("Invalid week anchor.");
  const daysSinceMonday = (anchor.getUTCDay() + 6) % 7;
  anchor.setUTCDate(anchor.getUTCDate() - daysSinceMonday + offset * 7);
  const start = anchor.toISOString().slice(0, 10);
  anchor.setUTCDate(anchor.getUTCDate() + 6);
  return { start, end: anchor.toISOString().slice(0, 10) };
}

export function monthDateRange(month: string): { start: string; end: string } {
  const match = /^(\d{4})-(\d{2})$/.exec(month);
  if (!match) throw new Error("Invalid calendar month.");
  const year = Number(match[1]), monthNumber = Number(match[2]);
  if (monthNumber < 1 || monthNumber > 12) throw new Error("Invalid calendar month.");
  const start = new Date(Date.UTC(year, monthNumber - 1, 1));
  const end = new Date(Date.UTC(year, monthNumber, 0));
  return { start: start.toISOString().slice(0, 10), end: end.toISOString().slice(0, 10) };
}

export function localDateTimeAt(instant: string | Date, timezone: string): string {
  const epochMs = instant instanceof Date ? instant.getTime() : Date.parse(instant);
  if (!Number.isFinite(epochMs)) return "";
  const p = partsAt(epochMs, timezone);
  return `${dateKeyAt(instant, timezone)}T${String(p.hour).padStart(2, "0")}:${String(p.minute).padStart(2, "0")}`;
}

export function instantFromLocalDateTime(value: string, timezone: string): string {
  const desired = parseLocal(value);
  const naiveMs = Date.UTC(desired.year, desired.month - 1, desired.day, desired.hour, desired.minute, desired.second);
  const desiredKey = dateTimeKey(desired);
  const offsets = new Set<number>();

  // Sampling both sides of the requested wall time discovers both offsets around DST folds.
  for (let hours = -48; hours <= 48; hours += 3) {
    const sample = naiveMs + hours * 60 * 60 * 1000;
    const local = partsAt(sample, timezone);
    const representedMs = Date.UTC(local.year, local.month - 1, local.day, local.hour, local.minute, local.second);
    offsets.add(representedMs - sample);
  }

  const matches = [...offsets]
    .map(offset => naiveMs - offset)
    .filter(candidate => dateTimeKey(partsAt(candidate, timezone)) === desiredKey)
    .sort((a, b) => a - b);

  if (matches.length === 0) throw new Error("This local time does not exist in the selected timezone.");
  return new Date(matches[0]).toISOString();
}
