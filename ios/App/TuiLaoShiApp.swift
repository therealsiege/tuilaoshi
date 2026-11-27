import SwiftUI
import TuiLaoShiCore

@main
struct TuiLaoShiApp: App {
    var body: some Scene {
        WindowGroup { CampView() }
    }
}

struct CampView: View {
    @State private var line = "Opening camp…"

    var body: some View {
        VStack(spacing: 12) {
            Text("推老师").font(.largeTitle)
            Text(line).foregroundStyle(.secondary)
            if line != "Opening camp…" { Text("Today · Bounties · Letter · Gazette · Diary · Progress · Settings").font(.footnote) }
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
        var campError: NSError?
        let json = core.camp(&campError)
        if let campError { return "Couldn't load camp: \(campError.localizedDescription)" }
        return json
    }
}
