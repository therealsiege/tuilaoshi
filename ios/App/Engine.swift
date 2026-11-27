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

/// The learner's settings as the engine's Settings() reports them.
struct AppSettings: Codable, Equatable, Sendable {
    var speech: Bool
    var effects: Bool
    var toneColours: Bool
    var reviewChunk: Int
    var newWords: Int
    var difficulty: String
    var scheduler: String

    /// Makes sound, display and study follow these settings.
    @MainActor func apply() {
        SoundSettings.voices = speech
        SoundSettings.cues = effects
        SoundSettings.toneColours = toneColours
        StudySession.size = reviewChunk
    }
}

/// Camp's Today panel, as the engine's CampToday() reports it.
struct Today: Decodable, Equatable, Sendable {
    struct Bounty: Decodable, Equatable, Sendable { let text: String; let have: Int; let target: Int; let pay: Int; let done: Bool }
    struct Festival: Decodable, Equatable, Sendable { let name: String; let english: String; let on: Bool; let days: Int }
    struct Term: Decodable, Equatable, Sendable { let name: String; let english: String; let text: String; let starts: Bool; let weather: String }
    let reviewed: Int
    let correct: Int
    let dailyGoal: Int
    let bounties: [Bounty]
    let shields: Int
    let festival: Festival?
    let term: Term
    let missed: Int
    let demonWakes: Int
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

    /// How progress last synced with the other devices.
    enum SyncState: Equatable {
        case never
        case off(String)  // sync isn't available, and why
        case synced(Date)
        case failed(String)
    }

    private(set) var status = Status.opening
    private(set) var sync = SyncState.never
    private var handle: CoreHandle?

    /// This device's name among the synced ones.
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
            _ = try? await settings()
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

    /// Merges the other devices' progress in and sends this one's out,
    /// through the game's own server once signed in (#64). Until then
    /// progress stays on this iPhone.
    func syncNow() async {
        guard handle != nil else { return }
        sync = .off("Sync with the Mac comes with signing in, which isn't built yet. Progress is kept on this iPhone.")
    }

    /// The learner's settings, shared with the Mac.
    func settings() async throws -> AppSettings {
        let s: AppSettings = try await call { h in
            var error: NSError?
            let json = h.core.settings(&error)
            if let error { throw error }
            return try JSONDecoder().decode(AppSettings.self, from: Data(json.utf8))
        }
        s.apply()
        return s
    }

    /// Saves the learner's settings and applies them at once.
    func saveSettings(_ s: AppSettings) async throws {
        let json = String(decoding: try JSONEncoder().encode(s), as: UTF8.self)
        try await call { h in try h.core.saveSettings(json) }
        s.apply()
    }

    /// Camp's Today panel.
    func today() async throws -> Today {
        try await call { h in
            var error: NSError?
            let json = h.core.campToday(&error)
            if let error { throw error }
            return try JSONDecoder().decode(Today.self, from: Data(json.utf8))
        }
    }

    /// The readings and journey episodes reached so far.
    func readings() async throws -> [ReadingItem] {
        try await call { h in
            var error: NSError?
            let json = h.core.readings(&error)
            if let error { throw error }
            return try JSONDecoder().decode([ReadingItem].self, from: Data(json.utf8))
        }
    }

    /// One reading, split into glossed words.
    func reading(_ id: String) async throws -> Reading {
        try await call { h in
            var error: NSError?
            let json = h.core.reading(id, error: &error)
            if let error { throw error }
            return try JSONDecoder().decode(Reading.self, from: Data(json.utf8))
        }
    }

    /// Sends a word met in reading back for review; false when it isn't in the decks.
    func markUnknown(_ word: String) async throws -> Bool {
        try await call { h in
            var ok: ObjCBool = false
            try h.core.markUnknown(word, ret0_: &ok)
            return ok.boolValue
        }
    }

    /// Keeps a word with its sentence as a card; false when it already has one.
    func addSentenceCard(_ word: String, pinyin: String, english: String, line: String) async throws -> Bool {
        try await call { h in
            var ok: ObjCBool = false
            try h.core.addSentenceCard(word, pinyin: pinyin, english: english, line: line, ret0_: &ok)
            return ok.boolValue
        }
    }

    /// Runs work against the engine off the main thread.
    private func call<T: Sendable>(_ work: @Sendable @escaping (CoreHandle) throws -> T) async throws -> T {
        guard let handle else { throw CocoaError(.fileNoSuchFile) }
        return try await Task.detached { try work(handle) }.value
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

    /// Merges every other device's exported progress under root into this
    /// one, then writes this device's to root/syncName. The server fetches
    /// and stores the folders; the merge is the engine's.
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
