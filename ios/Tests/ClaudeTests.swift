import Network
import XCTest
import TuiLaoShiCore
@testable import TuiLaoShi

/// A stand-in for the game's server's POST /api/claude on localhost: it
/// streams Claude's events for a fixed reply, as the server does, or
/// refuses with a status, and keeps the request bodies it was sent.
final class StubServer: @unchecked Sendable {
    private let listener: NWListener
    private let queue = DispatchQueue(label: "stub-server")
    private let lock = NSLock()
    private var _bodies: [String] = []
    private let reply: String
    private let status: Int

    var bodies: [String] { lock.withLock { _bodies } }
    private(set) var url = ""

    init(reply: String, status: Int = 200) throws {
        self.reply = reply
        self.status = status
        listener = try NWListener(using: .tcp, on: .any)
        let ready = DispatchSemaphore(value: 0)
        listener.stateUpdateHandler = { if case .ready = $0 { ready.signal() } }
        listener.newConnectionHandler = { [weak self] conn in self?.serve(conn) }
        listener.start(queue: queue)
        guard ready.wait(timeout: .now() + 10) == .success, let port = listener.port else {
            throw CocoaError(.featureUnsupported)
        }
        url = "http://127.0.0.1:\(port.rawValue)"
    }

    deinit { listener.cancel() }

    private func serve(_ conn: NWConnection) {
        conn.start(queue: queue)
        read(conn, Data())
    }

    /// Reads a whole request (headers, then Content-Length bytes), then answers.
    private func read(_ conn: NWConnection, _ got: Data) {
        conn.receive(minimumIncompleteLength: 1, maximumLength: 1 << 16) { [weak self] data, _, done, _ in
            guard let self else { return }
            var all = got
            if let data { all.append(data) }
            let text = String(decoding: all, as: UTF8.self)
            if let split = text.range(of: "\r\n\r\n") {
                let head = text[..<split.lowerBound].lowercased()
                let length = head.split(separator: "\r\n").first { $0.hasPrefix("content-length:") }
                    .flatMap { Int($0.dropFirst("content-length:".count).trimmingCharacters(in: .whitespaces)) } ?? 0
                let body = text[split.upperBound...]
                if body.utf8.count >= length {
                    self.lock.withLock { self._bodies.append(String(body)) }
                    self.answer(conn)
                    return
                }
            }
            if done { conn.cancel() } else { self.read(conn, all) }
        }
    }

