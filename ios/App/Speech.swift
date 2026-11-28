import AVFoundation
import AudioToolbox
import Speech
import SwiftUI
import TuiLaoShiCore

/// The learner's settings that sound and display follow, as last read
/// from the engine (they live in the progress file and sync with the Mac).
@MainActor
enum SoundSettings {
    static var voices = !ProcessInfo.processInfo.arguments.contains("-quiet")
    static var cues = !ProcessInfo.processInfo.arguments.contains("-quiet")
    static var toneColours = true
}

/// Reads text aloud in the system's Mandarin or English voice, the phone's
/// stand-in for the Mac's say. Silent when voices are off.
@MainActor
enum Speaker {
    private static let synth = AVSpeechSynthesizer()

    static func chinese(_ text: String) { say(text, "zh-CN") }
    static func english(_ text: String) { say(text, "en-US") }

    private static func say(_ text: String, _ language: String) {
        guard SoundSettings.voices, !text.isEmpty else { return }
        let u = AVSpeechUtterance(string: text)
        u.voice = AVSpeechSynthesisVoice(language: language)
        u.rate = AVSpeechUtteranceDefaultSpeechRate * 0.9
        synth.stopSpeaking(at: .immediate)
        synth.speak(u)
    }
}

/// Short cues for game events, with the phone's own system sounds, as the
/// Mac uses its system sounds (internal/audio/sfx.go). Silent when cues are off.
enum Cue: SystemSoundID {
    case right = 1057
    case wrong = 1053
    case coins = 1103
    case levelUp = 1025

    @MainActor func play() {
        if SoundSettings.cues { AudioServicesPlaySystemSound(rawValue) }
    }
}

/// A shadowed line, scored character by character by the Go engine.
struct ShadowResult: Decodable, Equatable {
    struct Char: Decodable, Equatable {
        let char: String
        let heard: Bool
    }
    let chars: [Char]
    let percent: Int
}

/// Shadowing: the learner says a line back, Apple's speech recognition
/// hears it in Mandarin (on the device where it can), and the engine scores
/// it by character, as the Mac does with whisper.cpp.
@MainActor @Observable
final class Shadower {
    enum State: Equatable {
        case idle
        case listening
        case scored(heard: String, ShadowResult)
        case failed(String)
    }

    private(set) var state = State.idle
    private var heard = ""
    private let audio = AVAudioEngine()
    private var task: SFSpeechRecognitionTask?
    private var request: SFSpeechAudioBufferRecognitionRequest?
    private let recognizer = SFSpeechRecognizer(locale: Locale(identifier: "zh-CN"))

    /// Asks for the microphone and speech recognition, then listens.
    func start() async {
        let speech = await withCheckedContinuation { c in
            SFSpeechRecognizer.requestAuthorization { c.resume(returning: $0) }
        }
        guard speech == .authorized else {
            state = .failed("Shadowing needs speech recognition: allow it in Settings › 推老师.")
            return
        }
        guard await AVAudioApplication.requestRecordPermission() else {
            state = .failed("Shadowing needs the microphone: allow it in Settings › 推老师.")
            return
        }
        guard let recognizer, recognizer.isAvailable else {
            state = .failed("Mandarin speech recognition isn't available on this iPhone right now.")
            return
        }
        do {
            try listen(with: recognizer)
            state = .listening
        } catch {
            state = .failed("Couldn't start listening: \(error.localizedDescription)")
        }
    }

    private func listen(with recognizer: SFSpeechRecognizer) throws {
        heard = ""
        let session = AVAudioSession.sharedInstance()
        try session.setCategory(.playAndRecord, mode: .measurement, options: [.duckOthers, .defaultToSpeaker])
        try session.setActive(true, options: .notifyOthersOnDeactivation)
        let request = SFSpeechAudioBufferRecognitionRequest()
        request.shouldReportPartialResults = true
        request.requiresOnDeviceRecognition = recognizer.supportsOnDeviceRecognition
        self.request = request
        let input = audio.inputNode
        input.installTap(onBus: 0, bufferSize: 1024, format: input.outputFormat(forBus: 0)) { buffer, _ in
            request.append(buffer)
        }
        audio.prepare()
        try audio.start()
        task = recognizer.recognitionTask(with: request) { [weak self] result, _ in
            guard let text = result?.bestTranscription.formattedString else { return }
            Task { @MainActor in self?.heard = text }
        }
    }

    /// Stops listening and scores what was heard against line.
    func stop(scoring line: String) async {
        guard state == .listening else { return }
        audio.stop()
        audio.inputNode.removeTap(onBus: 0)
        request?.endAudio()
        // Give the recognizer a moment to settle its last words.
        try? await Task.sleep(for: .milliseconds(600))
        task?.cancel()
        task = nil
        request = nil
        try? AVAudioSession.sharedInstance().setActive(false, options: .notifyOthersOnDeactivation)
        do {
            state = .scored(heard: heard, try Self.score(line: line, heard: heard))
        } catch {
            state = .failed("Couldn't score that: \(error.localizedDescription)")
        }
    }

