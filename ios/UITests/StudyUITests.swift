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

    private func shot(_ app: XCUIApplication, _ name: String) {
        let a = XCTAttachment(screenshot: app.screenshot())
        a.name = name
        a.lifetime = .keepAlways
        add(a)
    }
}
