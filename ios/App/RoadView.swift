import SwiftUI
import UIKit

/// The region the pilgrim walks, as the engine's Road() reports it. Tiles
/// are rows of letters: B blocked, R the road, G wild ground, T a town,
/// S a tribulation, E the door east, W the door west, C the crossing.
struct RoadMap: Decodable, Equatable, Sendable {
    struct Region: Decodable, Equatable, Sendable { let id: String; let name: String; let english: String; let terrain: String }
    struct Town: Decodable, Equatable, Sendable { let id: String; let name: String; let english: String; let x: Int; let y: Int }
    struct Stage: Decodable, Equatable, Sendable { let number: Int; let x: Int; let y: Int; let state: String }
    struct Crossing: Decodable, Equatable, Sendable { let kind: String; let name: String; let english: String; let x: Int; let y: Int; let crossed: Bool }

    let region: Region
    let width: Int
    let height: Int
    let tiles: [String]
    let towns: [Town]
    let stages: [Stage]
    let crossing: Crossing?
    let pos: RoadPoint
}

struct RoadPoint: Decodable, Equatable, Sendable { let x: Int; let y: Int }

/// What a step came to, as the engine's Step() reports it.
struct StepOutcome: Decodable, Equatable, Sendable {
    struct Name: Decodable, Equatable, Sendable { let id: String?; let name: String; let english: String }
    let kind: String // moved, blocked, town, tribulation, encounter, region, crossing, event or plot
    let text: String?
    let chinese: String?
    let stage: Int?
    let fight: Bool?
    let town: RoadMap.Town?
    let soldier: Name?
    let region: Name?
    let day: String?       // a day ended on this step, and what it brought
    let collapsed: Bool?   // the master collapsed and was carried back
    let event: RoadEvent?  // an event or the crossing, to choose
    let plot: RoadPlot?    // demons overheard, to answer
    let gate: RoadGate?    // the lair's grammar gate, to lay
    let pos: RoadPoint
}

extension StepOutcome {
    /// A banner of the app's own, where the engine said nothing.
    init(blocked text: String, at pos: RoadPoint) {
        self.init(kind: "blocked", text: text, chinese: nil, stage: nil, fight: nil, town: nil, soldier: nil, region: nil,
                  day: nil, collapsed: nil, event: nil, plot: nil, gate: nil, pos: pos)
    }

    /// What it brought that opens as a sheet, if anything.
    var sheet: RoadSheet? {
        if let event { return .event(event) }
        if let plot { return .plot(plot) }
        if let gate { return .gate(gate) }
        return nil
    }
}

/// The road: the region's tile map over its painting, where there is one,
/// walked a cell at a time by tapping a neighbouring tile, swiping, or the
/// direction pad. The camera keeps on the pilgrim. The trail's day and
/// supplies run under the region's name; the day's events, demons
/// overheard, the region's crossing and the grammar gate before its lair
/// play as sheets (RoadSheets.swift). Battles and towns come to the phone
/// later (#62, #63): meeting one only says so.
struct RoadView: View {
    @Environment(Engine.self) private var engine
    @State private var map: RoadMap?
    @State private var pos = RoadPoint(x: 0, y: 0)
    @State private var banner: StepOutcome?
    @State private var walking = false
    @State private var queued: [RoadPoint] = [] // steps asked for while one is under way
    @State private var trail: Trail?
    @State private var note: String? // the last day's news, or the last change to the trail
    @State private var sheet: RoadSheet?
    @State private var error: String?

    /// Columns across the screen; the rest of the region scrolls in.
    private let across: CGFloat = 13

