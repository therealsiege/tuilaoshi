import SwiftUI

/// Pinyin drawn a syllable at a time in its tone's colour, as the terminal
/// game does (internal/ui/tones.go), so tones are seen every time a word is.
enum Pinyin {
    /// Neutral, then tones 1 to 4.
    static let colours: [Color] = [
        Color(red: 0.60, green: 0.63, blue: 0.65),
        Color(red: 0.88, green: 0.42, blue: 0.35),
        Color(red: 0.90, green: 0.66, blue: 0.29),
        Color(red: 0.44, green: 0.75, blue: 0.45),
        Color(red: 0.42, green: 0.63, blue: 0.88),
    ]

    private static let marks: [Character: Int] = {
        var out: [Character: Int] = [:]
        for (i, vowels) in ["āēīōūǖĀĒĪŌŪǕ", "áéíóúǘÁÉÍÓÚǗ", "ǎěǐǒǔǚǍĚǏǑǓǙ", "àèìòùǜÀÈÌÒÙǛ"].enumerated() {
            for v in vowels { out[v] = i + 1 }
        }
        return out
    }()

    /// The tone of a syllable from its mark; 0 when it has none.
    static func tone(of syllable: Substring) -> Int {
        syllable.lazy.compactMap { marks[$0] }.first ?? 0
    }

    /// The pinyin as text, each syllable in its tone's colour (or plain,
    /// when the learner turns tone colours off).
    @MainActor static func coloured(_ pinyin: String) -> Text {
        var text = Text("")
        for (i, word) in pinyin.split(separator: " ", omittingEmptySubsequences: false).enumerated() {
            if i > 0 { text = text + Text(" ") }
            text = text + Text(String(word)).foregroundStyle(SoundSettings.toneColours ? colours[tone(of: word)] : Theme.slate)
        }
        return text
    }
}
