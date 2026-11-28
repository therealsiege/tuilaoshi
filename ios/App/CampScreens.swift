import Charts
import SwiftUI

/// A letter or a gazette issue: Chinese with its English.
struct Writing: Decodable, Equatable, Sendable {
    let week: String
    let text: String
    let english: String
    let byHand: Bool?
    /// Why Claude didn't write it, when the clerk set it instead.
    let missed: String?
}

struct DiaryEntry: Decodable, Identifiable, Equatable, Sendable {
    let date: String
    let day: Int
    let text: String
    let score: Int
    let feedback: String?
    var id: String { date + text }
}

/// My progress, as the engine's Progress() reports it.
struct ProgressReport: Decodable, Equatable, Sendable {
    struct Day: Decodable, Equatable, Sendable { let date: String; let reviewed: Int; let correct: Int }
    struct Hard: Decodable, Equatable, Sendable { let chinese: String; let pinyin: String; let english: String; let lapses: Int }
    let week: [Day]
    let ahead: [Int]
    let hard: [Hard]
    let streak: Int
    let characters: Int
    let words: Int
    let grammar: Int
    let readings: Int
}

/// The things camp keeps to read: the letter, the gazette, the diary and
/// My progress.
enum Paper: String, CaseIterable, Identifiable, Hashable {
    case letter, gazette, diary, progress, web
    var id: Self { self }

    var title: String {
        switch self {
        case .letter: "观音来信 Guanyin's letter"
        case .gazette: "邸报 The court gazette"
        case .diary: "日记 Travel diary"
        case .progress: "进度 My progress"
        case .web: "盘丝洞 The spider's web"
        }
    }

    var symbol: String {
        switch self {
        case .letter: "envelope"
        case .gazette: "newspaper"
        case .diary: "book.closed"
        case .progress: "chart.bar"
        case .web: "circle.hexagongrid"
        }
    }
}

struct PaperView: View {
    let paper: Paper
    @Environment(Account.self) private var account

    var body: some View {
        switch paper {
        case .letter: WritingView(title: paper.title, load: { try await $0.letter() },
                                  empty: "Guanyin writes each Monday about the week before, once there's been a week with reviews.")
        case .gazette:
            // Signed in, Claude writes the week's issue; else the clerk sets it.
            WritingView(title: paper.title, load: { [claude = account.claude] engine in
                if let claude { return try await engine.claudeGazette(via: claude) }
                return try await engine.gazette()
            }, empty: "")
        case .diary: DiaryView()
        case .progress: ProgressScreen()
        case .web: WebScreen()
        }
    }
}

/// A letter or gazette, with its English beneath and read-aloud.
struct WritingView: View {
    /// Why an issue was set by the clerk: Claude couldn't write it, or the
    /// learner isn't signed in.
    nonisolated static func byHandNote(missed: String?) -> String {
        if let missed {
            return "Set by the court's clerk from a template. \(ClaudeTrouble.explain(message: missed))"
        }
        return "Set by the court's clerk from a template. " + ClaudeTrouble.signedOut("An issue written by Claude")
    }

    let title: String
    let load: @MainActor (Engine) async throws -> Writing?
    let empty: String
    @Environment(Engine.self) private var engine
    @State private var writing: Writing?
    @State private var loaded = false
    @State private var showEnglish = false
    @State private var error: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                if let w = writing {
                    Text("Week of \(w.week)").font(.caption).foregroundStyle(.secondary)
                    Text(w.text).font(.title3).lineSpacing(6).textSelection(.enabled)
                    if showEnglish { Text(w.english).foregroundStyle(.secondary) }
                    if w.byHand == true {
                        Text(Self.byHandNote(missed: w.missed))
                            .font(.caption).foregroundStyle(.secondary)
                    }
                } else if let error {
                    Text(error).foregroundStyle(Theme.cinnabar)
                } else if loaded {
                    ContentUnavailableView("Nothing yet", systemImage: "envelope", description: Text(empty))
                } else {
                    ProgressView()
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding()
        }
        .navigationTitle(title)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            if let w = writing {
                ToolbarItemGroup(placement: .topBarTrailing) {
                    Button { Speaker.chinese(w.text) } label: { Image(systemName: "speaker.wave.2") }
                        .accessibilityLabel("Read aloud")
                    Toggle("English", isOn: $showEnglish).toggleStyle(.button)
                }
            }
        }
        .task {
            do { writing = try await load(engine) } catch { self.error = error.localizedDescription }
            loaded = true
        }
    }
}