    private func answer(_ conn: NWConnection) {
        var out: String
        if status != 200 {
            let body = #"{"error":"daily Claude limit reached: 30 calls a day; it resets at midnight UTC"}"#
            out = "HTTP/1.1 \(status) Refused\r\nContent-Type: application/json\r\nContent-Length: \(body.utf8.count)\r\nConnection: close\r\n\r\n" + body
        } else {
            let text = String(decoding: try! JSONEncoder().encode(reply), as: UTF8.self)
            let events = [
                ("message_start", #"{"type":"message_start","message":{"id":"msg_stub","type":"message","role":"assistant","model":"claude-stub","content":[],"stop_reason":null,"usage":{"input_tokens":5,"output_tokens":1}}}"#),
                ("content_block_start", #"{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}"#),
                ("content_block_delta", #"{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"# + text + "}}"),
                ("content_block_stop", #"{"type":"content_block_stop","index":0}"#),
                ("message_delta", #"{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":9}}"#),
                ("message_stop", #"{"type":"message_stop"}"#),
            ]
            out = "HTTP/1.1 200 OK\r\nContent-Type: text/event-stream; charset=utf-8\r\nConnection: close\r\n\r\n"
            for (type, data) in events { out += "event: \(type)\ndata: \(data)\n\n" }
        }
        conn.send(content: Data(out.utf8), contentContext: .finalMessage, isComplete: true, completion: .contentProcessed { _ in conn.cancel() })
    }
}

/// The Claude features, called from Swift the way the app calls them, with
/// the game's server stubbed on localhost.
final class ClaudeTests: XCTestCase {
    private func openCore() throws -> (MobileCore, CoreHandle) {
        let bundle = Bundle(for: ClaudeTests.self)
        let wiki = try XCTUnwrap(bundle.url(forResource: "Fixture", withExtension: nil))
        let progress = FileManager.default.temporaryDirectory.appendingPathComponent(UUID().uuidString)
        try FileManager.default.createDirectory(at: progress, withIntermediateDirectories: true)
        var error: NSError?
        let core = try XCTUnwrap(MobileOpen(wiki.path, progress.path, &error), "\(String(describing: error))")
        return (core, CoreHandle(core))
    }

    func testADiaryEntryIsGradedThroughTheServerAndKept() throws {
        let (core, h) = try openCore()
        defer { try? core.close() }
        let server = try StubServer(reply: "SCORE: 2\nFEEDBACK: 很好，很自然。")
        let grade = try Engine.writeDiary(h, via: ClaudeSession(server: server.url, token: "tok"), text: "今天我们走了很远。我很累，可是很高兴。")
        XCTAssertTrue(grade.kept)
        XCTAssertEqual(grade.score, 2)
        XCTAssertEqual(grade.feedback, "很好，很自然。")
        XCTAssertTrue(server.bodies.first?.contains("第1天") ?? false, "graded with the Mac's task")
        XCTAssertFalse(server.bodies.first?.contains("\"model\"") ?? true, "the server picks the model")

        var error: NSError?
        let json = core.diary(&error)
        XCTAssertNil(error)
        let entries = try JSONDecoder().decode([DiaryEntry].self, from: Data(json.utf8))
        XCTAssertEqual(entries.map(\.score), [2])
        XCTAssertEqual(entries.first?.date, DiaryView.today())
    }

    func testOverTheCapTheGazetteFallsBackToTheClerkAndSaysWhy() throws {
        let (core, h) = try openCore()
        defer { try? core.close() }
        let server = try StubServer(reply: "", status: 429)
        let issue = try Engine.claudeGazette(h, via: ClaudeSession(server: server.url, token: "tok"))
        XCTAssertEqual(issue.byHand, true)
        XCTAssertFalse(issue.text.isEmpty)
        let missed = try XCTUnwrap(issue.missed)
        XCTAssertTrue(ClaudeTrouble.explain(message: missed).contains("used up"), missed)
        XCTAssertTrue(WritingView.byHandNote(missed: missed).contains("used up"))
    }

    func testClaudeWritesTheGazetteWhenItCan() throws {
        let (core, h) = try openCore()
        defer { try? core.close() }
        let server = try StubServer(reply: "TEXT: 本报讯：一行继续西行。\nENGLISH: Our report: the party went on west.")
        let issue = try Engine.claudeGazette(h, via: ClaudeSession(server: server.url, token: "tok"))
        XCTAssertEqual(issue.text, "本报讯：一行继续西行。")
        XCTAssertNotEqual(issue.byHand, true)
        XCTAssertNil(issue.missed)
    }

    func testTheStorytellerTellsAChapterToRead() throws {
        let (core, h) = try openCore()
        defer { try? core.close() }
        let server = try StubServer(reply: "TITLE: 石猴出世\nTEXT: 花果山上有一块石头。石头里出来一只猴子。\nENGLISH: A monkey is born.\nQ: 猴子从哪里来？ | 石头 | 水 | 天")
        let ch = try Engine.tellChapter(h, via: ClaudeSession(server: server.url, token: "tok"))
        XCTAssertEqual(ch.id, "story:1")
        var error: NSError?
        let shelf = try JSONDecoder().decode(StoryShelf.self, from: Data(core.storyteller(&error).utf8))
        XCTAssertEqual(shelf.chapters.map(\.number), [1])
        XCTAssertFalse(shelf.canTell, "one chapter a day")
        let reading = try JSONDecoder().decode(Reading.self, from: Data(core.reading(ch.id, error: &error).utf8))
        XCTAssertNil(error)
        XCTAssertEqual(reading.title, "说书人 第1回 石猴出世")
        XCTAssertEqual(reading.lines.count, 3)
    }

    func testClaudeGradesAnAnswerThatIsntExact() throws {
        let (core, h) = try openCore()
        defer { try? core.close() }
        let server = try StubServer(reply: "CORRECT: true\nSCORE: 5\nFEEDBACK: 对！")
        let grade = try Engine.gradeAnswer(h, via: ClaudeSession(server: server.url, token: "tok"), card: "characters/shu1.md", answer: "書", mode: "production")
        XCTAssertTrue(grade.correct)
        XCTAssertEqual(grade.rating, 3)
        XCTAssertTrue(server.bodies.first?.contains("書") ?? false)
    }

    func testTroubleIsExplained() {
        XCTAssertTrue(ClaudeTrouble.explain(message: "today's Claude limit on the game's server is reached (x)").contains("midnight UTC"))
        XCTAssertTrue(ClaudeTrouble.explain(message: "the game's server didn't accept this device's sign-in; sign in again").contains("Settings"))
        XCTAssertTrue(ClaudeTrouble.explain(message: "boom").hasPrefix("Couldn't reach Claude"))
    }
}
