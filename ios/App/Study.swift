import SwiftUI

/// A due card, as the Go engine's DueCards() reports it.
struct Card: Decodable, Identifiable, Hashable, Sendable {
    let id: String
    let chinese: String
    let pinyin: String
    let english: String
    let examples: [String]?
    let cloze: String?
    let sentence: String?
}

/// How a session asks about each card.
enum StudyMode: String, CaseIterable, Identifiable, Hashable {
    case review, recognition, production, listening, sentence

    var id: Self { self }

    var title: String {
        switch self {
        case .review: "复习 Review"
        case .recognition: "认读 Recognition"
        case .production: "写作 Production"
        case .listening: "听力 Listening"
        case .sentence: "造句 Sentence"
        }
    }

    var subtitle: String {
        switch self {
        case .review: "See the card, recall it, rate yourself"
        case .recognition: "See Chinese, recall the meaning"
        case .production: "See English, type the Chinese"
        case .listening: "Hear Chinese, recall what was said"
        case .sentence: "Fill the missing word in a sentence"
        }
    }

    /// The session kind the engine records: reviews or drills.
    var kind: String { self == .review ? "review" : "drill" }
}

/// One study session: the due cards, one at a time. Each answer is saved
/// as it is given, so leaving early keeps every answer so far; the session
/// itself is written when it ends or is left.
@MainActor @Observable
final class StudySession {
    let mode: StudyMode
    private(set) var cards: [Card] = []
    private(set) var index = 0
    private(set) var answers: [(id: String, correct: Bool, rating: Int)] = []
    private(set) var loading = true
    private(set) var error: String?
    var revealed = false
    var typed = ""

    private let engine: Engine
    private let started = Date.now
    private var finished = false

    /// Cards per session, so a big backlog comes in sittings (the
    /// learner's "words a sitting" setting).
    static var size = 20

    private let preset: [Card]?

    /// A session over the due cards, or over the given cards (a walk of
    /// the spider's web).
    init(mode: StudyMode, engine: Engine, cards: [Card]? = nil) {
        self.mode = mode
        self.engine = engine
        self.preset = cards
    }

    var card: Card? { index < cards.count ? cards[index] : nil }
    var done: Bool { !loading && card == nil }

    /// For production: the typed answer is exactly the word.
    var typedRight: Bool { typed.trimmingCharacters(in: .whitespaces) == card?.chinese }

    func load() async {
        if let preset {
            cards = preset
            loading = false
            return
        }
        do {
            let all = try await engine.dueCards(limit: Self.size * 3)
            // Sentence mode needs an example sentence holding the word.
            let usable = mode == .sentence ? all.filter { $0.cloze != nil } : all
            cards = Array(usable.prefix(Self.size))
        } catch {
            self.error = "Couldn't load cards: \(error.localizedDescription)"
        }
        loading = false
    }

    /// Records a rating (0 again, 1 hard, 2 good, 3 easy) and moves on.
    func rate(_ rating: Int) async {
        guard let card else { return }
        do {
            try await engine.review(card.id, rating: rating)
            (rating > 0 ? Cue.right : Cue.wrong).play()
            answers.append((card.id, rating > 0, rating))
            index += 1
            revealed = false
            typed = ""
            if card == nil { await finish() }
        } catch {
            self.error = "Couldn't save that answer: \(error.localizedDescription)"
        }
    }

    /// Writes the session, once, for the streak and the day's stats.
    func finish() async {
        guard !finished, !answers.isEmpty else { return }
        finished = true
        try? await engine.finishSession(kind: mode.kind, started: started, answers: answers)
        await engine.refresh()
    }
}

/// Picks how to study, then runs a session.
struct StudyView: View {
    @Environment(Engine.self) private var engine

    var body: some View {
        List(StudyMode.allCases) { mode in
            NavigationLink(value: mode) {
                VStack(alignment: .leading, spacing: 2) {
                    Text(mode.title).font(.headline)
                    Text(mode.subtitle).font(.subheadline).foregroundStyle(.secondary)
                }
            }
            .accessibilityIdentifier("mode.\(mode.rawValue)")
        }
        .navigationTitle("复习 Study")
        .navigationDestination(for: StudyMode.self) { SessionView(session: StudySession(mode: $0, engine: engine)) }
    }
}
