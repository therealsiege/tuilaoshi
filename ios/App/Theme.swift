import SwiftUI
import UIKit

/// The app's look: Songti for headings, PingFang (the system font) for
/// text, and the art bible's palette of jade, ochre, slate and cinnabar.
enum Theme {
    static let jade = Color(red: 0.24, green: 0.52, blue: 0.43)
    static let ochre = Color(red: 0.78, green: 0.58, blue: 0.27)
    static let slate = Color(red: 0.33, green: 0.42, blue: 0.53)
    static let cinnabar = Color(red: 0.80, green: 0.27, blue: 0.20)

    /// A heading in Songti, falling back to the system serif where Songti
    /// isn't installed.
    static func heading(_ size: CGFloat) -> Font {
        for name in ["STSongti-SC-Bold", "Songti SC Bold", "STSong"] where UIFont(name: name, size: size) != nil {
            return .custom(name, size: size, relativeTo: .title)
        }
        return .system(size: size, weight: .bold, design: .serif)
    }
}