    var body: some View {
        Group {
            if let map {
                GeometryReader { geo in
                    // As wide as across cells, and short enough on a small phone to leave the rest room.
                    let cell = max(16, min(geo.size.width / across, (geo.size.height - reserved) / CGFloat(map.height)))
                    VStack(spacing: 10) {
                        header(map)
                        if let trail { TrailStrip(trail: trail, note: note) { c in Task { await change(c) } } }
                        board(map, cell: cell, width: geo.size.width)
                            .frame(height: cell * CGFloat(map.height))
                        bannerView
                        Spacer(minLength: 0)
                        pad
                    }
                }
                .padding(.vertical)
            } else if let error {
                ContentUnavailableView("The road didn't open", systemImage: "map", description: Text(error))
            } else {
                ProgressView("Finding the road…")
            }
        }
        .navigationTitle("上路 The road")
        .navigationBarTitleDisplayMode(.inline)
        .task { await load() }
        .sheet(item: $sheet) { s in
            Group {
                switch s {
                case .event(let e):
                    EventSheet(event: e, choose: { try await engine.choose($0) }, walkOn: walkOn)
                case .plot(let p):
                    PlotSheet(plot: p, answer: { try await engine.answer($0) }, walkOn: walkOn)
                case .gate(let g):
                    GateSheet(gate: g, lay: { try await engine.lay($0) }, walkOn: walkOn)
                }
            }
            .environment(engine)
        }
    }

    /// Height kept for the header, the trail, the banner and the pad.
    private let reserved: CGFloat = 330

    private func load() async {
        do {
            let m = try await engine.road()
            map = m
            pos = m.pos
            trail = try await engine.trail()
        } catch {
            self.error = error.localizedDescription
        }
    }

    private func change(_ c: TrailChange) async {
        do {
            let t = try await engine.trail(c)
            (trail, note) = (t, t.note)
        } catch { banner = StepOutcome(blocked: error.localizedDescription, at: pos) }
    }

    /// Closes the sheet met: across the river once a way is chosen, past
    /// the gate, or on along the road.
    private func walkOn() {
        Task {
            do {
                let o = try await engine.walkOn()
                sheet = nil
                pos = o.pos
                banner = (o.text ?? "").isEmpty ? nil : o
                if map?.crossing.map({ !$0.crossed && $0.x == o.pos.x && $0.y == o.pos.y }) == true {
                    await load() // across: the crossing shows crossed
                } else {
                    trail = try await engine.trail()
                }
            } catch {
                banner = StepOutcome(blocked: error.localizedDescription, at: pos)
            }
        }
    }

    private func header(_ map: RoadMap) -> some View {
        HStack(alignment: .firstTextBaseline) {
            VStack(alignment: .leading, spacing: 2) {
                Text(map.region.name).font(Theme.heading(24))
                Text(map.region.english).font(.subheadline).foregroundStyle(.secondary)
            }
            Spacer()
            Text("\(pos.x), \(pos.y)")
                .font(.caption.monospacedDigit()).foregroundStyle(.secondary)
                .accessibilityLabel("Position \(pos.x), \(pos.y)")
                .accessibilityIdentifier("road.position")
            Button { Speaker.chinese(map.region.name) } label: { Image(systemName: "speaker.wave.2") }
                .accessibilityLabel("Read the region's name aloud")
        }
        .padding(.horizontal)
    }

    // MARK: The map

    /// The region's tiles over its painting, offset so the pilgrim stays in view.
    private func board(_ map: RoadMap, cell: CGFloat, width: CGFloat) -> some View {
        let full = CGFloat(map.width) * cell
        let camera = min(0, max(width - full, width / 2 - (CGFloat(pos.x) + 0.5) * cell))
        return ZStack(alignment: .topLeading) {
            if let painting = UIImage(named: "map-\(map.region.id)") {
                Image(uiImage: painting).resizable().opacity(0.55)
            }
            TileGrid(map: map, painted: UIImage(named: "map-\(map.region.id)") != nil)
            pilgrim(cell)
                .offset(x: CGFloat(pos.x) * cell, y: CGFloat(pos.y) * cell)
        }
        .frame(width: full, height: CGFloat(map.height) * cell)
        .contentShape(Rectangle())
        .onTapGesture(coordinateSpace: .local) { at in
            let dx = Int(at.x / cell) - pos.x, dy = Int(at.y / cell) - pos.y
            if abs(dx) + abs(dy) == 1 { walk(dx, dy) }
        }
        .offset(x: camera)
        .animation(.easeOut(duration: 0.2), value: pos)
        .frame(width: width, alignment: .leading)
        .clipped()
        .gesture(DragGesture(minimumDistance: 24).onEnded { g in
            let t = g.translation
            if abs(t.width) > abs(t.height) { walk(t.width > 0 ? 1 : -1, 0) } else { walk(0, t.height > 0 ? 1 : -1) }
        })
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("Map of \(map.region.english). Use the direction buttons to walk.")
        .accessibilityIdentifier("road.map")
    }

