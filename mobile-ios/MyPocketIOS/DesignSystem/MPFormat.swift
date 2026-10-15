import Foundation
import SwiftUI

/// Syntax/locale normalization only; domain validators still own positivity,
/// scale and range restrictions particular to each financial operation.
enum MPNumberInputKind: Sendable {
    case decimal
    case integer
    case percentage

    func normalized(_ raw: String) -> String? {
        let value = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !value.isEmpty else { return "" }
        let normalized = value.replacingOccurrences(of: ",", with: ".")
        let parts = normalized.split(separator: ".", omittingEmptySubsequences: false)
        guard parts.count <= (self == .integer ? 1 : 2),
              parts.allSatisfy({ part in
                  !part.isEmpty && part.utf8.allSatisfy { (48...57).contains($0) }
              }) else { return nil }
        if self == .percentage {
            let whole = parts[0].drop(while: { $0 == "0" })
            guard whole.count <= 3, let number = Int(whole.isEmpty ? "0" : String(whole)), number <= 100 else { return nil }
            if number == 100 && parts.count == 2 && !parts[1].allSatisfy({ $0 == "0" }) { return nil }
        }
        return normalized
    }
}

private struct MPFinancialValuesMaskedKey: EnvironmentKey {
    static let defaultValue = false
}

extension EnvironmentValues {
    var mpFinancialValuesMasked: Bool {
        get { self[MPFinancialValuesMaskedKey.self] }
        set { self[MPFinancialValuesMaskedKey.self] = newValue }
    }
}

struct MPSavingsProgress: Equatable, Sendable {
    let remaining: Int64
    let percentage: Double
    let reached: Bool
}

enum MPFormat {
    static let maskedValue = "••••••"

    static let currency: NumberFormatter = {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "VND"
        formatter.currencySymbol = "₫"
        formatter.maximumFractionDigits = 0
        formatter.locale = Locale(identifier: "vi_VN")
        return formatter
    }()

