import SwiftUI

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

struct MPWalletSelector: View {
    let title: String
    let options: [String]
    @Binding var selection: String

    var body: some View {
        Menu {
            ForEach(options, id: \.self) { option in
                Button(option) { selection = option }
            }
        } label: {
            HStack {
                Text(title)
                Spacer()
                Text(selection)
                    .foregroundStyle(MPColor.secondary)
                Image(systemName: "chevron.up.chevron.down")
                    .font(.caption)
            }
            .frame(minHeight: MPMetrics.controlHeight)
            .padding(.horizontal, 12)
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        }
        .foregroundStyle(MPColor.ink)
        .accessibilityLabel("\(title): \(selection)")
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