/// The travel diary: once a day, two or three sentences in Chinese about
/// the day on the road, graded by Claude as on the Mac. A passing entry is
/// kept and lifts the party's spirits; one that doesn't can be revised.
struct DiaryView: View {
    @Environment(Engine.self) private var engine
    @Environment(Account.self) private var account
    @State private var entries: [DiaryEntry] = []
    @State private var text = ""
    @State private var sending = false
    @State private var grade: DiaryGrade?
    @State private var note: String?

    /// Today's entry is in the diary.
    private var written: Bool { entries.first?.date == DiaryView.today() }

    nonisolated static func today(_ now: Date = .now) -> String {
        let f = DateFormatter()
        f.calendar = Calendar(identifier: .gregorian)
        f.locale = Locale(identifier: "en_US_POSIX")
        f.dateFormat = "yyyy-MM-dd"
        return f.string(from: now)
    }

    var body: some View {
        List {
            Section {
                if let claude = account.claude, !written {
                    TextField("今天走到了哪里？遇到了什么？心情怎么样？", text: $text, axis: .vertical)
                        .lineLimit(3...8)
                        .font(.title3)
                        .disabled(sending)
                        .accessibilityIdentifier("diary.text")
                    Button {
                        Task { await send(via: claude) }
                    } label: {
                        if sending { ProgressView() } else { Text("Give it to Claude to read") }
                    }
                    .disabled(sending || text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                    if let grade, !grade.kept {
                        Text("Not quite yet: \(grade.feedback) Revise it and send it again.").font(.footnote)
                    }
                } else if written {
                    if let grade, grade.kept {
                        Label("Kept. The party's spirits lift (+\(grade.score == 2 ? 10 : 5) morale).", systemImage: "checkmark.seal")
                            .foregroundStyle(Theme.jade)
                        if !grade.feedback.isEmpty { Text(grade.feedback).font(.footnote) }
                    } else {
                        Text("Today's entry is written. Write the next one tomorrow.").font(.footnote).foregroundStyle(.secondary)
                    }
                } else {
                    Text("Writing an entry needs Claude to grade it. " + ClaudeTrouble.signedOut("Writing here") + " Entries written on the Mac show here once progress syncs.")
                        .font(.footnote).foregroundStyle(.secondary)
                }
                if let note { Text(note).font(.footnote).foregroundStyle(Theme.cinnabar) }
            }
            ForEach(entries) { e in
                VStack(alignment: .leading, spacing: 4) {
                    Text("第\(e.day)天 · \(e.date)").font(.caption).foregroundStyle(.secondary)
                    Text(e.text)
                    if let f = e.feedback, !f.isEmpty { Text(f).font(.footnote).foregroundStyle(.secondary) }
                }
            }
        }
        .navigationTitle("日记 Travel diary")
        .task { entries = (try? await engine.diary()) ?? [] }
    }

    private func send(via claude: ClaudeSession) async {
        sending = true
        note = nil
        defer { sending = false }
        do {
            let g = try await engine.writeDiary(text, via: claude)
            grade = g
            if g.kept {
                text = ""
                entries = (try? await engine.diary()) ?? entries
            }
        } catch {
            note = ClaudeTrouble.explain(error)
        }
    }
}

struct ProgressScreen: View {
    @Environment(Engine.self) private var engine
    @State private var report: ProgressReport?