    func reset() { state = .idle }

    nonisolated static func score(line: String, heard: String) throws -> ShadowResult {
        var error: NSError?
        let json = MobileShadow(line, heard, &error)
        if let error { throw error }
        return try JSONDecoder().decode(ShadowResult.self, from: Data(json.utf8))
    }
}

/// Say a line back and see which characters were heard.
struct ShadowView: View {
    let line: String
    let pinyin: String
    @State private var shadower = Shadower()
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            VStack(spacing: 20) {
                Text(line).font(.system(size: 34)).multilineTextAlignment(.center)
                if !pinyin.isEmpty { Pinyin.coloured(pinyin).font(.title3.italic()) }
                Button { Speaker.chinese(line) } label: { Label("Hear it first", systemImage: "speaker.wave.2") }
                Spacer(minLength: 0)
                result
                Spacer(minLength: 0)
                button
            }
            .padding()
            .navigationTitle("跟读 Shadowing")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar { ToolbarItem(placement: .confirmationAction) { Button("Done") { dismiss() } } }
        }
    }

    @ViewBuilder
    private var result: some View {
        switch shadower.state {
        case .idle:
            Text("Tap Speak, say the line, then tap Stop.").foregroundStyle(.secondary)
        case .listening:
            Label("Listening…", systemImage: "waveform").foregroundStyle(Theme.cinnabar)
        case .scored(let heard, let r):
            VStack(spacing: 8) {
                r.chars.reduce(Text("")) { $0 + Text($1.char).foregroundStyle($1.heard ? Theme.jade : Theme.cinnabar) }
                    .font(.system(size: 34))
                Text("\(r.percent)% heard").font(.title3.bold())
                Text(heard.isEmpty ? "Nothing was heard." : "Heard: \(heard)").font(.footnote).foregroundStyle(.secondary)
            }
            .onAppear { (r.percent >= 80 ? Cue.right : Cue.wrong).play() }
        case .failed(let why):
            Text(why).foregroundStyle(Theme.cinnabar).multilineTextAlignment(.center)
        }
    }

    @ViewBuilder
    private var button: some View {
        if shadower.state == .listening {
            Button("Stop") { Task { await shadower.stop(scoring: line) } }
                .buttonStyle(.borderedProminent).tint(Theme.cinnabar).controlSize(.large)
        } else {
            Button { Task { shadower.reset(); await shadower.start() } } label: {
                Label("Speak", systemImage: "mic")
            }
            .buttonStyle(.borderedProminent).controlSize(.large)
        }
    }
}

/// The learner's settings, shared with the Mac: read aloud, sound cues,
/// tone colours, how many words a sitting takes, new words a day, battle
/// difficulty and how reviews are spaced.
struct SettingsView: View {
    @Environment(Engine.self) private var engine
    @State private var s: AppSettings?
    @State private var error: String?

    var body: some View {
        Form {
            if let s {
                let bind = { (kp: WritableKeyPath<AppSettings, Bool>) in
                    Binding(get: { s[keyPath: kp] }, set: { var n = s; n[keyPath: kp] = $0; save(n) })
                }
                Section("Sound and display") {
                    Toggle("Read Chinese aloud", isOn: bind(\.speech))
                    Toggle("Sound cues", isOn: bind(\.effects))
                    Toggle("Tone colours", isOn: bind(\.toneColours))
                }
                Section("Study") {
                    Stepper("Words a sitting: \(s.reviewChunk)", value: Binding(get: { s.reviewChunk }, set: { var n = s; n.reviewChunk = $0; save(n) }), in: 5...100, step: 5)
                    Stepper("New words a day: \(s.newWords)", value: Binding(get: { s.newWords }, set: { var n = s; n.newWords = $0; save(n) }), in: 0...50)
                    Picker("Review spacing", selection: Binding(get: { s.scheduler }, set: { var n = s; n.scheduler = $0; save(n) })) {
                        Text("Classic").tag("classic")
                        Text("FSRS").tag("fsrs")
                    }
                }
                Section("Battles") {
                    Picker("Difficulty", selection: Binding(get: { s.difficulty }, set: { var n = s; n.difficulty = $0; save(n) })) {
                        Text("Adaptive").tag("adaptive")
                        Text("Gentler").tag("gentler")
                        Text("Sterner").tag("sterner")
                    }
                }
                AccountSection()
                Section { Text("Settings travel with your progress, so the Mac game uses them too.").font(.footnote).foregroundStyle(.secondary) }
            } else if let error {
                Text(error).foregroundStyle(Theme.cinnabar)
            } else {
                ProgressView()
            }
        }
        .navigationTitle("设置 Settings")
        .task {
            do { s = try await engine.settings() } catch { self.error = error.localizedDescription }
        }
    }

    private func save(_ n: AppSettings) {
        s = n
        Task {
            do { try await engine.saveSettings(n) } catch { self.error = error.localizedDescription }
        }
    }
}
