import XCTest

/// A review session driven by touch, end to end, on a fresh install:
/// answer one card, leave early, and find the answer kept at camp.
@MainActor
final class StudyUITests: XCTestCase {
    func testAReviewLeftEarlyKeepsItsAnswer() throws {
        continueAfterFailure = false
        let app = XCUIApplication()
        app.launchArguments = ["-fresh-progress", "-quiet"]
        app.launch()

        let due = app.staticTexts["camp.due"]
        XCTAssertTrue(due.waitForExistence(timeout: 30))
        let before = Int(due.label) ?? -1
        XCTAssertGreaterThan(before, 0)
        shot(app, "camp")

        app.buttons["door.study"].tap()
        app.buttons["mode.review"].tap()
        XCTAssertTrue(app.buttons["Show answer"].waitForExistence(timeout: 10))
        shot(app, "card")
        app.buttons["Show answer"].tap()
        shot(app, "answer")
        app.buttons["Good"].tap()
        XCTAssertTrue(app.buttons["Show answer"].waitForExistence(timeout: 5), "the next card should come up")

        // Leave early: back to study, back to camp.
        app.navigationBars.buttons.element(boundBy: 0).tap()
        app.navigationBars.buttons.element(boundBy: 0).tap()
        let fewer = NSPredicate(format: "label == %@", String(before - 1))
        wait(for: [XCTNSPredicateExpectation(predicate: fewer, object: due)], timeout: 10)
        XCTAssertEqual(app.staticTexts["camp.streak"].label, "1")
        shot(app, "camp-after")
    }

    /// One card in each flashcard mode, answered by touch.
    func testEveryFlashcardModeAnswersACard() throws {
        continueAfterFailure = false
        let app = XCUIApplication()
        app.launchArguments = ["-fresh-progress", "-quiet"]
        app.launch()
        XCTAssertTrue(app.staticTexts["camp.due"].waitForExistence(timeout: 30))
        app.buttons["door.study"].tap()
        for mode in ["recognition", "production", "listening", "sentence"] {
            app.buttons["mode.\(mode)"].tap()
            let reveal = mode == "production" ? app.buttons["Check"] : app.buttons["Show answer"]
            if mode == "production" {
                XCTAssertTrue(app.textFields.firstMatch.waitForExistence(timeout: 10))
                app.textFields.firstMatch.tap()
                app.textFields.firstMatch.typeText("书\n")
            } else if !reveal.waitForExistence(timeout: 10) {
                // Sentence mode needs a due card with an example sentence.
                XCTAssertEqual(mode, "sentence", "\(mode) should show a card")
                XCTAssertTrue(app.staticTexts["Nothing due"].exists)
                app.navigationBars.buttons.element(boundBy: 0).tap()
                continue
            } else {
                reveal.tap()
            }
            XCTAssertTrue(app.buttons["Good"].waitForExistence(timeout: 5), "\(mode): rating buttons after the answer")
            shot(app, mode)
            app.buttons["Good"].tap()
            app.navigationBars.buttons.element(boundBy: 0).tap()
        }
    }

    /// Settings turn the voice off, and it stays off: settings live in the
    /// progress file, which also syncs them with the Mac.
    func testVoicesCanBeTurnedOff() throws {
        continueAfterFailure = false
        let app = XCUIApplication()
        app.launchArguments = ["-fresh-progress", "-quiet"]
        app.launch()
        XCTAssertTrue(app.staticTexts["camp.due"].waitForExistence(timeout: 30))
        XCTAssertTrue(app.buttons["Read camp aloud"].exists)
        app.buttons["Settings"].tap()
        let voices = app.switches["Read Chinese aloud"]
        XCTAssertTrue(voices.waitForExistence(timeout: 5))
        if voices.value as? String == "1" { voices.switches.firstMatch.tap() }
        XCTAssertEqual(voices.value as? String, "0")
        app.terminate()
        app.launchArguments = ["-quiet"] // keep the progress file, where settings live
        app.launch()
        XCTAssertTrue(app.staticTexts["camp.due"].waitForExistence(timeout: 30))
        app.buttons["Settings"].tap()
        XCTAssertTrue(app.switches["Read Chinese aloud"].waitForExistence(timeout: 5))
        XCTAssertEqual(app.switches["Read Chinese aloud"].value as? String, "0", "the setting should be kept")
        app.switches["Read Chinese aloud"].switches.firstMatch.tap()
    }

