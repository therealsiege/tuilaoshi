package audio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

const fishAPIBase = "https://api.fish.audio"

// Voice represents a Fish Audio voice model.
type Voice struct {
	ID   string
	Name string
	Lang string // "zh" or "en"
}

// Chinese voices
var (
	VoiceChiangKaiShek = Voice{ID: "918a8277663d476b95e2c4867da0f6a6", Name: "蒋介石", Lang: "zh"}
	VoiceDingZhen      = Voice{ID: "54a5170264694bfc8e9ad98df7bd89c3", Name: "丁真", Lang: "zh"}
	VoiceFemalStudent  = Voice{ID: "5c353fdb312f4888836a9a5680099ef0", Name: "女大学生", Lang: "zh"}
	VoiceJiaLan        = Voice{ID: "fbe02f8306fc4d3d915e9871722a39d5", Name: "嘉岚3.0", Lang: "zh"}
)

// English voices
var (
	VoiceEthan = Voice{ID: "536d3a5e000945adb7038665781a4aca", Name: "Ethan", Lang: "en"}
	VoiceSarah = Voice{ID: "933563129e564b19a115bedd57b7406a", Name: "Sarah", Lang: "en"}
)

// ChineseVoices returns all available Chinese voices.
func ChineseVoices() []Voice {
	return []Voice{VoiceJiaLan, VoiceFemalStudent, VoiceDingZhen, VoiceChiangKaiShek}
}

// EnglishVoices returns all available English voices.
func EnglishVoices() []Voice {
	return []Voice{VoiceSarah, VoiceEthan}
}

// ttsRequest is the Fish Audio /v1/tts JSON payload.
type ttsRequest struct {
	Text        string     `json:"text"`
	ReferenceID string     `json:"reference_id"`
	Format      string     `json:"format"`
	MP3Bitrate  int        `json:"mp3_bitrate"`
	Normalize   bool       `json:"normalize"`
	Latency     string     `json:"latency"`
	Prosody     ttsProsody `json:"prosody"`
}

type ttsProsody struct {
	Speed  float64 `json:"speed"`
	Volume int     `json:"volume"`
}

// Client is a Fish Audio TTS client.
type Client struct {
	apiKey     string
	httpClient *http.Client
	zhVoice    Voice
	enVoice    Voice
}

// New creates a TTS client. Reads FISH_AUDIO_API_KEY from env.
func New() *Client {
	return &Client{
		apiKey:     os.Getenv("FISH_AUDIO_API_KEY"),
		httpClient: &http.Client{},
		zhVoice:    VoiceJiaLan,
		enVoice:    VoiceSarah,
	}
}

// Available reports whether the TTS client has an API key configured.
func (c *Client) Available() bool {
	return c.apiKey != ""
}

// SetChineseVoice selects the Chinese voice to use.
func (c *Client) SetChineseVoice(v Voice) { c.zhVoice = v }

// SetEnglishVoice selects the English voice to use.
func (c *Client) SetEnglishVoice(v Voice) { c.enVoice = v }

// ChineseVoice returns the currently selected Chinese voice.
func (c *Client) ChineseVoice() Voice { return c.zhVoice }

// EnglishVoice returns the currently selected English voice.
func (c *Client) EnglishVoice() Voice { return c.enVoice }

// Speak synthesizes text and plays it through the system audio.
// lang should be "zh" or "en" to select the appropriate voice.
func (c *Client) Speak(text, lang string) error {
	if !c.Available() {
		return fmt.Errorf("FISH_AUDIO_API_KEY not set")
	}
	if text == "" {
		return nil
	}

	voice := c.zhVoice
	if lang == "en" {
		voice = c.enVoice
	}

	audioData, err := c.synthesize(text, voice)
	if err != nil {
		return err
	}

	return playAudio(audioData)
}

// synthesize calls the Fish Audio TTS API and returns raw MP3 bytes.
func (c *Client) synthesize(text string, voice Voice) ([]byte, error) {
	payload := ttsRequest{
		Text:        text,
		ReferenceID: voice.ID,
		Format:      "mp3",
		MP3Bitrate:  192,
		Normalize:   true,
		Latency:     "normal",
		Prosody:     ttsProsody{Speed: 1.0, Volume: 0},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal tts request: %w", err)
	}

	req, err := http.NewRequest("POST", fishAPIBase+"/v1/tts", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create tts request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tts api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tts api error %d: %s", resp.StatusCode, string(errBody))
	}

	return io.ReadAll(resp.Body)
}

// playAudio plays MP3 data through the system's default audio output.
func playAudio(data []byte) error {
	// Write to a temp file — afplay (macOS) requires a file path, not stdin.
	tmpFile, err := os.CreateTemp("", "mandarin-tts-*.mp3")
	if err != nil {
		return fmt.Errorf("create temp audio file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp audio file: %w", err)
	}
	tmpFile.Close()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("afplay", tmpFile.Name())
	case "linux":
		if _, err := exec.LookPath("mpv"); err == nil {
			cmd = exec.Command("mpv", "--no-video", "--really-quiet", tmpFile.Name())
		} else {
			cmd = exec.Command("ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", tmpFile.Name())
		}
	default:
		return fmt.Errorf("audio playback not supported on %s", runtime.GOOS)
	}

	return cmd.Run()
}