    private func pilgrim(_ cell: CGFloat) -> some View {
        Text("僧")
            .font(.system(size: cell * 0.55, weight: .bold))
            .foregroundStyle(.white)
            .frame(width: cell * 0.86, height: cell * 0.86)
            .background(Theme.jade, in: .circle)
            .overlay(Circle().stroke(.white, lineWidth: 2))
            .shadow(radius: 2)
            .frame(width: cell, height: cell)
    }

    // MARK: Walking

    /// Asks for one step; steps asked for quickly are taken in turn.
    private func walk(_ dx: Int, _ dy: Int) {
        guard map != nil else { return }
        queued.append(RoadPoint(x: dx, y: dy))
        guard !walking else { return }
        walking = true
        Task {
            while !queued.isEmpty { await take(queued.removeFirst()) }
            walking = false
        }
    }

    private func take(_ d: RoadPoint) async {
        do {
            let o = try await engine.step(dx: d.x, dy: d.y)
            pos = o.pos
            banner = ["moved", "event", "plot"].contains(o.kind) || o.gate != nil ? nil : o
            if o.kind == "region" {
                queued.removeAll() // the new region's map comes first
                await load()
            } else if o.day != nil || o.sheet != nil {
                trail = try? await engine.trail()
            }
            if let day = o.day { note = day }
            if let next = o.sheet {
                queued.removeAll() // what the road brought comes first
                sheet = next
            }
            if o.collapsed == true, let day = o.day { banner = StepOutcome(blocked: day, at: o.pos) }
            if ["town", "encounter", "region", "crossing", "event"].contains(o.kind), let zh = o.chinese { Speaker.chinese(zh) }
        } catch {
            banner = StepOutcome(blocked: error.localizedDescription, at: pos)
        }
    }

    @ViewBuilder private var bannerView: some View {
        if let b = banner {
            HStack(alignment: .top, spacing: 10) {
                Image(systemName: symbol(b.kind)).foregroundStyle(tint(b.kind))
                VStack(alignment: .leading, spacing: 2) {
                    if let s = b.soldier {
                        Text("\(s.name) \(s.english)").font(.headline)
                    }
                    Text(b.text ?? "").font(.subheadline)
                }
                Spacer(minLength: 0)
                if let zh = b.chinese, !zh.isEmpty {
                    Button { Speaker.chinese(zh) } label: { Image(systemName: "speaker.wave.2") }
                        .accessibilityLabel("Read it aloud")
                }
            }
            .padding(12)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(tint(b.kind).opacity(0.12), in: .rect(cornerRadius: 12))
            .padding(.horizontal)
            .accessibilityElement(children: .combine)
            .accessibilityIdentifier("road.banner")
        }
    }

    private func symbol(_ kind: String) -> String {
        switch kind {
        case "blocked": "hand.raised"
        case "tribulation": "flame"
        case "encounter": "exclamationmark.triangle"
        case "crossing": "water.waves"
        case "town": "building.columns"
        default: "map"
        }
    }

    private func tint(_ kind: String) -> Color {
        switch kind {
        case "blocked", "crossing": Theme.slate
        case "tribulation", "encounter": Theme.cinnabar
        case "town": Theme.ochre
        default: Theme.jade
        }
    }

    /// The direction pad, for one step a press.
    private var pad: some View {
        Grid(horizontalSpacing: 8, verticalSpacing: 8) {
            GridRow {
                Color.clear.gridCellUnsizedAxes([.horizontal, .vertical])
                padButton("north", "chevron.up", 0, -1)
                Color.clear.gridCellUnsizedAxes([.horizontal, .vertical])
            }
            GridRow {
                padButton("west", "chevron.left", -1, 0)
                Image(systemName: "figure.walk").foregroundStyle(.secondary).frame(width: 52, height: 44)
                padButton("east", "chevron.right", 1, 0)
            }
            GridRow {
                Color.clear.gridCellUnsizedAxes([.horizontal, .vertical])
                padButton("south", "chevron.down", 0, 1)
                Color.clear.gridCellUnsizedAxes([.horizontal, .vertical])
            }
        }
    }

