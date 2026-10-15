import Foundation
import SQLite3

/// Only accessed synchronously by its owning OutboxStore actor.
final class OutboxDatabase {
    struct Row {
        let id: String
        let intentHash: String
        let sealed: Data
    }

    private var connection: OpaquePointer?
    private let transient = unsafeBitCast(-1, to: sqlite3_destructor_type.self)

    init(url: URL) throws {
        guard sqlite3_open_v2(url.path, &connection, SQLITE_OPEN_CREATE | SQLITE_OPEN_READWRITE | SQLITE_OPEN_FULLMUTEX, nil) == SQLITE_OK else {
            if let connection { sqlite3_close_v2(connection) }
            connection = nil
            throw OutboxStoreError.storageUnavailable
        }
        do {
            sqlite3_busy_timeout(connection, 5000)
            let version = try userVersion()
            guard version == 0 || version == 1 else { throw OutboxStoreError.unsupportedVersion }
            try execute("PRAGMA journal_mode=WAL")
            try execute("PRAGMA synchronous=FULL")
            if version == 0 {
                try transaction {
                    try execute("CREATE TABLE IF NOT EXISTS outbox (ordinal INTEGER PRIMARY KEY AUTOINCREMENT, id TEXT NOT NULL UNIQUE, namespace TEXT NOT NULL, intent_hash TEXT NOT NULL, sealed BLOB NOT NULL, UNIQUE(namespace,intent_hash))")
                    try execute("PRAGMA user_version=1")
                }
            }
        } catch {
            sqlite3_close_v2(connection)
            connection = nil
            throw error
        }
    }

    deinit { sqlite3_close_v2(connection) }

    func transaction<T>(_ body: () throws -> T) throws -> T {
        try execute("BEGIN IMMEDIATE")
        do {
            let result = try body()
            try execute("COMMIT")
            return result
        } catch {
            try? execute("ROLLBACK")
            throw error
        }
    }

    func rows(namespace: String) throws -> [Row] {
        let statement = try prepare("SELECT id,intent_hash,sealed FROM outbox WHERE namespace=? ORDER BY ordinal")
        defer { sqlite3_finalize(statement) }
        try bind(namespace, to: 1, in: statement)
        var result: [Row] = []
        while true {
            let status = sqlite3_step(statement)
            if status == SQLITE_DONE { return result }
            guard status == SQLITE_ROW,
                  let id = sqlite3_column_text(statement, 0),
                  let hash = sqlite3_column_text(statement, 1),
                  let bytes = sqlite3_column_blob(statement, 2),
                  sqlite3_column_bytes(statement, 2) > 0 else { throw OutboxStoreError.corruptStore }
            result.append(Row(id: String(cString: id), intentHash: String(cString: hash), sealed: Data(bytes: bytes, count: Int(sqlite3_column_bytes(statement, 2)))))
        }
    }

    func insert(id: String, namespace: String, intentHash: String, sealed: Data) throws {
        let statement = try prepare("INSERT INTO outbox(id,namespace,intent_hash,sealed) VALUES(?,?,?,?)")
        defer { sqlite3_finalize(statement) }
        try bind(id, to: 1, in: statement)
        try bind(namespace, to: 2, in: statement)
        try bind(intentHash, to: 3, in: statement)
        let bound = sealed.withUnsafeBytes { sqlite3_bind_blob(statement, 4, $0.baseAddress, Int32($0.count), transient) }
        guard bound == SQLITE_OK, sqlite3_step(statement) == SQLITE_DONE else { throw OutboxStoreError.storageUnavailable }
    }

    func update(id: String, namespace: String, sealed: Data) throws {
        let statement = try prepare("UPDATE outbox SET sealed=? WHERE id=? AND namespace=?")
        defer { sqlite3_finalize(statement) }
        let bound = sealed.withUnsafeBytes { sqlite3_bind_blob(statement, 1, $0.baseAddress, Int32($0.count), transient) }
        guard bound == SQLITE_OK else { throw OutboxStoreError.storageUnavailable }
        try bind(id, to: 2, in: statement)
        try bind(namespace, to: 3, in: statement)
        guard sqlite3_step(statement) == SQLITE_DONE, sqlite3_changes(connection) == 1 else { throw OutboxStoreError.storageUnavailable }
    }

    private func bind(_ value: String, to index: Int32, in statement: OpaquePointer) throws {
        guard sqlite3_bind_text(statement, index, value, -1, transient) == SQLITE_OK else { throw OutboxStoreError.storageUnavailable }
    }

    private func prepare(_ sql: String) throws -> OpaquePointer {
        var statement: OpaquePointer?
        guard sqlite3_prepare_v2(connection, sql, -1, &statement, nil) == SQLITE_OK, let statement else {
            if let statement { sqlite3_finalize(statement) }
            throw OutboxStoreError.storageUnavailable
        }
        return statement
    }

    private func execute(_ sql: String) throws {
        guard sqlite3_exec(connection, sql, nil, nil, nil) == SQLITE_OK else { throw OutboxStoreError.storageUnavailable }
    }

    private func userVersion() throws -> Int32 {
        let statement = try prepare("PRAGMA user_version")
        defer { sqlite3_finalize(statement) }
        guard sqlite3_step(statement) == SQLITE_ROW else { throw OutboxStoreError.corruptStore }
        return sqlite3_column_int(statement, 0)
    }
}
