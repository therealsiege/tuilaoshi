import XCTest

/// A review session driven by touch, end to end, on a fresh install:
/// answer one card, leave early, and find the answer kept at camp.
@MainActor
final class StudyUITests: XCTestCase {
    func testAReviewLeftEarlyKeepsItsAnswer() throws {
        continueAfterFailure = false
        let app = XCUIApplication()
        app.launchArguments = ["-fresh-progress"]
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
        app.launchArguments = ["-fresh-progress"]
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
        app.launchArguments = ["-fresh-progress"]
        app.launch()
        XCTAssertTrue(app.staticTexts["camp.due"].waitForExistence(timeout: 30))
        XCTAssertTrue(app.buttons["Read camp aloud"].exists)
        app.buttons["Settings"].tap()
        let voices = app.switches["Read Chinese aloud"]
        XCTAssertTrue(voices.waitForExistence(timeout: 5))
        if voices.value as? String == "1" { voices.switches.firstMatch.tap() }
        XCTAssertEqual(voices.value as? String, "0")
        app.terminate()
        app.launchArguments = [] // keep the progress file, where settings live
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
        app.launchArguments = ["-fresh-progress"]
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
        app.launchArguments = ["-fresh-progress"]
        app.launch()
        XCTAssertTrue(app.staticTexts["camp.due"].waitForExistence(timeout: 30))
        let checks: [(String, String)] = [("letter", "Nothing yet"), ("gazette", "Week of"), ("diary", "Writing an entry"), ("progress", "The last seven days")]
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

    private func shot(_ app: XCUIApplication, _ name: String) {
        let a = XCTAttachment(screenshot: app.screenshot())
        a.name = name
        a.lifetime = .keepAlways
        add(a)
    }
}
