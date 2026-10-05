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

    static func money(_ value: Int64) -> String {
        currency.string(from: NSNumber(value: value)) ?? "(value) ₫"
    }

    static func shortDate(_ raw: String) -> String {
        if let date = ISO8601DateFormatter().date(from: raw) { return self.date.string(from: date) }
        if let date = self.date.date(from: raw) { return self.date.string(from: date) }
        return raw
    }

    static var todayISO: String {
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter.string(from: Date())
    }

    static var currentMonth: String {
        String(todayISO.prefix(7))
    }
}
