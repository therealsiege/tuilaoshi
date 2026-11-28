import XCTest
@testable import TuiLaoShi

/// Deleting the account, with the server call stubbed: a real one needs a
/// real Sign in with Apple.
@MainActor
final class AccountTests: XCTestCase {
    private var saved: String?

    override func setUp() async throws {
        saved = Keychain.read()
        Keychain.save("token-before")
    }

    override func tearDown() async throws {
        if let saved { Keychain.save(saved) } else { Keychain.delete() }
    }

    func testDeletingForgetsTheTokenAndSaysSo() async throws {
        let calls = Calls()
        let account = Account { server, token in await calls.add("\(server) \(token)") }
        XCTAssertTrue(account.signedIn)

        await account.deleteAccount()

        let seen = await calls.all
        XCTAssertEqual(seen, ["\(Account.server) token-before"])
        XCTAssertFalse(account.signedIn)
        XCTAssertTrue(account.deleted)
        XCTAssertNil(account.error)
        XCTAssertNil(Keychain.read(), "the Keychain token should be gone")
    }

    func testAFailedDeleteKeepsTheSignIn() async throws {
        struct Refused: LocalizedError { var errorDescription: String? { "server: try again" } }
        let account = Account { _, _ in throw Refused() }

        await account.deleteAccount()

        XCTAssertTrue(account.signedIn)
        XCTAssertFalse(account.deleted)
        XCTAssertEqual(account.error, "Couldn't delete the account: server: try again")
        XCTAssertEqual(Keychain.read(), "token-before")
    }
}

private actor Calls {
    var all: [String] = []
    func add(_ s: String) { all.append(s) }
}
