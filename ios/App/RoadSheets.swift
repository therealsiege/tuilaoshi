import SwiftUI

/// The trail, as the engine's Trail() reports it: the day, the supplies,
/// the purse, pace and rations, the party's health and spirits, and HP.
struct Trail: Decodable, Equatable, Sendable {
    struct Member: Decodable, Equatable, Sendable { let name: String; let health: Int; let morale: Int; let ill: Bool? }
    struct Term: Decodable, Equatable, Sendable { let name: String; let english: String; let weather: String; let text: String }
    let day: Int
    let steps: Int
    let food: Int
    let water: Int
    let foodLow: Bool?
    let waterLow: Bool?
    let coins: Int
    let pace: String
    let paceNote: String
    let rations: String
    let rationNote: String
    let term: Term?
    let sandalDays: Int
    let sandalsWorn: Bool?
    let forewarned: Bool?
    let party: [Member]
    let hp: Int
    let maxHP: Int
    let note: String?
}

enum TrailChange: Sendable { case pace, rations, rest }

/// An event or the region's crossing, met by a choice. The English shows
/// once chosen, as on the Mac.
struct RoadEvent: Decodable, Equatable, Sendable {
    struct Choice: Decodable, Equatable, Sendable { let text: String; let english: String? }
    let id: String
    let crossing: Bool?
    let day: Int
    let text: String
    let english: String?
    let choices: [Choice]
    let chose: Int
    let result: String?
    let resultEnglish: String?
    let effect: String?
    let forewarned: Bool?
}

/// Demons overheard plotting at night; answered rightly, the party is forewarned.
struct RoadPlot: Decodable, Equatable, Sendable {
    struct Line: Decodable, Equatable, Sendable { let who: String; let text: String; let english: String? }
    struct Choice: Decodable, Equatable, Sendable { let text: String; let english: String?; let right: Bool? }
    let night: Int
    let lines: [Line]
    let speech: String
    let question: String
    let questionEnglish: String?
    let choices: [Choice]
    let chose: Int
    let forewarned: Bool?
    let warning: String?
    let warningEnglish: String?
}

/// The grammar gate before a lair: five sentences laid from word tiles.
struct RoadGate: Decodable, Equatable, Sendable {
    struct Example: Decodable, Equatable, Sendable { let chinese: String; let english: String }
    struct Tile: Decodable, Equatable, Sendable { let word: String; let laid: Bool? }
    let region: String
    let pattern: String
    let pinyin: String
    let explain: String
    let examples: [Example]
    let round: Int
    let rounds: Int
    let passed: Bool?
    let typed: Bool?
    let english: String?
    let tiles: [Tile]
    let built: String
    let note: String?
    let laid: String?
}

/// What the road brought, shown as a sheet until met.
enum RoadSheet: Identifiable, Equatable {
    case event(RoadEvent)
    case plot(RoadPlot)
    case gate(RoadGate)

    var id: String {
        switch self {
        case .event(let e): "event-\(e.id)"
        case .plot(let p): "plot-\(p.night)"
        case .gate(let g): "gate-\(g.region)"
        }
    }
}

/// The trail in a line under the region's name: day, 斋饭, water and
/// 铜钱, with the pace, rations and a day's rest a tap away.
struct TrailStrip: View {
    let trail: Trail
    let note: String? // what the last day or change brought
    let change: (TrailChange) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            HStack(spacing: 10) {
                Text("Day \(trail.day)").font(.subheadline.bold())
                supply("斋饭", trail.food, low: trail.foodLow == true)
                supply("水", trail.water, low: trail.waterLow == true)
                supply("铜钱", trail.coins, low: false)
                Text("HP \(trail.hp)/\(trail.maxHP)").font(.caption.monospacedDigit()).foregroundStyle(.secondary)
                Spacer(minLength: 0)
                Menu {
                    Button("Pace: \(trail.pace)\(trail.paceNote)") { change(.pace) }
                    Button("Rations: \(trail.rations)\(trail.rationNote)") { change(.rations) }
                    Button("Rest a day") { change(.rest) }
                } label: {
                    Image(systemName: "slider.horizontal.3")
                }
                .accessibilityLabel("Pace, rations and rest")
                .accessibilityIdentifier("road.trailMenu")
            }
            .accessibilityElement(children: .combine)
            .accessibilityLabel("Day \(trail.day), 斋饭 \(trail.food), water \(trail.water), 铜钱 \(trail.coins)")
            .accessibilityIdentifier("road.trail")
            if let line = statusLine {
                Text(line).font(.caption).foregroundStyle(.secondary).lineLimit(2)
                    .accessibilityIdentifier("road.trailNote")
            }
        }
        .padding(.horizontal)
    }

    private var statusLine: String? {
        var parts: [String] = []
        if let note, !note.isEmpty { parts.append(note) }
        if trail.forewarned == true { parts.append("Forewarned: the next danger does half its harm.") }
        if trail.sandalsWorn == true { parts.append("草鞋 worn out: buy new in a town.") }
        if let ill = trail.party.first(where: { $0.health < 40 }) { parts.append("\(ill.name) ♥\(ill.health)") }
        return parts.isEmpty ? nil : parts.joined(separator: " · ")
    }

    private func supply(_ name: String, _ n: Int, low: Bool) -> some View {
        Text("\(name) \(n)").font(.subheadline.monospacedDigit()).foregroundStyle(low ? Theme.cinnabar : .primary)
    }
}