    private func padButton(_ name: String, _ symbol: String, _ dx: Int, _ dy: Int) -> some View {
        Button { walk(dx, dy) } label: {
            Image(systemName: symbol).font(.title3.bold()).frame(width: 52, height: 44)
        }
        .buttonStyle(.bordered)
        .tint(Theme.jade)
        .accessibilityLabel("Walk \(name)")
        .accessibilityIdentifier("road.\(name)")
    }
}

/// The region's tiles in the app's colours, drawn faint over a painting
/// where the region has one.
private struct TileGrid: View {
    let map: RoadMap
    let painted: Bool

    var body: some View {
        Canvas { ctx, size in
            let cell = size.width / CGFloat(map.width)
            let rows = map.tiles.map(Array.init)
            let alpha = painted ? 0.28 : 1
            for (y, row) in rows.enumerated() {
                for (x, t) in row.enumerated() {
                    let r = CGRect(x: CGFloat(x) * cell, y: CGFloat(y) * cell, width: cell, height: cell)
                    ctx.fill(Path(r), with: .color(colour(t).opacity(alpha)))
                    if !painted || t != "B" {
                        ctx.stroke(Path(r), with: .color(.black.opacity(0.06)), lineWidth: 0.5)
                    }
                    if let mark = mark(t) {
                        ctx.draw(Text(mark).font(.system(size: cell * 0.5, weight: .semibold)).foregroundColor(.white), at: CGPoint(x: r.midX, y: r.midY))
                    }
                }
            }
            for s in map.stages {
                let r = CGRect(x: CGFloat(s.x) * cell, y: CGFloat(s.y) * cell, width: cell, height: cell).insetBy(dx: cell * 0.08, dy: cell * 0.08)
                ctx.fill(Path(roundedRect: r, cornerRadius: cell * 0.18), with: .color(stageColour(s.state)))
                ctx.draw(Text("\(s.number)").font(.system(size: cell * 0.42, weight: .bold).monospacedDigit()).foregroundColor(.white),
                         at: CGPoint(x: r.midX, y: r.midY))
            }
            for t in map.towns {
                let r = CGRect(x: CGFloat(t.x) * cell, y: CGFloat(t.y) * cell, width: cell, height: cell).insetBy(dx: cell * 0.08, dy: cell * 0.08)
                ctx.fill(Path(roundedRect: r, cornerRadius: cell * 0.18), with: .color(Theme.cinnabar.opacity(0.85)))
                ctx.draw(Text(String(t.name.prefix(1))).font(.system(size: cell * 0.5, weight: .bold)).foregroundColor(.white),
                         at: CGPoint(x: r.midX, y: r.midY))
            }
            if let c = map.crossing {
                let r = CGRect(x: CGFloat(c.x) * cell, y: CGFloat(c.y) * cell, width: cell, height: cell).insetBy(dx: cell * 0.08, dy: cell * 0.08)
                ctx.fill(Path(roundedRect: r, cornerRadius: cell * 0.18), with: .color(c.crossed ? Theme.jade : Theme.slate))
                ctx.draw(Text(String(c.name.prefix(1))).font(.system(size: cell * 0.5, weight: .bold)).foregroundColor(.white),
                         at: CGPoint(x: r.midX, y: r.midY))
            }
        }
    }

    private func colour(_ t: Character) -> Color {
        switch t {
        case "R": Theme.ochre.opacity(0.55)
        case "G": Theme.jade.opacity(0.45)
        case "E", "W": Theme.ochre
        case "B": Theme.slate.opacity(0.8)
        default: Theme.ochre.opacity(0.55) // towns, tribulations and the crossing stand on the road
        }
    }

    private func mark(_ t: Character) -> String? {
        switch t {
        case "E": "→"
        case "W": "←"
        default: nil
        }
    }

    private func stageColour(_ state: String) -> Color {
        switch state {
        case "cleared": Theme.jade
        case "open": Theme.cinnabar
        default: Theme.slate
        }
    }
}
