package input

// Mode determines how keystrokes are handled in the input.
type Mode int

const (
	ModeNormal  Mode = iota // EN — passthrough
	ModePinyin              // 拼 — numbered pinyin auto-converts on space
	ModeChinese             // 中 — pinyin composition with candidate selection
)

// Label returns a short display label for the mode.
func (m Mode) Label() string {
	switch m {
	case ModePinyin:
		return "拼"
	case ModeChinese:
		return "中"
	default:
		return "EN"
	}
}

// Next cycles to the next mode.
func (m Mode) Next() Mode {
	switch m {
	case ModeNormal:
		return ModePinyin
	case ModePinyin:
		return ModeChinese
	default:
		return ModeNormal
	}
}
