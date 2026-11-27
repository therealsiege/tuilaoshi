import Charts
import SwiftUI

/// A letter or a gazette issue: Chinese with its English.
struct Writing: Decodable, Equatable, Sendable {
    let week: String
    let text: String
    let english: String
    let byHand: Bool?
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
    case letter, gazette, diary, progress
    var id: Self { self }

    var title: String {
        switch self {
        case .letter: "观音来信 Guanyin's letter"
        case .gazette: "邸报 The court gazette"
        case .diary: "日记 Travel diary"
        case .progress: "进度 My progress"
        }
    }

    var symbol: String {
        switch self {
        case .letter: "envelope"
        case .gazette: "newspaper"
        case .diary: "book.closed"
        case .progress: "chart.bar"
        }
    }
}

struct PaperView: View {
    let paper: Paper

    var body: some View {
        switch paper {
        case .letter: WritingView(title: paper.title, load: { try await $0.letter() },
                                  empty: "Guanyin writes each Monday about the week before, once there's been a week with reviews.")
        case .gazette: WritingView(title: paper.title, load: { try await $0.gazette() }, empty: "")
        case .diary: DiaryView()
        case .progress: ProgressScreen()
        }
    }
}

/// A letter or gazette, with its English beneath and read-aloud.
struct WritingView: View {
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
                        Text("Set by the court's clerk from a template; Claude writes it once the game's server is up.")
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

struct DiaryView: View {
    @Environment(Engine.self) private var engine
    @State private var entries: [DiaryEntry] = []

    var body: some View {
        List {
            Section {
                Text("Writing an entry needs Claude to grade it, so it comes with the game's server. Entries written on the Mac show here once progress syncs.")
                    .font(.footnote).foregroundStyle(.secondary)
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
