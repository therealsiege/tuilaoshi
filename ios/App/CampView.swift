import SwiftUI

/// Camp: today's numbers, the stage being worked on, the party, and the
/// doors to every other part of the game.
struct CampView: View {
    @Environment(Engine.self) private var engine

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
                        if let stage = camp.stage { StageCard(stage: stage) }
                        party(camp)
                        doors
                        ProgressSection()
                    }
                    .padding()
                }
                .refreshable { await engine.refresh() }
            }
        }
        .navigationTitle("营地 Camp")
        .navigationBarTitleDisplayMode(.inline)
        .navigationDestination(for: Door.self) { door in
            switch door {
            case .study: StudyView()
            default: DoorView(door: door)
            }
        }
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
        case .synced(let at): "Synced with the Mac through iCloud Drive \(at.formatted(.relative(presentation: .named)))."
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
    case study, road, towns

    var id: Self { self }

    var title: String {
        switch self {
        case .study: "复习 Study"
        case .road: "上路 The road"
        case .towns: "城镇 Towns"
        }
    }

    var subtitle: String {
        switch self {
        case .study: "Today's reviews and new words"
        case .road: "Walk the region to the next tribulation"
        case .towns: "Shops, alms and notices"
        }
    }

    var symbol: String {
        switch self {
        case .study: "rectangle.stack"
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