/// A sheet's frame: a scrolling column of its content and, at the foot,
/// the button to walk on once met.
private struct SheetFrame<Content: View>: View {
    let title: String
    @ViewBuilder let content: Content

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 14) { content }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding()
            }
            .navigationTitle(title)
            .navigationBarTitleDisplayMode(.inline)
        }
        .interactiveDismissDisabled()
    }
}

/// A line of Chinese with a button to hear it.
private struct Spoken: View {
    let text: String
    var font: Font = .title3

    var body: some View {
        HStack(alignment: .firstTextBaseline) {
            Text(text).font(font).frame(maxWidth: .infinity, alignment: .leading)
            Button { Speaker.chinese(text) } label: { Image(systemName: "speaker.wave.2") }
                .accessibilityLabel("Read it aloud")
        }
    }
}

@MainActor private func choiceButton(_ text: String, english: String?, tint: Color, id: String, action: @escaping () -> Void) -> some View {
    Button(action: action) {
        VStack(alignment: .leading, spacing: 2) {
            Text(text).font(.body)
            if let english { Text(english).font(.caption).foregroundStyle(.secondary) }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
    .buttonStyle(.bordered)
    .tint(tint)
    .accessibilityIdentifier(id)
}

@MainActor private func walkOnButton(_ id: String, action: @escaping () -> Void) -> some View {
    Button(action: action) { Text("Walk on").frame(maxWidth: .infinity) }
        .buttonStyle(.borderedProminent)
        .tint(Theme.jade)
        .accessibilityIdentifier(id)
}

/// An event on the trail, or the region's crossing: its Chinese and
/// choices, then what came of the one chosen.
struct EventSheet: View {
    @State var event: RoadEvent
    let choose: (Int) async throws -> RoadEvent
    let walkOn: () -> Void
    @State private var error: String?

    var body: some View {
        SheetFrame(title: event.crossing == true ? "过 The crossing" : "Day \(event.day) on the road") {
            Spoken(text: event.text)
            if let en = event.english { Text(en).foregroundStyle(.secondary) }
            if event.forewarned == true, event.chose < 0 {
                Label("Forewarned: whatever harm comes is halved.", systemImage: "shield").foregroundStyle(Theme.jade)
            }
            ForEach(Array(event.choices.enumerated()), id: \.offset) { i, c in
                choiceButton(c.text, english: c.english, tint: i == event.chose ? Theme.ochre : Theme.slate, id: "event.choice.\(i)") {
                    Task { await pick(i) }
                }
                .allowsHitTesting(event.chose < 0)
            }
            if event.chose >= 0 {
                if let r = event.result { Spoken(text: r, font: .body) }
                if let r = event.resultEnglish { Text(r).font(.subheadline).foregroundStyle(.secondary) }
                if let fx = event.effect { Text(fx).font(.subheadline.bold()).foregroundStyle(Theme.ochre) }
                walkOnButton("event.walkOn", action: walkOn)
            }
            if let error { Text(error).foregroundStyle(Theme.cinnabar) }
        }
        .accessibilityIdentifier(event.crossing == true ? "event.crossing" : "event")
    }

    private func pick(_ i: Int) async {
        do { event = try await choose(i) } catch { self.error = error.localizedDescription }
    }
}

/// Demons overheard plotting: their lines (read aloud on a tap), the
/// question and its choices; once answered, the English and what came of it.
struct PlotSheet: View {
    @State var plot: RoadPlot
    let answer: (Int) async throws -> RoadPlot
    let walkOn: () -> Void
    @State private var error: String?

    var body: some View {
        SheetFrame(title: "Night \(plot.night) · 妖怪密谋") {
            if plot.chose < 0 {
                Text("The fire burns low. In the dark, two demons are whispering…").font(.subheadline).foregroundStyle(.secondary)
            }
            Button { Speaker.chinese(plot.speech) } label: { Label("Hear them", systemImage: "ear") }
                .accessibilityIdentifier("plot.hear")
            ForEach(Array(plot.lines.enumerated()), id: \.offset) { _, l in
                VStack(alignment: .leading, spacing: 2) {
                    HStack(alignment: .firstTextBaseline, spacing: 8) {
                        Text(l.who).font(.subheadline.bold()).foregroundStyle(Theme.cinnabar)
                        Spoken(text: l.text, font: .body)
                    }
                    if let en = l.english { Text(en).font(.caption).foregroundStyle(.secondary) }
                }
            }
            Spoken(text: plot.question, font: .headline)
            if let en = plot.questionEnglish { Text(en).font(.subheadline).foregroundStyle(.secondary) }
            ForEach(Array(plot.choices.enumerated()), id: \.offset) { i, c in
                choiceButton(c.text, english: c.english, tint: tint(i, c), id: "plot.choice.\(i)") { Task { await pick(i) } }
                    .allowsHitTesting(plot.chose < 0)
            }
            if plot.chose >= 0 {
                if plot.forewarned == true {
                    if let w = plot.warning { Spoken(text: w, font: .body) }
                    if let w = plot.warningEnglish { Text(w).font(.subheadline).foregroundStyle(.secondary) }
                    Label("Forewarned: the next danger on the road does half its harm.", systemImage: "shield").foregroundStyle(Theme.jade)
                } else {
                    Text("The demons slip away into the night. The road ahead is as it was.").foregroundStyle(Theme.ochre)
                }
                walkOnButton("plot.walkOn", action: walkOn)
            }
            if let error { Text(error).foregroundStyle(Theme.cinnabar) }
        }
        .accessibilityIdentifier("plot")
    }

    private func tint(_ i: Int, _ c: RoadPlot.Choice) -> Color {
        guard plot.chose >= 0 else { return Theme.slate }
        if c.right == true { return Theme.jade }
        return i == plot.chose ? Theme.cinnabar : Theme.slate
    }

    private func pick(_ i: Int) async {
        do { plot = try await answer(i) } catch { self.error = error.localizedDescription }
    }
}

/// The grammar gate: its pattern and model sentences, then each round's
/// tiles, tapped (or dragged up into the sentence) in order.
struct GateSheet: View {
    @State var gate: RoadGate
    let lay: (Int?) async throws -> RoadGate
    let walkOn: () -> Void
    @State private var error: String?

    var body: some View {
        SheetFrame(title: "语法关 Grammar gate") {
            HStack(alignment: .firstTextBaseline) {
                Text(gate.pattern).font(Theme.heading(28))
                Text(gate.pinyin).foregroundStyle(.secondary)
                Spacer()
                if gate.passed != true {
                    Text("Round \(gate.round + 1) of \(gate.rounds)").font(.caption.monospacedDigit()).foregroundStyle(.secondary)
                        .accessibilityIdentifier("gate.round")
                }
            }
            Text(gate.explain).font(.subheadline)
            ForEach(Array(gate.examples.prefix(2).enumerated()), id: \.offset) { _, e in
                VStack(alignment: .leading, spacing: 2) {
                    Spoken(text: "例 " + e.chinese, font: .body)
                    Text(e.english).font(.caption).foregroundStyle(.secondary)
                }
            }
            Divider()
            if gate.passed == true {
                note
                Label("语法关已过！ The gate opens: the demon's lair lies ahead.", systemImage: "checkmark.seal").foregroundStyle(Theme.jade)
                walkOnButton("gate.walkOn", action: walkOn)
            } else {
                if let en = gate.english { Text("Lay the tiles in order: \(en)").font(.headline) }
                built
                FlowLayout(spacing: 8, lineSpacing: 8) {
                    ForEach(Array(gate.tiles.enumerated()), id: \.offset) { i, t in
                        Button { Task { await place(i) } } label: { Text(t.word).font(.title3) }
                            .buttonStyle(.bordered)
                            .tint(t.laid == true ? Theme.slate : Theme.ochre)
                            .disabled(t.laid == true)
                            .draggable("\(i)")
                            .accessibilityIdentifier("gate.tile.\(i)")
                    }
                }
                note
                HStack {
                    Button("Undo") { Task { await place(nil) } }.disabled(gate.built.isEmpty)
                        .accessibilityIdentifier("gate.undo")
                    Spacer()
                    Button("Leave the gate", role: .cancel, action: walkOn).accessibilityIdentifier("gate.leave")
                }
            }
            if let error { Text(error).foregroundStyle(Theme.cinnabar) }
        }
        .accessibilityIdentifier("gate")
    }

    /// The sentence laid so far; tiles dropped on it are laid next.
    private var built: some View {
        Text(gate.built.isEmpty ? "…" : gate.built)
            .font(.title2)
            .foregroundStyle(gate.built.isEmpty ? .secondary : .primary)
            .frame(maxWidth: .infinity, minHeight: 44, alignment: .leading)
            .padding(.horizontal, 10)
            .background(Theme.ochre.opacity(0.12), in: .rect(cornerRadius: 10))
            .dropDestination(for: String.self) { items, _ in
                guard let i = items.first.flatMap(Int.init) else { return false }
                Task { await place(i) }
                return true
            }
            .accessibilityIdentifier("gate.built")
    }

    @ViewBuilder private var note: some View {
        if let n = gate.note, !n.isEmpty {
            Text(n).font(.subheadline).foregroundStyle(n.hasPrefix("对") ? Theme.jade : Theme.cinnabar)
                .accessibilityIdentifier("gate.note")
        }
    }

    private func place(_ i: Int?) async {
        do {
            gate = try await lay(i)
            if gate.laid == "right" { Cue.right.play() } else if gate.laid == "wrong" { Cue.wrong.play() }
        } catch {
            self.error = error.localizedDescription
        }
    }
}
