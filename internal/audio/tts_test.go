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