    /// A reading opens, a tapped word shows its gloss, and it can become a
    /// sentence card.
    func testTappingAWordGlossesIt() throws {
        continueAfterFailure = false
        let app = XCUIApplication()
        app.launchArguments = ["-fresh-progress", "-quiet"]
        app.launch()
        XCTAssertTrue(app.staticTexts["camp.due"].waitForExistence(timeout: 30))
        app.buttons["door.read"].tap()
        let first = app.cells.firstMatch
        XCTAssertTrue(first.waitForExistence(timeout: 10), "there should be readings")
        first.tap()
        let word = app.buttons.matching(identifier: "word").element(boundBy: 2)
        XCTAssertTrue(word.waitForExistence(timeout: 10))
        word.tap()
        XCTAssertTrue(app.buttons["I didn't know it"].waitForExistence(timeout: 5), "the gloss should open")
        shot(app, "gloss")
        let card = app.buttons["Sentence card"]
        if card.isEnabled {
            card.tap()
            XCTAssertTrue(app.staticTexts.containing(NSPredicate(format: "label CONTAINS 'sentence'")).firstMatch.waitForExistence(timeout: 5))
        }
    }

    /// Every camp screen opens: the letter (none yet for a new pilgrim),
    /// the gazette set by the clerk, the diary, and My progress.
    func testCampsPapersOpen() throws {
        continueAfterFailure = false
        let app = XCUIApplication()
        app.launchArguments = ["-fresh-progress", "-quiet"]
        app.launch()
        XCTAssertTrue(app.staticTexts["camp.due"].waitForExistence(timeout: 30))
        let checks: [(String, String)] = [("letter", "Nothing yet"), ("gazette", "Week of"), ("diary", "Writing an entry"), ("progress", "The last seven days"), ("web", "No web yet")]
        for (paper, expect) in checks {
            let link = app.buttons["paper.\(paper)"]
            for _ in 0..<6 where !link.isHittable { app.swipeUp() }
            link.tap()
            let found = app.staticTexts.containing(NSPredicate(format: "label CONTAINS %@", expect)).firstMatch
            XCTAssertTrue(found.waitForExistence(timeout: 10), "\(paper) should show \(expect)")
            shot(app, paper)
            app.navigationBars.buttons.element(boundBy: 0).tap()
        }
    }

    /// The road opens on a fresh journey and is walked by the direction
    /// pad: west along the road and back east, which changes where the
    /// pilgrim stands, then east once more into the first region's edge,
    /// which is blocked and says so.
    func testTheRoadIsWalkedAStepAtATime() throws {
        continueAfterFailure = false
        let app = XCUIApplication()
        app.launchArguments = ["-fresh-progress", "-quiet"]
        app.launch()
        XCTAssertTrue(app.staticTexts["camp.due"].waitForExistence(timeout: 30))
        let door = app.buttons["door.road"]
        for _ in 0..<6 where !door.isHittable { app.swipeUp() }
        door.tap()
        let position = app.staticTexts["road.position"]
        XCTAssertTrue(position.waitForExistence(timeout: 20), "the road should open")
        let start = position.label
        for _ in 0..<3 { app.buttons["road.west"].tap() }
        let moved = NSPredicate(format: "label != %@", start)
        wait(for: [XCTNSPredicateExpectation(predicate: moved, object: position)], timeout: 5)
        let west = position.label
        for _ in 0..<3 { app.buttons["road.east"].tap() }
        let back = NSPredicate(format: "label != %@", west)
        wait(for: [XCTNSPredicateExpectation(predicate: back, object: position)], timeout: 5)
        XCTAssertEqual(position.label, start, "three steps east should undo three steps west")
        meet(app) // the sixth step ends the day, which may bring an event or demons
        let banner = app.descendants(matching: .any)["road.banner"]
        XCTAssertFalse(banner.exists, "plain steps show no banner")
        app.buttons["road.east"].tap()
        XCTAssertTrue(banner.waitForExistence(timeout: 5), "a blocked step should say so")
        XCTAssertTrue(banner.label.contains("blocked"), "banner: \(banner.label)")
        XCTAssertEqual(position.label, start, "a blocked step doesn't move the pilgrim")
        shot(app, "road")
    }

