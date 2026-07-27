package tts

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
)

// QwenProvider drives the qwen_tts CLI, reading raw PCM from its --stdout.
// Text is passed as a discrete argv entry (never through a shell), so card
// content can't inject commands.
type QwenProvider struct {
	Bin      string  // path to qwen_tts; defaults to "qwen_tts"
	Model    string  // -d, e.g. "qwen3-tts-1.7b"
	Language string  // -l, e.g. "Chinese"
	Speaker  string  // -s, e.g. "serena"
	Temp     float64 // -T, e.g. 0.3
	Seed     int     // --seed; <0 to omit

	// The raw stream qwen_tts emits. Kept explicit so it flows into ffmpeg's
	// input flags rather than being hard-coded in two places.
	SampleRate int // e.g. 24000
	Channels   int // e.g. 1
}

func (q QwenProvider) Speak(ctx context.Context, text string) (RawAudio, error) {
	bin := q.Bin
	if bin == "" {
		bin = "qwen_tts"
	}

	args := []string{
		"-d", q.Model,
		"-l", q.Language,
		"-s", q.Speaker,
		"-T", strconv.FormatFloat(q.Temp, 'g', -1, 64),
	}
	if q.Seed >= 0 {
		args = append(args, "--seed", strconv.Itoa(q.Seed))
	}
	args = append(args, "--text", text, "--stdout")

	cmd := exec.CommandContext(ctx, bin, args...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return RawAudio{}, fmt.Errorf("tts: qwen cancelled: %w", ctxErr)
		}
		return RawAudio{}, fmt.Errorf("tts: qwen: %w: %s", err, stderr.String())
	}

	return RawAudio{
		Data:       out.Bytes(),
		SampleRate: q.SampleRate,
		Channels:   q.Channels,
	}, nil
}
