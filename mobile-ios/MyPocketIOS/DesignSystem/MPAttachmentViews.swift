import SwiftUI
import UniformTypeIdentifiers

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

struct MPDocumentPicker: View {
    let title: String
    let count: Int
    let maxCount: Int
    @Binding var isPresented: Bool
    let onImport: ([URL]) -> Void

    var body: some View {
        MPButton(title, style: .secondary) { isPresented = true }
            .disabled(count >= maxCount)
            .fileImporter(
                isPresented: $isPresented,
                allowedContentTypes: [.image, .pdf],
                allowsMultipleSelection: true
            ) { result in
                if case .success(let urls) = result { onImport(Array(urls.prefix(maxCount - count))) }
            }
            .accessibilityHint("JPEG, PNG hoặc PDF; tối đa 5 MiB mỗi tệp")
    }
}
