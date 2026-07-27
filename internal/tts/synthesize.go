package tts

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
)

// Provider turns text into raw PCM audio. Implementations wrap whatever
// TTS engine you use (qwen_tts, piper, a cloud API, …).
type Provider interface {
	Speak(ctx context.Context, text string) (RawAudio, error)
}

// Synthesizer produces encoded audio for a piece of text: it asks the
// Provider for raw PCM, then pipes it through ffmpeg to reach the
// requested EncodeSettings.
type Synthesizer struct {
	provider   Provider
	logger     *slog.Logger
	ffmpegPath string // defaults to "ffmpeg" on PATH
}

func NewSynthesizer(p Provider, logger *slog.Logger) *Synthesizer {
	return &Synthesizer{provider: p, logger: logger, ffmpegPath: "ffmpeg"}
}

// Synthesize renders text to audio encoded per settings and returns the
// bytes. ctx bounds the whole operation, including the ffmpeg subprocess —
// cancelling ctx kills ffmpeg (CommandContext).
func (s *Synthesizer) Synthesize(ctx context.Context, text string, settings EncodeSettings) ([]byte, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	raw, err := s.provider.Speak(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("tts: provider speak: %w", err)
	}

	// Encode by piping raw PCM in on stdin and reading the result on stdout.
	cmd := exec.CommandContext(ctx, s.ffmpegPath, settings.args(raw, "-")...)
	cmd.Stdin = bytes.NewReader(raw.Data)

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	s.logger.DebugContext(ctx, "encoding audio",
		"format", settings.Format, "pcm_bytes", len(raw.Data))

	if err := cmd.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("tts: encode cancelled: %w", ctxErr)
		}
		return nil, fmt.Errorf("tts: ffmpeg: %w: %s", err, stderr.String())
	}

	return out.Bytes(), nil
}
