import SwiftUI
import UIKit

struct MPTextField: View {
    let title: String
    @Binding var text: String
    var placeholder: String = ""
    var error: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(.subheadline.weight(.semibold))
                .foregroundStyle(MPColor.ink)
            TextField(placeholder, text: $text)
                .textFieldStyle(.roundedBorder)
                .frame(minHeight: MPMetrics.controlHeight)
                .accessibilityLabel(title)
            if let error {
                Text(error)
                    .font(.footnote)
                    .foregroundStyle(MPColor.danger)
                    .accessibilityLabel("Lỗi: \(error)")
            }
        }
    }
}

struct MPAmountField: View {
    let title: String
    @Binding var text: String
    var error: String?

    var body: some View {
        MPTextField(title: title, text: $text, placeholder: "0", error: error)
            .keyboardType(.numberPad)
            .toolbar {
                ToolbarItemGroup(placement: .keyboard) {
                    Spacer()
                    Button("Xong") {
                        UIApplication.shared.sendAction(#selector(UIResponder.resignFirstResponder), to: nil, from: nil, for: nil)
                    }
                }
            }
    }
}

struct MPTextEditor: View {
    let title: String
    @Binding var text: String
    var placeholder: String = ""

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(.subheadline.weight(.semibold))
                .foregroundStyle(MPColor.ink)
            ZStack(alignment: .topLeading) {
                if text.isEmpty {
                    Text(placeholder)
                        .font(.body)
                        .foregroundStyle(MPColor.secondary)
                        .padding(12)
                }
                TextEditor(text: $text)
                    .frame(minHeight: 110)
                    .padding(8)
                    .scrollContentBackground(.hidden)
            }
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
            .accessibilityLabel(title)
        }
    }
}

struct MPDateField: View {
    let title: String
    @Binding var date: Date

    var body: some View {
        DatePicker(title, selection: $date, displayedComponents: .date)
            .datePickerStyle(.compact)
            .frame(minHeight: MPMetrics.controlHeight)
            .accessibilityLabel(title)
    }
}

struct MPDateTimeField: View {
    let title: String
    @Binding var date: Date

    var body: some View {
        DatePicker(title, selection: $date, displayedComponents: [.date, .hourAndMinute])
            .datePickerStyle(.compact)
            .frame(minHeight: MPMetrics.controlHeight)
            .accessibilityLabel(title)
    }
}

struct MPDateRangePicker: View {
    @Binding var start: Date
    @Binding var end: Date

    var body: some View {
        VStack(spacing: 8) {
            MPDateField(title: "Từ ngày", date: $start)
            MPDateField(title: "Đến ngày", date: $end)
        }
    }
}

/// Keeps wire values stable while presenting the semantic labels used by the
/// Vietnamese web UI. Unknown values intentionally fall back to the raw value
/// so a newly-added backend enum is still visible and selectable.
enum MPSelectionLabels {
    static func display(for value: String) -> String {
        switch value {
        case "expense": return "Chi"
        case "income": return "Thu"
        case "basic": return "Thông thường"
        case "goal": return "Tiết kiệm"
        case "credit": return "Tín dụng"
        case "none": return "Không đặt"
        case "fixed": return "Số tiền cố định"
        case "percent": return "Phần trăm thu nhập"
        case "daily": return "Mỗi ngày"
        case "weekly": return "Mỗi tuần"
        case "monthly": return "Mỗi tháng"
        case "yearly": return "Mỗi năm"
        case "buy": return "Mua"
        case "sell": return "Bán"
        case "purchase": return "Mua hàng"
        case "refund": return "Hoàn tiền"
        case "fee": return "Phí"
        case "interest": return "Lãi"
        case "increase": return "Tăng"
        case "decrease": return "Giảm"
        case "bug": return "Lỗi"
        case "feature": return "Tính năng"
        case "improvement": return "Cải thiện"
        case "finance:read": return "Đọc tài chính"
        case "advisor:read": return "Đọc trợ lý"
        case "advisor:chat": return "Chat trợ lý"
        case "feedback:read": return "Đọc phản hồi"
        case "feedback:write": return "Gửi phản hồi"
        default: return value
        }
    }
}

struct MPWalletSelector: View {
    let title: String
    let options: [String]
    @Binding var selection: String

    var body: some View {
        Menu {
            ForEach(options, id: \.self) { option in
                Button(MPSelectionLabels.display(for: option)) { selection = option }
            }
        } label: {
            HStack {
                Text(title)
                Spacer()
                Text(MPSelectionLabels.display(for: selection))
                    .foregroundStyle(MPColor.secondary)
                Image(systemName: "chevron.up.chevron.down")
                    .font(.caption)
            }
            .frame(minHeight: MPMetrics.controlHeight)
            .padding(.horizontal, 12)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        }
        .foregroundStyle(MPColor.ink)
        .accessibilityLabel("\(title): \(MPSelectionLabels.display(for: selection))")
    }
}

struct MPSelectionOption: Identifiable, Hashable {
    let id: String
    let label: String
}

/// Selector dành cho dữ liệu có identity thật. Không dùng label làm khóa vì
/// người dùng có thể tạo hai ví hoặc danh mục trùng tên.
struct MPOptionSelector: View {
    let title: String
    let options: [MPSelectionOption]
    @Binding var selectionID: String
    var emptyLabel: String = "Chưa chọn"

    private var selectedLabel: String {
        options.first(where: { $0.id == selectionID })?.label ?? emptyLabel
    }

    var body: some View {
        Menu {
            ForEach(options) { option in
                Button(option.label) { selectionID = option.id }
            }
        } label: {
            HStack {
                Text(title)
                Spacer()
                Text(selectedLabel)
                    .foregroundStyle(MPColor.secondary)
                Image(systemName: "chevron.up.chevron.down")
                    .font(.caption)
            }
            .frame(minHeight: MPMetrics.controlHeight)
            .padding(.horizontal, 12)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        }
        .foregroundStyle(MPColor.ink)
        .accessibilityLabel("\(title): \(selectedLabel)")
    }
}

/// Multi-select selector for relationships that may apply to more than one
/// wallet. An empty selection means "all wallets", matching the API contract.
struct MPMultiOptionSelector: View {
    let title: String
    let options: [MPSelectionOption]
    @Binding var selectionIDs: Set<String>
    var allLabel: String = "Tất cả"

    private var selectedLabel: String {
        if selectionIDs.isEmpty { return allLabel }
        return "Đã chọn \(selectionIDs.count)"
    }

    var body: some View {
        Menu {
            ForEach(options) { option in
                Button {
                    if selectionIDs.contains(option.id) {
                        selectionIDs.remove(option.id)
                    } else {
                        selectionIDs.insert(option.id)
                    }
                } label: {
                    HStack {
                        Text(option.label)
                        Spacer()
                        if selectionIDs.contains(option.id) {
                            Image(systemName: "checkmark")
                        }
                    }
                }
            }
            if !selectionIDs.isEmpty {
                Divider()
                Button("Bỏ chọn (tất cả ví)") { selectionIDs.removeAll() }
            }
        } label: {
            HStack {
                Text(title)
                Spacer()
                Text(selectedLabel)
                    .foregroundStyle(MPColor.secondary)
                Image(systemName: "checklist")
                    .font(.caption)
            }
            .frame(minHeight: MPMetrics.controlHeight)
            .padding(.horizontal, 12)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        }
        .foregroundStyle(MPColor.ink)
        .accessibilityLabel("\(title): \(selectedLabel)")
    }
}
