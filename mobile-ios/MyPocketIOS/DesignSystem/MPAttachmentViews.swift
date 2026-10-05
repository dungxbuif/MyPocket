import SwiftUI
import UniformTypeIdentifiers
import UIKit

struct MPAttachmentThumb: View {
    let name: String
    let data: Data?
    let remove: () -> Void

    init(name: String, data: Data? = nil, remove: @escaping () -> Void) {
        self.name = name
        self.data = data
        self.remove = remove
    }

    var body: some View {
        HStack(spacing: 8) {
            if let data, let image = UIImage(data: data) {
                Image(uiImage: image)
                    .resizable()
                    .scaledToFill()
                    .frame(width: 44, height: 44)
                    .clipShape(RoundedRectangle(cornerRadius: 6))
            } else {
                Image(systemName: "doc.fill")
                    .foregroundStyle(MPColor.action)
                    .frame(width: 44, height: 44)
            }
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
