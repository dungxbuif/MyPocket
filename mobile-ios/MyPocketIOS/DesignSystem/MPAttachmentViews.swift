import SwiftUI
import UniformTypeIdentifiers
import UIKit
import PhotosUI
import ImageIO

@MainActor
enum MPImageAttachmentPolicy {
    static let maxSourceBytes = 25 << 20

    static func normalizedPNG(_ data: Data) -> Data? {
        guard data.count <= maxSourceBytes,
              let source = CGImageSourceCreateWithData(data as CFData, nil),
              let thumbnail = CGImageSourceCreateThumbnailAtIndex(source, 0, [
                kCGImageSourceCreateThumbnailFromImageAlways: true,
                kCGImageSourceCreateThumbnailWithTransform: true,
                kCGImageSourceThumbnailMaxPixelSize: 1600
              ] as CFDictionary) else { return nil }
        return ScreenSnapshot.pngData(for: UIImage(cgImage: thumbnail))
    }
}

/// Single optional image shared by feedback entry points. The picker handles
/// conversion and errors; the consumer owns the multipart submit operation.
struct MPImageAttachmentPicker: View {
    @Binding var data: Data?
    @State private var isPresented = false
    @State private var selectedItem: PhotosPickerItem?
    @State private var isLoading = false
    @State private var error: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            if let data {
                MPAttachmentThumb(name: "Ảnh đính kèm", data: data) {
                    self.data = nil
                    selectedItem = nil
                    error = nil
                }
            }
            MPButton(data == nil ? "Đính kèm ảnh" : "Thay ảnh", style: .secondary, isLoading: isLoading) {
                isPresented = true
            }
            .accessibilityIdentifier("feedback-attachment-add")
            .photosPicker(isPresented: $isPresented, selection: $selectedItem, matching: .images)
            if let error { MPStatusMessage(text: error, tone: .danger) }
        }
        .onChange(of: selectedItem) { _, item in
            guard let item else { return }
            Task { @MainActor in
                isLoading = true
                error = nil
                defer { isLoading = false }
                do {
                    guard let source = try await item.loadTransferable(type: Data.self),
                          let png = MPImageAttachmentPolicy.normalizedPNG(source) else {
                        error = "Không thể đọc ảnh. Hãy chọn ảnh khác (tệp tối đa 25 MiB)."
                        return
                    }
                    data = png
                } catch { self.error = "Không thể tải ảnh đã chọn. Vui lòng thử lại." }
            }
        }
    }
}

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
                    .mpAvoidFloatingFeedback()
            } else {
                Image(systemName: "doc.fill")
                    .foregroundStyle(MPColor.action)
                    .frame(width: 44, height: 44)
                    .mpAvoidFloatingFeedback()
            }
            Text(name)
                .font(MPFont.bodySmall)
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

struct MPPickedImage: Equatable, Sendable {
    let name: String
    let data: Data
    let mimeType: String
}

struct MPPhotoImportBatch: Sendable {
    let images: [MPPickedImage]
    let errors: [String]
}

@MainActor
enum MPDocumentImagePolicy {
    static let maxSourceBytes = 25 << 20
    static let maxFileBytes = 5 << 20

    /// File-data boundary shared with system Photos; no provider/upload request.
    static func loadBatch(count: Int, capacity: Int,
                          load: @MainActor (Int) async throws -> Data?) async -> MPPhotoImportBatch {
        var images: [MPPickedImage] = []
        var errors: [String] = []
        for index in 0..<min(max(0, count), max(0, capacity)) {
            do {
                guard let data = try await load(index),
                      let image = normalized(data, name: "IMG_\(UUID().uuidString)") else {
                    errors.append("Không thể đọc ảnh \(index + 1) hoặc chuyển ảnh về tối đa 5 MiB. Hãy chọn ảnh khác.")
                    continue
                }
                images.append(image)
            } catch { errors.append("Không tải được ảnh \(index + 1). Hãy chọn lại khi có kết nối.") }
        }
        return MPPhotoImportBatch(images: images, errors: errors)
    }

    static func normalized(_ data: Data, name: String) -> MPPickedImage? {
        guard !data.isEmpty, data.count <= maxSourceBytes,
              let source = CGImageSourceCreateWithData(data as CFData, nil),
              let type = CGImageSourceGetType(source) as String?,
              CGImageSourceGetCount(source) > 0 else { return nil }
        let base = URL(fileURLWithPath: name).deletingPathExtension().lastPathComponent
        if type == UTType.jpeg.identifier || type == UTType.png.identifier {
            guard data.count <= maxFileBytes,
                  CGImageSourceCreateThumbnailAtIndex(source, 0, [
                    kCGImageSourceCreateThumbnailFromImageAlways: true,
                    kCGImageSourceThumbnailMaxPixelSize: 1
                  ] as CFDictionary) != nil else { return nil }
            let jpeg = type == UTType.jpeg.identifier
            return MPPickedImage(name: base + (jpeg ? ".jpg" : ".png"), data: data,
                                 mimeType: jpeg ? "image/jpeg" : "image/png")
        }
        guard let image = CGImageSourceCreateThumbnailAtIndex(source, 0, [
            kCGImageSourceCreateThumbnailFromImageAlways: true,
            kCGImageSourceCreateThumbnailWithTransform: true,
            kCGImageSourceThumbnailMaxPixelSize: 4096
        ] as CFDictionary) else { return nil }
        let bitmap = UIImage(cgImage: image)
        for quality in [0.9, 0.75, 0.55] {
            if let jpeg = bitmap.jpegData(compressionQuality: quality), jpeg.count <= maxFileBytes {
                return MPPickedImage(name: base + ".jpg", data: jpeg, mimeType: "image/jpeg")
            }
        }
        return nil
    }
}

