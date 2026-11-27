import XCTest
import TuiLaoShiCore
@testable import TuiLaoShi

/// The Go engine, called from Swift in the iOS simulator: open a wiki,
/// list due cards, record a review, and see the progress kept.
final class CoreTests: XCTestCase {
    func testTheEngineReviewsACardOnIOS() throws {
        let bundle = Bundle(for: CoreTests.self)
        let wiki = try XCTUnwrap(bundle.url(forResource: "Fixture", withExtension: nil))
        let progress = FileManager.default.temporaryDirectory.appendingPathComponent(UUID().uuidString)
        try FileManager.default.createDirectory(at: progress, withIntermediateDirectories: true)

        var error: NSError?
        let core = try XCTUnwrap(MobileOpen(wiki.path, progress.path, &error), "\(String(describing: error))")
        defer { try? core.close() }

        var due = 0
        try core.dueCount(&due)
        XCTAssertEqual(due, 1)

        let list = core.dueCards(10, error: &error)
        XCTAssertNil(error)
        struct Card: Decodable { let id: String; let chinese: String }
        let cards = try JSONDecoder().decode([Card].self, from: Data(list.utf8))
        XCTAssertEqual(cards.first?.chinese, "书")

        _ = core.review(cards[0].id, rating: 2, error: &error)
        XCTAssertNil(error)
        try core.dueCount(&due)
        XCTAssertEqual(due, 0)
        XCTAssertTrue(FileManager.default.fileExists(atPath: progress.appendingPathComponent("progress.db").path))
    }
}

/// The journey's content, embedded in the Go engine, loads on iOS.
final class ContentTests: XCTestCase {
    func testTheEmbeddedContentLoads() throws {
        var error: NSError?
        let today = MobileToday(&error)
        XCTAssertNil(error)
        struct Today: Decodable { let term: String; let tribulations: Int }
        let t = try JSONDecoder().decode(Today.self, from: Data(today.utf8))
        XCTAssertEqual(t.tribulations, 81)
        XCTAssertFalse(t.term.isEmpty)
    }
}

/// Camp's numbers come back from Go in the shape the app reads.
final class CampTests: XCTestCase {
    func testCampDecodesForANewPilgrim() async throws {
        let bundle = Bundle(for: CampTests.self)
        let wiki = try XCTUnwrap(bundle.url(forResource: "Fixture", withExtension: nil))
        let progress = FileManager.default.temporaryDirectory.appendingPathComponent(UUID().uuidString)
        try FileManager.default.createDirectory(at: progress, withIntermediateDirectories: true)
        var error: NSError?
        let core = try XCTUnwrap(MobileOpen(wiki.path, progress.path, &error), "\(String(describing: error))")
        defer { try? core.close() }

        let camp = try await Engine.camp(CoreHandle(core))
        XCTAssertEqual(camp.due, 1)
        XCTAssertEqual(camp.party, ["唐僧"])
        XCTAssertEqual(camp.stage?.number, 1)
        XCTAssertFalse(camp.term.isEmpty)
    }
}