    /// From a fresh install, the first region's river is crossed: walk the
    /// road west from the start to the crossing, meeting whatever the days
    /// bring on the way, choose a way across in its sheet, and walk on to
    /// stand on the far bank, then on along the road.
    func testTheFirstRegionsRiverIsCrossed() throws {
        continueAfterFailure = false
        let app = XCUIApplication()
        app.launchArguments = ["-fresh-progress", "-quiet"]
        app.launch()
        XCTAssertTrue(app.staticTexts["camp.due"].waitForExistence(timeout: 30))
        let door = app.buttons["door.road"]
        for _ in 0..<6 where !door.isHittable { app.swipeUp() }
        door.tap()
        let position = app.staticTexts["road.position"]
        XCTAssertTrue(position.waitForExistence(timeout: 20), "the road should open")
        XCTAssertEqual(position.label, "Position 44, 6", "a new journey starts at the first region's east end")
        let trail = app.descendants(matching: .any)["road.trail"]
        XCTAssertTrue(trail.waitForExistence(timeout: 5), "the trail should show")
        XCTAssertTrue(trail.label.contains("Day 1"), "trail: \(trail.label)")

        // West along the road to the river at 38, 4.
        let route = ["west", "west", "west", "north", "west", "north", "west"]
        for dir in route {
            let before = position.label
            app.buttons["road.\(dir)"].tap()
            let moved = NSPredicate(format: "label != %@", before)
            wait(for: [XCTNSPredicateExpectation(predicate: moved, object: position)], timeout: 5)
            meet(app)
        }
        XCTAssertEqual(position.label, "Position 39, 4", "beside the river")
        app.buttons["road.west"].tap()
        let crossing = app.descendants(matching: .any)["event.crossing"]
        XCTAssertTrue(app.buttons["event.choice.0"].waitForExistence(timeout: 5), "the river should open its crossing")
        XCTAssertTrue(crossing.exists, "the sheet should be the crossing")
        XCTAssertEqual(position.label, "Position 39, 4", "the river bars the way until crossed")
        shot(app, "crossing")
        app.buttons["event.choice.0"].tap()
        let walkOn = app.buttons["event.walkOn"]
        XCTAssertTrue(walkOn.waitForExistence(timeout: 5), "a choice made, walk on")
        shot(app, "crossing-chosen")
        walkOn.tap()
        let across = NSPredicate(format: "label == %@", "Position 38, 4")
        wait(for: [XCTNSPredicateExpectation(predicate: across, object: position)], timeout: 5)
        let banner = app.descendants(matching: .any)["road.banner"]
        XCTAssertTrue(banner.waitForExistence(timeout: 5) && banner.label.contains("across"), "banner: \(banner.label)")

        app.buttons["road.west"].tap()
        let on = NSPredicate(format: "label == %@", "Position 37, 4")
        wait(for: [XCTNSPredicateExpectation(predicate: on, object: position)], timeout: 5)
        meet(app)
        shot(app, "across")
    }

    /// Meets whatever sheet a step brought, if any, by its first choice,
    /// then walks on.
    private func meet(_ app: XCUIApplication) {
        for kind in ["event", "plot"] where app.buttons["\(kind).choice.0"].waitForExistence(timeout: 1) {
            app.buttons["\(kind).choice.0"].tap()
            let walkOn = app.buttons["\(kind).walkOn"]
            XCTAssertTrue(walkOn.waitForExistence(timeout: 5), "\(kind): walk on after choosing")
            walkOn.tap()
            XCTAssertTrue(app.buttons["road.west"].waitForExistence(timeout: 5))
        }
    }

    private func shot(_ app: XCUIApplication, _ name: String) {
        let a = XCTAttachment(screenshot: app.screenshot())
        a.name = name
        a.lifetime = .keepAlways
        add(a)
    }
}
