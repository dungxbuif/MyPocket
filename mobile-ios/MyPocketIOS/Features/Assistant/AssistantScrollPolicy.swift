import Foundation

struct AssistantScrollAnchor: Equatable {
    let messageID: String
    let offsetFromViewportTop: CGFloat
}

enum AssistantScrollPolicy {
    static let followThreshold: CGFloat = 64

    static func shouldFollowLatest(distanceFromBottom: CGFloat, isLoadingOlder: Bool) -> Bool {
        !isLoadingOlder && distanceFromBottom <= followThreshold
    }

    static func distanceFromBottom(contentFrame: CGRect, viewportHeight: CGFloat) -> CGFloat {
        guard viewportHeight > 0 else { return 0 }
        return max(0, contentFrame.maxY - viewportHeight)
    }

    static func visibleAnchor(messageFrames: [String: CGRect], viewportHeight: CGFloat) -> AssistantScrollAnchor? {
        guard viewportHeight > 0 else { return nil }
        let visible = messageFrames.compactMap { id, frame -> AssistantScrollAnchor? in
            guard frame.maxY > 0, frame.minY < viewportHeight, frame.height > 0 else { return nil }
            return AssistantScrollAnchor(messageID: id, offsetFromViewportTop: frame.minY)
        }
        return visible.min { lhs, rhs in
            abs(lhs.offsetFromViewportTop) < abs(rhs.offsetFromViewportTop)
        }
    }

    static func restoreSpacerHeight(for anchor: AssistantScrollAnchor) -> CGFloat {
        max(0, anchor.offsetFromViewportTop)
    }
}
