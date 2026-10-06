import UIKit

@MainActor
enum ScreenSnapshot {
    static let maxPNGBytes = 1 << 20

    static func pngData(for image: UIImage, maxBytes: Int = maxPNGBytes) -> Data? {
        guard maxBytes > 0 else { return nil }
        var current = image
        for _ in 0..<8 {
            guard let data = current.pngData() else { return nil }
            if data.count <= maxBytes { return data }

            let ratio = sqrt(Double(maxBytes) / Double(data.count)) * 0.9
            let scale = max(0.25, min(0.75, ratio))
            let nextSize = CGSize(
                width: max(160, floor(current.size.width * scale)),
                height: max(160, floor(current.size.height * scale))
            )
            guard nextSize.width < current.size.width || nextSize.height < current.size.height else { return nil }
            let format = UIGraphicsImageRendererFormat()
            format.scale = 1
            format.opaque = true
            current = UIGraphicsImageRenderer(size: nextSize, format: format).image { _ in
                current.draw(in: CGRect(origin: .zero, size: nextSize))
            }
        }

        guard let data = current.pngData(), data.count <= maxBytes else { return nil }
        return data
    }

    static func capture() -> Data? {
        guard let window = UIApplication.shared.connectedScenes
            .compactMap({ $0 as? UIWindowScene })
            .flatMap(\.windows)
            .first(where: { $0.isKeyWindow }) else { return nil }
        let renderer = UIGraphicsImageRenderer(bounds: window.bounds)
        let image = renderer.image { _ in
            window.drawHierarchy(in: window.bounds, afterScreenUpdates: false)
        }
        return pngData(for: image)
    }
}
