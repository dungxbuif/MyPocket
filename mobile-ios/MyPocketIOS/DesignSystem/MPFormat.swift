import Foundation

enum MPFormat {
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

    static func signedMoney(_ value: Int64) -> String {
        if value == 0 { return money(0) }
        return (value > 0 ? "+" : "−") + money(abs(value))
    }

    static func percentageFromBPS(_ value: Int64?) -> String {
        guard let value else { return "Chưa có dữ liệu" }
        return String(format: "%.1f%%", Double(value) / 100.0)
    }

    static func parseCalendarDate(_ raw: String?) -> Date? {
        guard let raw, !raw.isEmpty else { return nil }
        return calendarDate.date(from: raw)
    }

    static func calendarDateString(_ value: Date) -> String {
        calendarDate.string(from: value)
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

    static func todayISO(timeZone: TimeZone = .current) -> String {
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.timeZone = timeZone
        return formatter.string(from: Date())
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
