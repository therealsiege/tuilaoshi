import XCTest

/// The App Store screenshots, on a fresh install: camp, a study card with
/// its answer, the reader with a gloss open, the road, and settings. Each is
/// a kept attachment named for its file. Run it alone on the 6.9" iPhone
/// (iPhone 17 Pro Max) with -resultBundlePath, then copy the PNGs from
/// `xcrun xcresulttool export attachments` into ios/AppStore/screenshots/6.9.
@MainActor
final class ScreenshotTests: XCTestCase {
    override func setUp() { continueAfterFailure = false }

    func testStoreScreenshots() throws {
        var app = launch()
        save("1-camp")

        app.buttons["door.study"].tap()
        app.buttons["mode.review"].tap()
        XCTAssertTrue(app.buttons["Show answer"].waitForExistence(timeout: 10))
        app.buttons["Show answer"].tap()
        XCTAssertTrue(app.buttons["Good"].waitForExistence(timeout: 5))
        save("2-study-card")

        app = launch()
        app.buttons["door.read"].tap()
        let reading = app.cells.firstMatch
        XCTAssertTrue(reading.waitForExistence(timeout: 10), "there should be readings")
        reading.tap()
        let word = app.buttons.matching(identifier: "word").element(boundBy: 2)
        XCTAssertTrue(word.waitForExistence(timeout: 10))
        word.tap()
        XCTAssertTrue(app.buttons["I didn't know it"].waitForExistence(timeout: 5), "the gloss should open")
        save("3-reader-gloss")

        app = launch()
        let door = app.buttons["door.road"]
        for _ in 0..<6 where !door.isHittable { app.swipeUp() }
        door.tap()
        let position = app.staticTexts["road.position"]
        XCTAssertTrue(position.waitForExistence(timeout: 20), "the road should open")
        let start = position.label
        for _ in 0..<2 { app.buttons["road.west"].tap() }
        let moved = NSPredicate(format: "label != %@", start)
        wait(for: [XCTNSPredicateExpectation(predicate: moved, object: position)], timeout: 5)
        save("4-road")

        app = launch()
        app.buttons["Settings"].tap()
        XCTAssertTrue(app.switches["Read Chinese aloud"].waitForExistence(timeout: 5))
        save("5-settings")
    }

    /// A fresh install at camp, silent.
    private func launch() -> XCUIApplication {
        let app = XCUIApplication()
        app.launchArguments = ["-fresh-progress", "-quiet"]
        app.launch()
        XCTAssertTrue(app.staticTexts["camp.due"].waitForExistence(timeout: 30))
        return app
    }

    /// The whole screen, status bar included, at the device's native size.
    private func save(_ name: String) {
        let a = XCTAttachment(screenshot: XCUIScreen.main.screenshot(), quality: .original)
        a.name = name
        a.lifetime = .keepAlways
        add(a)
    }
}
