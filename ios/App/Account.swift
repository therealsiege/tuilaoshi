import AuthenticationServices
import Security
import SwiftUI
import TuiLaoShiCore

/// The learner's account with the game's server: signed in with Apple,
/// the session token kept in the Keychain. Sync, pairing the Mac and the
/// Claude features go through it.
@MainActor @Observable
final class Account {
    /// The server, from the app's Info.plist (TLSServer); empty until it is deployed.
    nonisolated static let server = (Bundle.main.object(forInfoDictionaryKey: "TLSServer") as? String ?? "")
        .trimmingCharacters(in: .whitespaces)

    /// Deletes the account on a server, given the server and the token.
    typealias Remover = @Sendable (_ server: String, _ token: String) async throws -> Void

    private(set) var token: String?
    var error: String?
    /// True once the account was deleted from this iPhone, until the next sign-in.
    private(set) var deleted = false
    private(set) var deleting = false
    private let remove: Remover

    var signedIn: Bool { token != nil }

    /// Tests pass a remover that doesn't call the server.
    init(remove: @escaping Remover = Account.removeOnServer) {
        token = Keychain.read()
        self.remove = remove
    }

    nonisolated static let removeOnServer: Remover = { server, token in
        try await Task.detached {
            var e: NSError?
            _ = MobileDeleteAccount(server, token, &e)
            if let e { throw e }
        }.value
    }

    /// Trades Sign in with Apple's identity token for the server's session.
    func signIn(with result: Result<ASAuthorization, any Error>) async {
        error = nil
        guard !Self.server.isEmpty else {
            error = "The game's server isn't set up yet."
            return
        }
        switch result {
        case .failure(let e):
            if (e as? ASAuthorizationError)?.code != .canceled { error = e.localizedDescription }
        case .success(let auth):
            guard let cred = auth.credential as? ASAuthorizationAppleIDCredential,
                  let data = cred.identityToken, let identity = String(data: data, encoding: .utf8) else {
                error = "Apple didn't return an identity token."
                return
            }
            do {
                let server = Self.server
                let json = try await Task.detached { () throws -> String in
                    var e: NSError?
                    let s = MobileSignInWithApple(server, identity, Engine.syncName, &e)
                    if let e { throw e }
                    return s
                }.value
                struct Session: Decodable { let token: String }
                let s = try JSONDecoder().decode(Session.self, from: Data(json.utf8))
                Keychain.save(s.token)
                token = s.token
                deleted = false
            } catch {
                self.error = error.localizedDescription
            }
        }
    }

    func signOut() {
        Keychain.delete()
        token = nil
    }

    /// Deletes the account and the progress synced to the server, for every
    /// device, then forgets the token here. Progress on this iPhone stays.
    func deleteAccount() async {
        guard let token, !deleting else { return }
        error = nil
        deleting = true
        defer { deleting = false }
        do {
            try await remove(Self.server, token)
            Keychain.delete()
            self.token = nil
            deleted = true
        } catch {
            self.error = "Couldn't delete the account: \(error.localizedDescription)"
        }
    }

    /// A code the Mac claims with `tuilaoshi -pair CODE`.
    func pairingCode() async throws -> String {
        guard let token else { throw CocoaError(.userCancelled) }
        let server = Self.server
        return try await Task.detached { () throws -> String in
            var e: NSError?
            let code = MobilePairingCode(server, token, &e)
            if let e { throw e }
            return code
        }.value
    }
}

/// The session token in the Keychain, this device only.
enum Keychain {
    nonisolated(unsafe) private static let query: [String: Any] = [
        kSecClass as String: kSecClassGenericPassword,
        kSecAttrService as String: "com.therealsiege.tuilaoshi.session",
        kSecAttrAccount as String: "server",
    ]

    static func read() -> String? {
        var q = query
        q[kSecReturnData as String] = true
        var out: AnyObject?
        guard SecItemCopyMatching(q as CFDictionary, &out) == errSecSuccess, let data = out as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    static func save(_ token: String) {
        delete()
        var q = query
        q[kSecValueData as String] = Data(token.utf8)
        q[kSecAttrAccessible as String] = kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly
        SecItemAdd(q as CFDictionary, nil)
    }

    static func delete() { SecItemDelete(query as CFDictionary) }
}

/// Signing in, pairing the Mac, and deleting the account, in Settings.
struct AccountSection: View {
    @Environment(Account.self) private var account
    @Environment(Engine.self) private var engine
    @State private var code: String?
    @State private var confirmDelete = false

    var body: some View {
        Section("Account and sync") {
            if Account.server.isEmpty {
                Text("Sync and Claude come with the game's server, which isn't set up yet. Progress is kept on this iPhone.")
                    .font(.footnote).foregroundStyle(.secondary)
            } else if account.signedIn {
                Label("Signed in", systemImage: "checkmark.seal")
                Button("Sync now") { Task { await engine.syncNow() } }
                Button("Pair a Mac") { Task { code = try? await account.pairingCode() } }
                if let code {
                    Text("On the Mac, run:  tuilaoshi -pair \(code)").font(.callout.monospaced()).textSelection(.enabled)
                    Text("The code works once, for 10 minutes.").font(.caption).foregroundStyle(.secondary)
                }
                Button("Sign out", role: .destructive) { account.signOut() }
                Button("Delete account", role: .destructive) { confirmDelete = true }
                    .disabled(account.deleting)
                    .confirmationDialog("Delete your account?", isPresented: $confirmDelete, titleVisibility: .visible) {
                        Button("Delete account", role: .destructive) {
                            Task {
                                await account.deleteAccount()
                                code = nil
                                await engine.syncNow()
                            }
                        }
                        Button("Cancel", role: .cancel) {}
                    } message: {
                        Text("This deletes your account and the progress synced to the server, for every device, and can't be undone. The progress on this iPhone stays.")
                    }
            } else {
                if account.deleted {
                    Label("Account deleted", systemImage: "trash")
                    Text("Your account and the progress synced to the server are gone. The progress on this iPhone stays; sign in again to start a new account.")
                        .font(.footnote).foregroundStyle(.secondary)
                }
                SignInWithAppleButton(.signIn) { $0.requestedScopes = [.email] } onCompletion: { result in
                    Task {
                        await account.signIn(with: result)
                        await engine.syncNow()
                    }
                }
                .frame(height: 44)
            }
            if let e = account.error { Text(e).font(.footnote).foregroundStyle(Theme.cinnabar) }
        }
    }
}
