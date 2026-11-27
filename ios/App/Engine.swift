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

    private(set) var status = Status.opening
    private var handle: CoreHandle?

    /// Opens the wiki bundled in the app, keeping progress in Application Support.
    func open() async {
        guard let wiki = Bundle.main.url(forResource: "wiki", withExtension: nil) else {
            status = .failed("The app has no wiki in it.")
            return
        }
        let progress = URL.applicationSupportDirectory
        do {
            handle = try await Self.open(wiki: wiki, progress: progress)
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
