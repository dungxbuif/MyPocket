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
