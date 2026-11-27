import SwiftUI

/// The app's entry point: it opens the journey from the bundled wiki and
/// starts at camp, the hub every other screen is reached from.
@main
struct TuiLaoShiApp: App {
    @State private var engine = Engine()
    @Environment(\.scenePhase) private var phase

    var body: some Scene {
        WindowGroup {
            NavigationStack {
                CampView()
            }
            .environment(engine)
            .tint(Theme.cinnabar)
            .task { await engine.open() }
            .onChange(of: phase) { _, now in
                // Write progress out for the Mac whenever the app is put away.
                if now == .background { Task { await engine.syncNow() } }
            }
        }
    }
}
