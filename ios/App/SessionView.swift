import SwiftUI

/// A session: one card at a time. Swipe right for good, left for again,
/// or use the buttons, which add hard and easy.
struct SessionView: View {
    @State var session: StudySession
    @State private var drag = CGSize.zero
    @FocusState private var typing: Bool

    var body: some View {
        VStack(spacing: 16) {
            if session.loading {
                ProgressView("Finding due cards…")
            } else if let error = session.error {
                ContentUnavailableView("Something went wrong", systemImage: "exclamationmark.triangle", description: Text(error))
            } else if let card = session.card {
                progress
                face(card)
                    .offset(x: drag.width)
                    .rotationEffect(.degrees(drag.width / 25))
                    .gesture(swipe, including: session.revealed ? .all : .subviews)
                Spacer(minLength: 0)
                controls(card)
            } else {
                summary
            }
        }
        .padding()
        .navigationTitle(session.mode.title)
        .navigationBarTitleDisplayMode(.inline)
        .sensoryFeedback(.selection, trigger: session.index)
        .task { await session.load() }
        .onDisappear { Task { await session.finish() } }
    }

    private var progress: some View {
        ProgressView(value: Double(session.index), total: Double(max(session.cards.count, 1))) {
            Text("\(session.index + 1) of \(session.cards.count)").font(.caption).foregroundStyle(.secondary)
        }
    }

    @ViewBuilder
    private func face(_ card: Card) -> some View {
        VStack(spacing: 12) {
            switch session.mode {
            case .review, .recognition:
                Text(card.chinese).font(.system(size: 64))
                if session.revealed { answer(card, showChinese: false) }
            case .production:
                Text(card.english).font(.title2).multilineTextAlignment(.center)
                TextField("Type the Chinese", text: $session.typed)
                    .font(.title)
                    .multilineTextAlignment(.center)
                    .textFieldStyle(.roundedBorder)
                    .focused($typing)
                    .submitLabel(.done)
                    .onSubmit { session.revealed = true }
                    .disabled(session.revealed)
                if session.revealed {
                    Label(session.typedRight ? "Exactly right" : "Not quite", systemImage: session.typedRight ? "checkmark.circle" : "xmark.circle")
                        .foregroundStyle(session.typedRight ? Theme.jade : Theme.cinnabar)
                    answer(card, showChinese: true)
                }
            case .listening:
                Button { Voice.say(card.chinese) } label: {
                    Label("Play again", systemImage: "speaker.wave.2").font(.title2)
                }
                .onAppear { Voice.say(card.chinese) }
                if session.revealed { answer(card, showChinese: true) }
            case .sentence:
                Text(card.cloze ?? "").font(.title2).multilineTextAlignment(.center)
                Text(card.english).font(.subheadline).foregroundStyle(.secondary)
                if session.revealed {
                    Text(card.sentence ?? "").font(.title3)
                    answer(card, showChinese: true)
                }
            }
        }
        .frame(maxWidth: .infinity, minHeight: 260)
        .padding()
        .background(.background.secondary, in: .rect(cornerRadius: 20))
        .contentShape(.rect)
        .onTapGesture { if session.mode != .production { session.revealed = true } }
        .id(card.id)
    }

    private func answer(_ card: Card, showChinese: Bool) -> some View {
        VStack(spacing: 6) {
            if showChinese { Text(card.chinese).font(.system(size: 48)) }
            Pinyin.coloured(card.pinyin).font(.title2.italic())
            Text(card.english).font(.title3).multilineTextAlignment(.center)
            if let ex = card.examples?.first, session.mode != .sentence {
                Text(ex.hasPrefix("- ") ? String(ex.dropFirst(2)) : ex).font(.callout).foregroundStyle(.secondary).multilineTextAlignment(.center)
            }
        }
    }

    @ViewBuilder
    private func controls(_ card: Card) -> some View {
        if session.revealed {
            HStack(spacing: 8) {
                rating("Again", 0, Theme.cinnabar)
                rating("Hard", 1, Theme.ochre)
                rating("Good", 2, Theme.jade)
                rating("Easy", 3, Theme.slate)
            }
            Text("Or swipe: right for good, left for again").font(.caption).foregroundStyle(.secondary)
        } else if session.mode == .production {
            Button("Check") { session.revealed = true; typing = false }
                .buttonStyle(.borderedProminent).controlSize(.large)
        } else {
            Button("Show answer") { session.revealed = true }
                .buttonStyle(.borderedProminent).controlSize(.large)
        }
    }

    private func rating(_ title: String, _ value: Int, _ tint: Color) -> some View {
        Button { Task { await session.rate(value) } } label: {
            Text(title).lineLimit(1).minimumScaleFactor(0.8).frame(maxWidth: .infinity)
        }
        .buttonStyle(.bordered).tint(tint).controlSize(.large)
        .accessibilityLabel(title)
    }

    private var swipe: some Gesture {
        DragGesture()
            .onChanged { drag = $0.translation }
            .onEnded { value in
                let dx = value.translation.width
                drag = .zero
                guard session.revealed, abs(dx) > 100 else { return }
                Task { await session.rate(dx > 0 ? 2 : 0) }
            }
    }

    private var summary: some View {
        let right = session.answers.filter(\.correct).count
        return ContentUnavailableView {
            Label(session.answers.isEmpty ? "Nothing due" : "Session done", systemImage: "checkmark.seal")
        } description: {
            Text(session.answers.isEmpty
                 ? "No cards are due for this mode right now."
                 : "\(right) of \(session.answers.count) remembered. They're saved, and camp is updated.")
        }
    }
}
