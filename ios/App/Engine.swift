import Foundation
import TuiLaoShiCore

/// Camp's numbers, as the Go engine's Camp() reports them.
struct Camp: Decodable, Equatable {
    struct Stage: Decodable, Equatable {
        let number: Int
        let name: String
        let english: String
        let location: String
        let region: String
        let demon: String
        let demonEnglish: String
    }

    let due: Int
    let streak: Int
    let tongqian: Int
    let yuanbao: Int
    let rank: String
    let rankEnglish: String
    let level: Int
    let learned: Int
    let mastered: Int
    let stage: Stage?
    let party: [String]
    let term: String
    let termEnglish: String
}

/// The Go engine's handle. Go guards its own state, so the handle may be
/// used from any thread.
final class CoreHandle: @unchecked Sendable {
    let core: MobileCore
    init(_ core: MobileCore) { self.core = core }
}

/// The Go engine with the bundled wiki and the app's own progress. All
/// calls into Go run off the main thread.
@MainActor @Observable
final class Engine {
    enum Status: Equatable {
        case opening
        case ready(Camp)
        case failed(String)
    }

    /// How progress last synced with the Mac through iCloud Drive.
    enum SyncState: Equatable {
        case never
        case off(String)  // iCloud Drive isn't available, and why
        case synced(Date)
        case failed(String)
    }

    private(set) var status = Status.opening
    private(set) var sync = SyncState.never
    private var handle: CoreHandle?

    /// The iCloud container both the phone and the Mac use. On the Mac it
    /// is ~/Library/Mobile Documents/iCloud~com~therealsiege~tuilaoshi/Documents,
    /// the folder to give TUILAOSHI_SYNC_DIR.
    nonisolated static let container = "iCloud.com.therealsiege.tuilaoshi"

    /// This device's folder name in the shared one.
    nonisolated static let syncName = "iphone"

    /// Opens the wiki bundled in the app, keeping progress in Application Support.
    func open() async {
        guard let wiki = Bundle.main.url(forResource: "wiki", withExtension: nil) else {
            status = .failed("The app has no wiki in it.")
            return
        }
        let progress = URL.applicationSupportDirectory
        // UI tests start from a new pilgrim's progress.
        if ProcessInfo.processInfo.arguments.contains("-fresh-progress") {
            try? FileManager.default.removeItem(at: progress)
        }
        do {
            handle = try await Self.open(wiki: wiki, progress: progress)
            await syncNow()
            await refresh()
        } catch {
            status = .failed("Couldn't open the journey: \(error.localizedDescription)")
        }
    }

    /// Reloads camp's numbers.
    func refresh() async {
        guard let handle else { return }
        do {
            status = .ready(try await Self.camp(handle))
        } catch {
            status = .failed("Couldn't read camp: \(error.localizedDescription)")
        }
    }

    /// Merges the Mac's progress in and writes the phone's out, through
    /// the iCloud Drive folder they share.
    func syncNow() async {
        guard let handle else { return }
        guard let root = await Self.syncRoot() else {
            sync = .off("iCloud Drive is off or not signed in, so progress stays on this iPhone.")
            return
        }
        do {
            try await Self.sync(handle, root: root)
            sync = .synced(.now)
        } catch {
            sync = .failed("Couldn't sync: \(error.localizedDescription)")
        }
    }

    /// Up to limit cards due now.
    func dueCards(limit: Int) async throws -> [Card] {
        guard let handle else { throw CocoaError(.fileNoSuchFile) }
        return try await Self.dueCards(handle, limit: limit)
    }

    /// Saves an answer: rating 0 again, 1 hard, 2 good, 3 easy.
    func review(_ id: String, rating: Int) async throws {
        guard let handle else { throw CocoaError(.fileNoSuchFile) }
        try await Self.review(handle, id: id, rating: rating)
    }

    /// Records a finished study session for the streak and the day's stats.
    func finishSession(kind: String, started: Date, answers: [(id: String, correct: Bool, rating: Int)]) async throws {
        guard let handle else { throw CocoaError(.fileNoSuchFile) }
        let list = answers.map { ["id": $0.id, "correct": $0.correct, "rating": $0.rating] as [String: Any] }
        let json = String(decoding: try JSONSerialization.data(withJSONObject: list), as: UTF8.self)
        try await Self.finish(handle, kind: kind, started: Int64(started.timeIntervalSince1970), answers: json)
    }

    nonisolated static func dueCards(_ handle: CoreHandle, limit: Int) async throws -> [Card] {
        var error: NSError?
        let json = handle.core.dueCards(limit, error: &error)
        if let error { throw error }
        return try JSONDecoder().decode([Card].self, from: Data(json.utf8))
    }

    nonisolated static func review(_ handle: CoreHandle, id: String, rating: Int) async throws {
        var error: NSError?
        _ = handle.core.review(id, rating: rating, error: &error)
        if let error { throw error }
    }

    nonisolated static func finish(_ handle: CoreHandle, kind: String, started: Int64, answers: String) async throws {
        try handle.core.finishSession(kind, started: started, answersJSON: answers)
    }

    /// Writes the progress as readable files into the app's Documents
    /// folder, which the Files app shows, and returns the folder.
    func export() async throws -> URL {
        guard let handle else { throw CocoaError(.fileNoSuchFile) }
        let stamp = Date.now.formatted(.iso8601.year().month().day())
        let dir = URL.documentsDirectory.appending(path: "Export \(stamp)")
        try await Self.export(handle, to: dir)
        return dir
    }

    /// The Documents folder of the shared iCloud container, or nil when
    /// iCloud Drive isn't available. Looking it up can block, so it runs
    /// off the main thread.
    nonisolated private static func syncRoot() async -> URL? {
        FileManager.default.url(forUbiquityContainerIdentifier: container)?.appending(path: "Documents")
    }

    nonisolated static func sync(_ handle: CoreHandle, root: URL) async throws {
        try FileManager.default.createDirectory(at: root, withIntermediateDirectories: true)
        try handle.core.sync(root.path, name: syncName)
    }

    nonisolated static func export(_ handle: CoreHandle, to dir: URL) async throws {
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        try handle.core.export(dir.path)
    }

    nonisolated private static func open(wiki: URL, progress: URL) async throws -> CoreHandle {
        try FileManager.default.createDirectory(at: progress, withIntermediateDirectories: true)
        var error: NSError?
        guard let core = MobileOpen(wiki.path, progress.path, &error) else {
            throw error ?? CocoaError(.fileReadUnknown)
        }
        return CoreHandle(core)
    }

    nonisolated static func camp(_ handle: CoreHandle) async throws -> Camp {
        var error: NSError?
        let json = handle.core.camp(&error)
        if let error { throw error }
        return try JSONDecoder().decode(Camp.self, from: Data(json.utf8))
    }
}
