package audio

import "testing"

func TestAvailableWithSayFallback(t *testing.T) {
	if (&Client{}).Available() {
		t.Error("client with no key and no say reports available")
	}
	if !(&Client{sayPath: "/usr/bin/say"}).Available() {
		t.Error("client with say reports unavailable")
	}
	if err := (&Client{}).Speak("你好", "zh"); err == nil {
		t.Error("Speak with no key and no say returned no error")
	}
}

func TestASecondSpeakerHasAnotherVoice(t *testing.T) {
	c := New()
	for _, v := range ChineseVoices() {
		c.SetChineseVoice(v)
		if o := c.OtherChineseVoice(); o.ID == v.ID || o.Lang != "zh" {
			t.Errorf("with %s chosen, the other voice is %s", v.Name, o.Name)
		}
	}
	if err := (&Client{}).SpeakLines([]Line{{Text: "你好"}}); err == nil {
		t.Error("SpeakLines with no key and no say returned no error")
	}
}