    var body: some View {
        List {
            if let r = report {
                Section("The last seven days") {
                    Chart(r.week, id: \.date) { d in
                        BarMark(x: .value("Day", String(d.date.suffix(5))), y: .value("Reviewed", d.reviewed))
                            .foregroundStyle(Theme.jade)
                        BarMark(x: .value("Day", String(d.date.suffix(5))), y: .value("Right", d.correct))
                            .foregroundStyle(Theme.ochre)
                    }
                    .frame(height: 160)
                    Text("\(r.streak)-day streak").font(.footnote).foregroundStyle(.secondary)
                }
                Section("Due in the week ahead") {
                    Chart(Array(r.ahead.enumerated()), id: \.offset) { i, n in
                        BarMark(x: .value("Day", i == 0 ? "Today" : "+\(i)"), y: .value("Due", n))
                            .foregroundStyle(Theme.slate)
                    }
                    .frame(height: 140)
                }
                Section("Words that keep slipping") {
                    if r.hard.isEmpty { Text("None yet.").foregroundStyle(.secondary) }
                    ForEach(r.hard, id: \.chinese) { h in
                        HStack {
                            Text(h.chinese).font(.title3)
                            Pinyin.coloured(h.pinyin).italic()
                            Text(h.english).font(.footnote).foregroundStyle(.secondary).lineLimit(1)
                            Spacer()
                            Text("missed \(h.lapses)×").font(.caption.monospacedDigit()).foregroundStyle(Theme.cinnabar)
                        }
                    }
                }
                Section("The wiki") {
                    LabeledContent("Characters", value: "\(r.characters)")
                    LabeledContent("Vocabulary pages", value: "\(r.words)")
                    LabeledContent("Grammar pages", value: "\(r.grammar)")
                    LabeledContent("Readings", value: "\(r.readings)")
                }
            } else {
                ProgressView()
            }
        }
        .navigationTitle("进度 My progress")
        .task { report = try? await engine.progress() }
    }
}

/// The spider's web around a word, as the engine's SpiderWeb() reports it.
struct SpiderWeb: Decodable, Equatable, Sendable {
    struct Node: Decodable, Equatable, Hashable, Sendable {
        let word: String
        let pinyin: String
        let english: String
        let weak: Bool
        let shared: String?
        let links: Int
    }
    let centre: Node?
    let links: [Node]
    let words: Int
    let strands: Int
    let weakStrands: Int
    let walk: Int
}

/// The spider's web (盘丝洞): studied words tied by the characters they
/// share. Slipping words show in cinnabar; walking the web reviews the
/// slipping words around the centre.
struct WebScreen: View {
    @Environment(Engine.self) private var engine
    @State private var web: SpiderWeb?
    @State private var trail: [String] = []
    @State private var walking: [Card]?

    var body: some View {
        List {
            if let web, let c = web.centre {
                Section {
                    VStack(alignment: .leading, spacing: 4) {
                        Text(c.word).font(.system(size: 48)).foregroundStyle(c.weak ? Theme.cinnabar : .primary)
                        Pinyin.coloured(c.pinyin).font(.title3.italic())
                        Text(c.english).foregroundStyle(.secondary)
                    }
                    if web.walk > 0 {
                        Button("Walk the web: review \(web.walk) slipping word\(web.walk == 1 ? "" : "s") here") {
                            Task { walking = try? await engine.walkWeb(c.word) }
                        }
                    }
                } footer: {
                    Text("\(web.words) words, \(web.strands) strands, \(web.weakStrands) with a slipping word")
                }
                Section("Tied by a shared character") {
                    if web.links.isEmpty { Text("No other studied word shares a character with this one yet.").foregroundStyle(.secondary) }
                    ForEach(web.links, id: \.word) { n in
                        Button { trail.append(c.word); Task { await open(n.word) } } label: {
                            HStack {
                                Text(n.word).font(.title3).foregroundStyle(n.weak ? Theme.cinnabar : .primary)
                                Text(n.english).font(.footnote).foregroundStyle(.secondary).lineLimit(1)
                                Spacer()
                                Text("· \(n.shared ?? "") ·").font(.caption).foregroundStyle(Theme.ochre)
                            }
                        }
                    }
                }
            } else if web != nil {
                ContentUnavailableView("No web yet", systemImage: "circle.hexagongrid",
                                       description: Text("Words join the web once they've been reviewed."))
            } else {
                ProgressView()
            }
        }
        .navigationTitle("盘丝洞 The spider's web")
        .toolbar {
            if !trail.isEmpty {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Back to \(trail.last ?? "")") { Task { await open(trail.removeLast()) } }
                }
            }
        }
        .navigationDestination(item: $walking) { cards in
            SessionView(session: StudySession(mode: .review, engine: engine, cards: cards))
        }
        .task { await open("") }
    }

    private func open(_ word: String) async {
        web = try? await engine.spiderWeb(word)
    }
}
