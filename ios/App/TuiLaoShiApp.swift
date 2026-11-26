import SwiftUI
import TuiLaoShiCore

/// The app's entry point. For now it shows that the Go engine runs on the
/// phone: it opens the bundled wiki and counts the cards due.
@main
struct TuiLaoShiApp: App {
    var body: some Scene {
        WindowGroup { CoreCheckView() }
    }
}

struct CoreCheckView: View {
    @State private var line = "Opening the wiki…"

    var body: some View {
        VStack(spacing: 12) {
            Text("推老师").font(.largeTitle)
            Text(line).foregroundStyle(.secondary)
        }
        .task { line = Self.check() }
    }

    static func check() -> String {
        guard let wiki = Bundle.main.url(forResource: "wiki", withExtension: nil) else {
            return "No wiki in the app yet"
        }
        let progress = URL.applicationSupportDirectory
        try? FileManager.default.createDirectory(at: progress, withIntermediateDirectories: true)
        var error: NSError?
        guard let core = MobileOpen(wiki.path, progress.path, &error) else {
            return "Couldn't open: \(error?.localizedDescription ?? "unknown")"
        }
        defer { try? core.close() }
        var due = 0
        do { try core.dueCount(&due) } catch { return "Couldn't count: \(error.localizedDescription)" }
        return "\(due) cards due"
    }
}