struct MPDocumentPicker: View {
    let title: String
    let count: Int
    let maxCount: Int
    @Binding var isPresented: Bool
    let onPhotoImport: (([MPPickedImage]) -> Void)?
    let onImport: ([URL]) -> Void
    @State private var isLoading = false

    init(title: String, count: Int, maxCount: Int, isPresented: Binding<Bool>,
         onPhotoImport: (([MPPickedImage]) -> Void)? = nil,
         onImport: @escaping ([URL]) -> Void) {
        self.title = title; self.count = count; self.maxCount = maxCount
        self._isPresented = isPresented
        self.onPhotoImport = onPhotoImport; self.onImport = onImport
    }

    var body: some View {
        if let onPhotoImport {
            MPButton(title, style: .secondary) { isPresented = true }
                .disabled(count >= maxCount || isLoading)
                .accessibilityHint("JPEG, PNG hoặc PDF; tối đa 5 MiB mỗi tệp")
                .mpFormBottomSheet(isPresented: $isPresented, title: "Đính kèm chứng từ",
                    accessibilityIdentifier: "attachment-source-picker",
                    isDismissDisabled: isLoading, onClose: { if !isLoading { isPresented = false } }) {
                    MPAttachmentSourceChooser(count: count, maxCount: maxCount,
                        isLoading: $isLoading, onImport: onImport,
                        onPhotoImport: onPhotoImport, onFinished: { isPresented = false })
                } footer: { Color.clear.frame(height: 0) }
        } else {
            filesOnlyPicker
        }
    }

    private var filesOnlyPicker: some View {
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

private struct MPAttachmentSourceChooser: View {
    let count: Int
    let maxCount: Int
    @Binding var isLoading: Bool
    let onImport: ([URL]) -> Void
    let onPhotoImport: ([MPPickedImage]) -> Void
    let onFinished: () -> Void
    @State private var source = "photos"
    @State private var photosPresented = false
    @State private var filesPresented = false
    @State private var selectedPhotos: [PhotosPickerItem] = []
    @State private var errors: [String] = []

    private var remaining: Int { max(0, maxCount - count) }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            MPSegmentedControl(options: [(id: "photos", title: "Ảnh"), (id: "files", title: "Tệp")], selection: $source)
            MPText("Còn \(remaining) tệp", style: .secondary)
                .accessibilityIdentifier("attachment-remaining-capacity")
            MPText("JPEG, PNG hoặc PDF; tối đa 5 MiB mỗi tệp.", style: .secondary)
            if source == "photos" {
                MPButton("Chọn ảnh", style: .secondary, isLoading: isLoading) { photosPresented = true }
                    .accessibilityIdentifier("attachment-open-photos")
                    .photosPicker(isPresented: $photosPresented, selection: $selectedPhotos,
                                  maxSelectionCount: max(1, remaining), matching: .images,
                                  preferredItemEncoding: .current)
            } else {
                MPButton("Chọn tệp", style: .secondary) { filesPresented = true }
                    .accessibilityIdentifier("attachment-open-files")
                    .fileImporter(isPresented: $filesPresented, allowedContentTypes: [.image, .pdf], allowsMultipleSelection: true, onCompletion: importResult)
            }
            ForEach(Array(errors.enumerated()), id: \.offset) { _, error in
                MPStatusMessage(text: error, tone: .danger)
            }
        }
        .disabled(isLoading || remaining == 0)
        .onChange(of: selectedPhotos) { _, items in
            guard !items.isEmpty, !isLoading else { return }
            let capacity = remaining
            isLoading = true
            errors = []
            Task { @MainActor in
                defer { selectedPhotos = []; isLoading = false }
                let batch = await MPDocumentImagePolicy.loadBatch(count: items.count, capacity: capacity) { index in
                    try await items[index].loadTransferable(type: Data.self)
                }
                errors = batch.errors
                if !batch.images.isEmpty { onPhotoImport(batch.images) }
                if errors.isEmpty { onFinished() }
            }
        }
    }

    private func importResult(_ result: Result<[URL], Error>) {
        switch result {
        case .success(let urls):
            onImport(Array(urls.prefix(remaining)))
            if urls.count > remaining { errors = ["Chỉ có thể thêm \(remaining) tệp."] }
            else { onFinished() }
        case .failure(let error):
            let cocoa = error as NSError
            if cocoa.domain != NSCocoaErrorDomain || cocoa.code != NSUserCancelledError {
                errors = ["Không thể mở tệp. Hãy thử chọn lại."]
            }
        }
    }
}
