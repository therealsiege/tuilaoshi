import SwiftUI

/// The teahouse storyteller (说书人): Journey to the West retold one chapter
/// a day by Claude, in the words the learner knows, as on the Mac. Told
/// chapters are kept with the progress, so those told on the Mac show here
/// too, and each opens in the reader word by word.
struct StorytellerView: View {
    @Environment(Engine.self) private var engine
    @Environment(Account.self) private var account
    @State private var shelf: StoryShelf?
    @State private var telling = false
    @State private var note: String?
    @State private var opened: ReadingItem?

    var body: some View {
        List {
            Section {
                if let shelf { tellRow(shelf) } else { ProgressView() }
                if telling {
                    Text("The storyteller clears his throat… 说书人清了清嗓子").font(.footnote).foregroundStyle(.secondary)
                }
                if let note { Text(note).font(.footnote).foregroundStyle(Theme.cinnabar) }
            }
            if let shelf, !shelf.chapters.isEmpty {
                Section("Told so far") {
                    ForEach(shelf.chapters, id: \.self) { ch in
                        NavigationLink(value: Self.item(ch)) {
                            VStack(alignment: .leading, spacing: 2) {
                                Text("第\(ch.number)回 \(ch.title)").font(.headline)
                                Text(ch.date).font(.caption).foregroundStyle(.secondary)
                            }
                        }
                    }
                }
            }
        }
        .navigationTitle("说书人 The Storyteller")
        .navigationDestination(item: $opened) { ReaderView(item: $0) }
        .task { await load() }
    }

    static func item(_ ch: StoryShelf.Chapter) -> ReadingItem {
        ReadingItem(id: ch.id, title: "第\(ch.number)回 \(ch.title)", tag: ch.date)
    }

    @ViewBuilder
    private func tellRow(_ shelf: StoryShelf) -> some View {
        if shelf.done {
            Text("The tale is told: all 100 chapters. Read them again any time.").foregroundStyle(.secondary)
        } else if !shelf.canTell {
            Text("第\(shelf.next)回 is told tomorrow. Come back to the teahouse then.").foregroundStyle(.secondary)
        } else if let claude = account.claude {
            Button("Tell today's chapter: 第\(shelf.next)回") { Task { await tell(via: claude) } }
                .disabled(telling)
        } else {
            Text(ClaudeTrouble.signedOut("A new chapter each day")).font(.footnote).foregroundStyle(.secondary)
        }
    }

    private func load() async {
        do { shelf = try await engine.storyteller() } catch { note = error.localizedDescription }
    }

    private func tell(via claude: ClaudeSession) async {
        telling = true
        note = nil
        defer { telling = false }
        do {
            let ch = try await engine.tellChapter(via: claude)
            await load()
            opened = Self.item(ch)
        } catch {
            note = "The storyteller lost his voice. " + ClaudeTrouble.explain(error)
        }
    }
}
