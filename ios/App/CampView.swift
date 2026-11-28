import SwiftUI

/// Camp: today's numbers, the stage being worked on, the party, and the
/// doors to every other part of the game.
struct CampView: View {
    @Environment(Engine.self) private var engine
    @State private var today: Today?

    var body: some View {
        Group {
            switch engine.status {
            case .opening:
                ProgressView("Opening the journey…")
            case .failed(let message):
                ContentUnavailableView("Camp didn't open", systemImage: "tent", description: Text(message))
            case .ready(let camp):
                ScrollView {
                    VStack(alignment: .leading, spacing: 20) {
                        header(camp)
                        numbers(camp)
                        if let today { TodayPanel(today: today) }
                        if let stage = camp.stage { StageCard(stage: stage) }
                        party(camp)
                        doors
                        papers
                        ProgressSection()
                    }
                    .padding()
                }
                .refreshable { await engine.refresh() }
            }
        }
        .task(id: engine.status) { today = try? await engine.today() }
        .navigationTitle("营地 Camp")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            if case .ready(let camp) = engine.status {
                ToolbarItem(placement: .topBarLeading) {
                    Button { Speaker.chinese(readAloud(camp)) } label: { Image(systemName: "speaker.wave.2") }
                        .accessibilityLabel("Read camp aloud")
                }
            }
            ToolbarItem(placement: .topBarTrailing) {
                NavigationLink { SettingsView() } label: { Image(systemName: "gearshape") }
                    .accessibilityLabel("Settings")
            }
        }
        .navigationDestination(for: Paper.self) { PaperView(paper: $0) }
        .navigationDestination(for: Door.self) { door in
            switch door {
            case .study: StudyView()
            case .read: ReadingListView()
            case .road: RoadView()
            default: DoorView(door: door)
            }
        }
    }

    /// Camp's Chinese, as one passage to read aloud.
    private func readAloud(_ camp: Camp) -> String {
        var parts = ["今天是\(camp.term)。"]
        if let s = camp.stage { parts.append("第\(s.number)难，\(s.name)，在\(s.location)。") }
        parts.append("\(camp.party.joined(separator: "、"))在路上。")
        return parts.joined()
    }

    private func header(_ camp: Camp) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("今日 · \(camp.term)").font(Theme.heading(28))
            Text("Solar term: \(camp.termEnglish)").font(.subheadline).foregroundStyle(.secondary)
        }
    }

    private func numbers(_ camp: Camp) -> some View {
        Grid(alignment: .leading, horizontalSpacing: 12, verticalSpacing: 12) {
            GridRow {
                Number(value: "\(camp.due)", label: "Cards due", tint: Theme.cinnabar, id: "camp.due")
                Number(value: "\(camp.streak)", label: "Day streak", tint: Theme.ochre, id: "camp.streak")
            }
            GridRow {
                Number(value: "\(camp.tongqian)", label: "铜钱 copper", tint: Theme.slate)
                Number(value: "Lv \(camp.level)", label: "\(camp.rank) \(camp.rankEnglish)", tint: Theme.jade)
            }
        }
    }

    private func party(_ camp: Camp) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("师徒 The party").font(Theme.heading(20))
            Text(camp.party.joined(separator: " · ")).font(.body)
            Text("\(count(camp.learned, "character")) learned, \(camp.mastered) mastered")
                .font(.footnote).foregroundStyle(.secondary)
        }
    }

    private var papers: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("营火边 By the fire").font(Theme.heading(20))
            ForEach(Paper.allCases) { p in
                NavigationLink(value: p) {
                    Label(p.title, systemImage: p.symbol)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(12)
                        .background(.background.secondary, in: .rect(cornerRadius: 12))
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("paper.\(p.rawValue)")
            }
        }
    }

    private var doors: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("去哪儿 Where to").font(Theme.heading(20))
            ForEach(Door.allCases) { door in
                NavigationLink(value: door) {
                    Label {
                        VStack(alignment: .leading) {
                            Text(door.title).font(.headline)
                            Text(door.subtitle).font(.subheadline).foregroundStyle(.secondary)
                        }
                    } icon: {
                        Image(systemName: door.symbol).foregroundStyle(Theme.jade)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding()
                    .background(.background.secondary, in: .rect(cornerRadius: 12))
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("door.\(door.rawValue)")
            }
        }
    }
}

