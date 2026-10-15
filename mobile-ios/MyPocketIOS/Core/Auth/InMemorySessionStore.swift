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

    func replace(_ session: AccessSession, ifCredentialsMatch original: AccessSession) throws -> Bool {
        lock.lock()
        defer { lock.unlock() }
        guard self.session?.hasSameCredentials(as: original) == true, session.user.id == original.user.id else { return false }
        self.session = session
        return true
    }

    func clear(ifCredentialsMatch original: AccessSession) -> Bool {
        lock.lock()
        defer { lock.unlock() }
        guard session?.hasSameCredentials(as: original) == true else { return false }
        session = nil
        return true
    }
}
#endif