    static let date: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "dd/MM/yyyy"
        formatter.locale = Locale(identifier: "vi_VN")
        return formatter
    }()

    private static let calendarDate: DateFormatter = {
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.timeZone = TimeZone(secondsFromGMT: 0)
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter
    }()

    static func money(_ value: Int64) -> String {
        currency.string(from: NSNumber(value: value)) ?? "(value) ₫"
    }

    static func money(_ value: Int64, masked: Bool) -> String {
        masked ? maskedValue : money(value)
    }

    /// Parses the integer contract used by money fields without mutating the
    /// user's draft. Financial input must contain ASCII decimal digits only;
    /// malformed characters are rejected instead of silently filtered out.
    static func parseNonNegativeInteger(_ raw: String, allowingEmpty: Bool = false) -> Int64? {
        let value = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        if value.isEmpty { return allowingEmpty ? 0 : nil }
        guard value.unicodeScalars.allSatisfy({ $0.value >= 48 && $0.value <= 57 }) else { return nil }
        guard let parsed = Int64(value), parsed >= 0 else { return nil }
        return parsed
    }

    static func parsePositiveInteger(_ raw: String) -> Int64? {
        guard let parsed = parseNonNegativeInteger(raw), parsed > 0 else { return nil }
        return parsed
    }

    static func signedMoney(_ value: Int64) -> String {
        if value == 0 { return money(0) }
        return (value > 0 ? "+" : "−") + money(abs(value))
    }

    static func signedMoney(_ value: Int64, masked: Bool) -> String {
        masked ? maskedValue : signedMoney(value)
    }

    static func savingsProgress(balance: Int64, target: Int64) -> MPSavingsProgress {
        let safeTarget = max(0, target)
        let safeBalance = max(0, balance)
        let remaining = max(0, safeTarget - safeBalance)
        let reached = safeTarget > 0 && safeBalance >= safeTarget
        let percentage = safeTarget > 0 ? min(1, Double(safeBalance) / Double(safeTarget)) : 0
        return MPSavingsProgress(remaining: remaining, percentage: percentage, reached: reached)
    }

    static func calendarDayDifference(to targetDate: String?, from now: Date = Date(), timeZone: TimeZone = .current) -> Int? {
        guard let target = parseCalendarDate(targetDate, timeZone: timeZone) else { return nil }
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        return calendar.dateComponents([.day], from: calendar.startOfDay(for: now), to: calendar.startOfDay(for: target)).day
    }

    static func percentageFromBPS(_ value: Int64?) -> String {
        guard let value else { return "Chưa có dữ liệu" }
        return String(format: "%.1f%%", Double(value) / 100.0)
    }

    static func parseCalendarDate(_ raw: String?) -> Date? {
        guard let raw, !raw.isEmpty else { return nil }
        return calendarDate.date(from: raw)
    }

    static func parseCalendarDate(_ raw: String?, timeZone: TimeZone) -> Date? {
        guard let raw, !raw.isEmpty else { return nil }
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.timeZone = timeZone
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter.date(from: raw)
    }

    static func calendarDateString(_ value: Date) -> String {
        calendarDate.string(from: value)
    }

    static func calendarDateString(_ value: Date, timeZone: TimeZone) -> String {
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.timeZone = timeZone
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter.string(from: value)
    }

    static func parseInstant(_ raw: String?) -> Date? {
        guard let raw, !raw.isEmpty else { return nil }
        let fractional = ISO8601DateFormatter()
        fractional.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        let standard = ISO8601DateFormatter()
        standard.formatOptions = [.withInternetDateTime]
        return fractional.date(from: raw) ?? standard.date(from: raw) ?? calendarDate.date(from: raw)
    }

    static func instantString(_ value: Date) -> String {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter.string(from: value)
    }

    static func shortDate(_ raw: String, timeZone: TimeZone = .current) -> String {
        guard let parsed = parseInstant(raw) else { return raw }
        let formatter = DateFormatter()
        formatter.dateFormat = "dd/MM/yyyy"
        formatter.locale = Locale(identifier: "vi_VN")
        formatter.timeZone = timeZone
        return formatter.string(from: parsed)
    }

    static func shortDateTime(_ raw: String, timeZone: TimeZone = .current) -> String {
        guard let parsed = parseInstant(raw) else { return raw }
        let formatter = DateFormatter()
        formatter.dateFormat = "dd/MM/yyyy HH:mm"
        formatter.locale = Locale(identifier: "vi_VN")
        formatter.timeZone = timeZone
        return formatter.string(from: parsed)
    }

    static func todayISO(timeZone: TimeZone = .current) -> String {
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.timeZone = timeZone
        return formatter.string(from: Date())
    }

    /// Returns the account-local Monday–Sunday week containing `now`.
    /// `offset` is measured in whole weeks; zero is the current week, negative
    /// values are previous weeks and positive values are future weeks.
    static func weekRange(now: Date = Date(), offset: Int = 0, timeZone: TimeZone = .current) -> (start: Date, endExclusive: Date) {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        calendar.firstWeekday = 2 // Monday, matching the web ledger contract.
        calendar.minimumDaysInFirstWeek = 1
        let currentStart = calendar.dateInterval(of: .weekOfYear, for: now)?.start ?? calendar.startOfDay(for: now)
        let start = calendar.date(byAdding: .weekOfYear, value: offset, to: currentStart) ?? currentStart
        let endExclusive = calendar.date(byAdding: .day, value: 7, to: start) ?? start
        return (start, endExclusive)
    }

    static func weekRangeLabel(start: Date, endExclusive: Date, timeZone: TimeZone = .current) -> String {
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.locale = Locale(identifier: "vi_VN")
        formatter.timeZone = timeZone
        formatter.dateFormat = "dd/MM/yyyy"
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        let endInclusive = calendar.date(byAdding: .day, value: -1, to: endExclusive) ?? endExclusive
        return "\(formatter.string(from: start)) – \(formatter.string(from: endInclusive))"
    }

    static func currentMonth(timeZone: TimeZone = .current) -> String {
        String(todayISO(timeZone: timeZone).prefix(7))
    }

    static func monthLabel(_ value: String) -> String {
        guard value.count == 7 else { return "Không rõ thời gian" }
        let parts = value.split(separator: "-")
        guard parts.count == 2, parts[0].count == 4, parts[1].count == 2,
              let month = Int(parts[1]), (1...12).contains(month) else { return "Không rõ thời gian" }
        return "Tháng \(parts[1]) năm \(parts[0])"
    }

    static func monthOptions(around month: String, past: Int = 24, future: Int = 1, timeZone: TimeZone = .current) -> [String] {
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = timeZone
        formatter.dateFormat = "yyyy-MM-dd"
        guard let anchor = formatter.date(from: "\(month)-01") else { return [currentMonth(timeZone: timeZone)] }
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        return (-past...future).compactMap { offset in
            guard let date = calendar.date(byAdding: .month, value: offset, to: anchor) else { return nil }
            formatter.dateFormat = "yyyy-MM"
            return formatter.string(from: date)
        }
    }

    static func accountMonth(of raw: String, timeZone: TimeZone) -> String? {
        guard let parsed = parseInstant(raw) else { return nil }
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = timeZone
        formatter.dateFormat = "yyyy-MM"
        return formatter.string(from: parsed)
    }

    static func monthRange(month: String, timeZone: TimeZone) -> (start: String, end: String)? {
        let parser = DateFormatter()
        parser.calendar = Calendar(identifier: .gregorian)
        parser.locale = Locale(identifier: "en_US_POSIX")
        parser.timeZone = timeZone
        parser.dateFormat = "yyyy-MM-dd"
        guard let first = parser.date(from: "\(month)-01") else { return nil }
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        guard let next = calendar.date(byAdding: .month, value: 1, to: first), let last = calendar.date(byAdding: .day, value: -1, to: next) else { return nil }
        return (parser.string(from: first), parser.string(from: last))
    }

    static var todayISO: String { todayISO(timeZone: .current) }
    static var currentMonth: String { currentMonth(timeZone: .current) }
}