/// Camp's Today panel: the day against the goal, the bounties, the
/// festival, the solar term and the Inner Demon, as on the Mac.
private struct TodayPanel: View {
    let today: Today

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("今天 Today").font(Theme.heading(20))
            HStack {
                ProgressView(value: Double(min(today.reviewed, today.dailyGoal)), total: Double(max(today.dailyGoal, 1)))
                    .tint(Theme.jade)
                Text("\(today.reviewed)/\(today.dailyGoal) reviewed").font(.footnote.monospacedDigit())
            }
            if today.reviewed > 0 {
                Text("\(today.correct * 100 / today.reviewed)% right").font(.footnote).foregroundStyle(.secondary)
            }
            ForEach(today.bounties, id: \.text) { b in
                HStack(alignment: .firstTextBaseline) {
                    Image(systemName: b.done ? "checkmark.circle.fill" : "circle").foregroundStyle(b.done ? Theme.jade : .secondary)
                    Text(b.text).font(.subheadline)
                    Spacer()
                    Text(b.done ? "+\(b.pay) 铜钱" : "\(b.have)/\(b.target)").font(.caption.monospacedDigit()).foregroundStyle(.secondary)
                }
            }
            Text(today.shields > 0 ? "All three bounties earn a streak shield · \(today.shields) held" : "All three bounties earn a streak shield")
                .font(.caption).foregroundStyle(.secondary)
            line("节气", "\(today.term.starts ? "今日" : "")\(today.term.name) \(today.term.english)", today.term.text + weather)
            if let f = today.festival {
                line("节日", f.on ? "\(f.name)快乐！ \(f.english)" : "\(f.name) \(f.english)", f.on ? "Today" : "In \(f.days) day\(f.days == 1 ? "" : "s")")
            }
            line("心魔", today.missed >= today.demonWakes ? "The Inner Demon wakes" : "The Inner Demon sleeps",
                 today.missed >= today.demonWakes ? "\(today.missed) words you missed this week" : "\(today.missed) of \(today.demonWakes) missed words this week wake it")
        }
    }

    private var weather: String {
        switch today.term.weather {
        case "cold": " · cold: the party eats more"
        case "heat": " · hot: the party drinks more"
        case "rain": " · rain fills the gourds"
        default: ""
        }
    }

    private func line(_ label: String, _ title: String, _ detail: String) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: 10) {
            Text(label).font(.caption.bold()).foregroundStyle(Theme.ochre)
            VStack(alignment: .leading, spacing: 2) {
                Text(title).font(.subheadline)
                Text(detail).font(.caption).foregroundStyle(.secondary)
            }
        }
    }
}

/// Where progress is kept: syncing with the Mac, and exporting it.
private struct ProgressSection: View {
    @Environment(Engine.self) private var engine
    @State private var exported: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("进度 Progress").font(Theme.heading(20))
            Text(syncLine).font(.footnote).foregroundStyle(.secondary)
            HStack {
                Button("Sync now") { Task { await engine.syncNow() } }
                Button("Export") {
                    Task {
                        do {
                            let dir = try await engine.export()
                            exported = "Exported to Files › On My iPhone › 推老师 › \(dir.lastPathComponent)"
                        } catch {
                            exported = "Couldn't export: \(error.localizedDescription)"
                        }
                    }
                }
            }
            .buttonStyle(.bordered)
            if let exported { Text(exported).font(.footnote).foregroundStyle(.secondary) }
        }
    }

    private var syncLine: String {
        switch engine.sync {
        case .never: "Not synced with the Mac yet."
        case .off(let why): why
        case .synced(let at): "Synced with your other devices \(at.formatted(.relative(presentation: .named)))."
        case .failed(let why): why
        }
    }
}

/// "1 character", "2 characters".
private func count(_ n: Int, _ noun: String) -> String {
    "\(n) \(noun)\(n == 1 ? "" : "s")"
}

private struct Number: View {
    let value: String
    let label: String
    let tint: Color
    var id = ""

    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(value).font(.title.bold().monospacedDigit()).foregroundStyle(tint)
                .accessibilityIdentifier(id)
            Text(label).font(.footnote).foregroundStyle(.secondary).lineLimit(2)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(.background.secondary, in: .rect(cornerRadius: 12))
    }
}

private struct StageCard: View {
    let stage: Camp.Stage

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("第\(stage.number)难 · \(stage.region)").font(.caption).foregroundStyle(.secondary)
            Text(stage.name).font(Theme.heading(24))
            Text(stage.english).font(.subheadline)
            Text("Facing \(stage.demon) \(stage.demonEnglish) at \(stage.location)")
                .font(.footnote).foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .overlay(RoundedRectangle(cornerRadius: 12).stroke(Theme.cinnabar.opacity(0.5)))
    }
}

/// The places camp leads to. Each becomes its own screen as the iOS work
/// goes on (#57 study, #61 the road, #62 towns).
enum Door: String, CaseIterable, Identifiable, Hashable {
    case study, read, road, towns

    var id: Self { self }

    var title: String {
        switch self {
        case .study: "复习 Study"
        case .read: "读 Read"
        case .road: "上路 The road"
        case .towns: "城镇 Towns"
        }
    }

    var subtitle: String {
        switch self {
        case .study: "Today's reviews and new words"
        case .read: "Readings and the journey's episodes, glossed word by word"
        case .road: "Walk the region to the next tribulation"
        case .towns: "Shops, alms and notices"
        }
    }

    var symbol: String {
        switch self {
        case .study: "rectangle.stack"
        case .read: "book"
        case .road: "map"
        case .towns: "building.columns"
        }
    }
}

private struct DoorView: View {
    let door: Door

    var body: some View {
        ContentUnavailableView(door.title, systemImage: door.symbol,
                               description: Text("This part of the journey isn't built yet."))
            .navigationTitle(door.title)
    }
}
