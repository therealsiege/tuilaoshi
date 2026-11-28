import Foundation
import TuiLaoShiCore

/// Where Claude is reached: the game's server, signed in with this
/// device's session. No API key is ever on the phone.
struct ClaudeSession: Sendable, Equatable {
    let server: String
    let token: String
}

/// What a diary entry came to, as the engine's WriteDiary() reports it.
struct DiaryGrade: Decodable, Equatable, Sendable {
    let date: String
    let day: Int
    let text: String
    let score: Int
    let feedback: String
    let kept: Bool
}

/// The storyteller's shelf, as the engine's Storyteller() reports it.
struct StoryShelf: Decodable, Equatable, Sendable {
    struct Chapter: Decodable, Equatable, Hashable, Sendable {
        let id: String
        let number: Int
        let title: String
        let date: String
    }
    let next: Int
    let canTell: Bool
    let done: Bool
    let chapters: [Chapter]
}

/// Claude's view of an answer that wasn't exact, as GradeAnswer() reports it.
struct ClaudeGrade: Decodable, Equatable, Sendable {
    let correct: Bool
    let score: Int
    let feedback: String
    /// The rating the grade suggests: 0 again, 1 hard, 2 good, 3 easy.
    let rating: Int
}

/// Why Claude couldn't help, in words for the learner.
enum ClaudeTrouble {
    /// Signed out: what signing in would bring.
    static func signedOut(_ what: String) -> String {
        Account.server.isEmpty
            ? "\(what) comes with the game's server, which isn't set up yet."
            : "\(what): sign in (Settings), and Claude reaches you through the game's server."
    }

    /// An error from a Claude call, explained: the server's daily cap, a
    /// sign-in the server no longer takes, or anything else.
    static func explain(_ error: any Error) -> String { explain(message: error.localizedDescription) }

    static func explain(message: String) -> String {
        if message.contains("Claude limit") || message.contains("daily Claude limit") {
            return "Today's Claude calls on the game's server are used up. They come back at midnight UTC."
        }
        if message.contains("sign in again") {
            return "The game's server didn't accept this iPhone's sign-in. Sign out and in again in Settings."
        }
        if message.contains("sign in (Settings) first") {
            return "Sign in (Settings) first: Claude comes through the game's server."
        }
        return "Couldn't reach Claude: \(message)"
    }
}

extension Account {
    /// The server and session for Claude, or nil when signed out.
    var claude: ClaudeSession? {
        guard !Self.server.isEmpty, let token else { return nil }
        return ClaudeSession(server: Self.server, token: token)
    }
}

/// The Claude calls, through the game's server. Each runs off the main
/// thread; the static forms take the engine's handle, for tests.
extension Engine {
    func writeDiary(_ text: String, via s: ClaudeSession) async throws -> DiaryGrade {
        try await call { try Self.writeDiary($0, via: s, text: text) }
    }

    func claudeGazette(via s: ClaudeSession) async throws -> Writing {
        try await call { try Self.claudeGazette($0, via: s) }
    }

    func storyteller() async throws -> StoryShelf {
        try await decode(StoryShelf.self) { h, e in h.core.storyteller(e) }
    }

    func tellChapter(via s: ClaudeSession) async throws -> StoryShelf.Chapter {
        try await call { try Self.tellChapter($0, via: s) }
    }

    func gradeAnswer(card: String, answer: String, mode: StudyMode, via s: ClaudeSession) async throws -> ClaudeGrade {
        try await call { try Self.gradeAnswer($0, via: s, card: card, answer: answer, mode: mode.rawValue) }
    }

    nonisolated static func writeDiary(_ h: CoreHandle, via s: ClaudeSession, text: String) throws -> DiaryGrade {
        var e: NSError?
        let json = h.core.writeDiary(s.server, token: s.token, text: text, error: &e)
        return try decoded(json, e)
    }

    nonisolated static func claudeGazette(_ h: CoreHandle, via s: ClaudeSession) throws -> Writing {
        var e: NSError?
        let json = h.core.claudeGazette(s.server, token: s.token, error: &e)
        return try decoded(json, e)
    }

    nonisolated static func tellChapter(_ h: CoreHandle, via s: ClaudeSession) throws -> StoryShelf.Chapter {
        var e: NSError?
        let json = h.core.tellChapter(s.server, token: s.token, error: &e)
        return try decoded(json, e)
    }

    nonisolated static func gradeAnswer(_ h: CoreHandle, via s: ClaudeSession, card: String, answer: String, mode: String) throws -> ClaudeGrade {
        var e: NSError?
        let json = h.core.gradeAnswer(s.server, token: s.token, cardID: card, answer: answer, mode: mode, error: &e)
        return try decoded(json, e)
    }

    nonisolated private static func decoded<T: Decodable>(_ json: String, _ error: NSError?) throws -> T {
        if let error { throw error }
        return try JSONDecoder().decode(T.self, from: Data(json.utf8))
    }
}
