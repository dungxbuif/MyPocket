import Foundation

enum AssistantScrollPolicy {
    static let followThreshold: CGFloat = 64

    static func shouldFollowLatest(distanceFromBottom: CGFloat, isLoadingOlder: Bool) -> Bool {
        !isLoadingOlder && distanceFromBottom <= followThreshold
    }
}
