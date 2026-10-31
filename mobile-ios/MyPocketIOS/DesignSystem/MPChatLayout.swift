import SwiftUI

enum MPChatFooterLayout {
    /// Reserved breathing room for the floating feedback control on every
    /// screen that hosts chat. Assistant is the deliberate exception because
    /// the shell hides feedback while the full-screen chat is visible.
    static let feedbackClearance: CGFloat = 60
}
