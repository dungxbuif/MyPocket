import Foundation

#if DEBUG
/// Process-local session storage for unsigned simulator UI tests.
/// Production builds always use `KeychainSessionStore`.
final class InMemorySessionStore: SessionStore, @unchecked Sendable {
    private let lock = NSLock()
    private var session: AccessSession?

    func load() -> AccessSession? {
        lock.lock()
        defer { lock.unlock() }
        return session
    }

    func save(_ session: AccessSession) throws {
        lock.lock()
        defer { lock.unlock() }
        self.session = session
    }

    func clear() {
        lock.lock()
        defer { lock.unlock() }
        session = nil
    }
}
#endif
