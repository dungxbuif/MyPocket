import SwiftUI

struct MPAttachmentThumb: View {
    let name: String
    let remove: () -> Void

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: "doc.fill")
                .foregroundStyle(MPColor.action)
            Text(name)
                .font(.footnote)
                .lineLimit(1)
            Spacer()
            MPIconButton(systemName: "xmark", label: "Bỏ tệp", action: remove)
        }
        .padding(.horizontal, 12)
        .background(MPColor.surface, in: RoundedRectangle(cornerRadius: 8))
        .accessibilityElement(children: .contain)
    }
}

struct MPAttachmentPicker: View {
    let title: String
    let count: Int
    let maxCount: Int
    let add: () -> Void

    var body: some View {
        MPButton(title, style: .secondary, action: add)
            .disabled(count >= maxCount)
            .accessibilityHint("Đã chọn \(count) trên \(maxCount) tệp")
    }
}
