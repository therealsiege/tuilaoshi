import SwiftUI

/// Something to read, as the Go engine's Readings() lists it.
struct ReadingItem: Decodable, Identifiable, Hashable {
    let id: String
    let title: String
    let tag: String
}

/// An opened reading: lines split into words, each word glossed.
struct Reading: Decodable {
    struct Token: Decodable, Hashable {
        let text: String
        let word: Bool
        let pinyin: String?
        let english: String?
        let inDeck: Bool?
    }

    struct Line: Decodable {
        let speaker: String?
        let tokens: [Token]
        let english: String?

        var text: String { tokens.map(\.text).joined() }
    }

    let title: String
    let lines: [Line]
}

/// The readings and journey episodes reached so far.
struct ReadingListView: View {
    @Environment(Engine.self) private var engine
    @State private var items: [ReadingItem] = []
    @State private var error: String?

    var body: some View {
        List(items) { item in
            NavigationLink(value: item) {
                VStack(alignment: .leading, spacing: 2) {
                    Text(item.title).font(.headline)
                    Text(item.tag).font(.caption).foregroundStyle(.secondary)
                }
            }
        }
        .overlay {
            if let error { ContentUnavailableView("Nothing to read", systemImage: "book", description: Text(error)) }
        }
        .navigationTitle("读 Read")
        .navigationDestination(for: ReadingItem.self) { ReaderView(item: $0) }
        .task {
            do { items = try await engine.readings() } catch { self.error = error.localizedDescription }
        }
    }
}

/// A text to read word by word: tap a word for its gloss, mark it unknown,
/// or keep it with its sentence as a card.
struct ReaderView: View {
    let item: ReadingItem
    @Environment(Engine.self) private var engine
    @State private var reading: Reading?
    @State private var selected: (line: Int, token: Reading.Token)?
    @State private var showEnglish = false
    @State private var marked: Set<String> = []
    @State private var status = ""
    @State private var error: String?

    var body: some View {
        VStack(spacing: 0) {
            ScrollView {
                VStack(alignment: .leading, spacing: 18) {
                    if let reading {
                        ForEach(reading.lines.indices, id: \.self) { i in line(i, reading.lines[i]) }
                    } else if let error {
                        Text(error).foregroundStyle(Theme.cinnabar)
                    } else {
                        ProgressView()
                    }
                }
                .padding()
            }
            if let selected, let reading { gloss(selected.token, in: reading.lines[selected.line]) }
        }
        .navigationTitle(item.title)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Toggle("English", isOn: $showEnglish).toggleStyle(.button)
            }
        }
        .task {
            do { reading = try await engine.reading(item.id) } catch { self.error = error.localizedDescription }
        }
    }

    private func line(_ i: Int, _ line: Reading.Line) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            if let speaker = line.speaker, !speaker.isEmpty {
                Text(speaker).font(.caption.bold()).foregroundStyle(Theme.slate)
            }
            FlowLayout(spacing: 0, lineSpacing: 6) {
                ForEach(Array(line.tokens.enumerated()), id: \.offset) { _, tok in
                    Text(tok.text)
                        .font(.system(size: 24))
                        .background(highlight(i, tok), in: .rect(cornerRadius: 4))
                        .foregroundStyle(marked.contains(tok.text) ? Theme.cinnabar : .primary)
                        .onTapGesture { if tok.word { selected = (i, tok); status = "" } }
                        .accessibilityAddTraits(tok.word ? .isButton : [])
                        .accessibilityIdentifier(tok.word ? "word" : "")
                }
            }
            if showEnglish, let en = line.english { Text(en).font(.callout).foregroundStyle(.secondary) }
        }
    }

    private func highlight(_ line: Int, _ tok: Reading.Token) -> Color {
        if let s = selected, s.line == line, s.token == tok { return Theme.ochre.opacity(0.35) }
        return .clear
    }

    private func gloss(_ tok: Reading.Token, in line: Reading.Line) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(alignment: .firstTextBaseline) {
                Text(tok.text).font(.system(size: 34))
                if let p = tok.pinyin { Pinyin.coloured(p).font(.title3.italic()) }
                Spacer()
                Button { Speaker.chinese(line.text) } label: { Image(systemName: "speaker.wave.2") }
                    .accessibilityLabel("Read the line aloud")
            }
            Text(tok.english ?? "Not in your decks or this text's word list yet.")
                .foregroundStyle(tok.english == nil ? .secondary : .primary)
            HStack {
                Button("I didn't know it") { Task { await mark(tok) } }
                    .disabled(tok.inDeck != true || marked.contains(tok.text))
                Button("Sentence card") { Task { await keep(tok, line) } }
                    .disabled(tok.english == nil)
            }
            .buttonStyle(.bordered)
            if !status.isEmpty { Text(status).font(.footnote).foregroundStyle(.secondary) }
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(.background.secondary)
    }

    private func mark(_ tok: Reading.Token) async {
        do {
            if try await engine.markUnknown(tok.text) {
                marked.insert(tok.text)
                status = "\(tok.text) will come back for review tomorrow."
            }
        } catch { status = error.localizedDescription }
    }

    private func keep(_ tok: Reading.Token, _ line: Reading.Line) async {
        do {
            let added = try await engine.addSentenceCard(tok.text, pinyin: tok.pinyin ?? "", english: tok.english ?? "", line: line.text)
            status = added ? "\(tok.text) joins your reviews with this sentence." : "\(tok.text) already has a sentence card."
        } catch { status = error.localizedDescription }
    }
}

/// Lays its children out in rows, wrapping like text.
struct FlowLayout: Layout {
    var spacing: CGFloat = 4
    var lineSpacing: CGFloat = 4

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let rows = arrange(width: proposal.width ?? .infinity, subviews)
        return CGSize(width: rows.map(\.width).max() ?? 0, height: rows.map(\.height).reduce(0, +) + lineSpacing * CGFloat(max(rows.count - 1, 0)))
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        var y = bounds.minY
        for row in arrange(width: bounds.width, subviews) {
            var x = bounds.minX
            for i in row.items {
                let size = subviews[i].sizeThatFits(.unspecified)
                subviews[i].place(at: CGPoint(x: x, y: y), proposal: ProposedViewSize(size))
                x += size.width + spacing
            }
            y += row.height + lineSpacing
        }
    }

    private struct Row { var items: [Int] = []; var width: CGFloat = 0; var height: CGFloat = 0 }

    private func arrange(width: CGFloat, _ subviews: Subviews) -> [Row] {
        var rows = [Row()]
        for i in subviews.indices {
            let size = subviews[i].sizeThatFits(.unspecified)
            if !rows[rows.count - 1].items.isEmpty, rows[rows.count - 1].width + spacing + size.width > width {
                rows.append(Row())
            }
            var row = rows[rows.count - 1]
            row.width += (row.items.isEmpty ? 0 : spacing) + size.width
            row.height = max(row.height, size.height)
            row.items.append(i)
            rows[rows.count - 1] = row
        }
        return rows
    }
}
