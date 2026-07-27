package tts

import (
	"fmt"
	"strconv"
)

type AudioFormat string

const (
	FormatMP3  AudioFormat = "mp3"
	FormatWAV  AudioFormat = "wav"
	FormatOGG  AudioFormat = "ogg"
	FormatOpus AudioFormat = "opus"
)

// codec returns the ffmpeg audio encoder for the format.
func (f AudioFormat) codec() string {
	switch f {
	case FormatMP3:
		return "libmp3lame"
	case FormatWAV:
		return "pcm_s16le"
	case FormatOGG:
		return "libvorbis"
	case FormatOpus:
		return "libopus"
	default:
		return ""
	}
}

// muxer returns the ffmpeg container/format name (-f). Required when
// writing to a pipe, since ffmpeg can't infer it from a filename.
func (f AudioFormat) muxer() string {
	if f == FormatOpus {
		return "opus"
	}
	return string(f)
}

// RawAudio is headerless PCM produced by a Provider (e.g. qwen_tts --stdout).
// ffmpeg needs the sample rate and channel count up front to read it, since
// there's no container to describe the stream. Sample format is fixed s16le.
type RawAudio struct {
	Data       []byte
	SampleRate int // Hz, e.g. 24000
	Channels   int // 1 = mono, 2 = stereo
}

// inputArgs are the ffmpeg flags that describe the raw stream on stdin.
// They must precede -i.
func (r RawAudio) inputArgs() []string {
	return []string{
		"-f", "s16le",
		"-ar", strconv.Itoa(r.SampleRate),
		"-ac", strconv.Itoa(r.Channels),
	}
}

type EncodeSettings struct {
	Format     AudioFormat
	SampleRate int // Hz, e.g. 24000; 0 = keep source rate
	Channels   int // 1 = mono, 2 = stereo; 0 = keep source channels
	BitrateK   int // kbps, e.g. 32; 0 = codec default
}

// Validate rejects configurations ffmpeg can't satisfy before we spawn it.
func (s EncodeSettings) Validate() error {
	if s.Format.codec() == "" {
		return fmt.Errorf("tts: unsupported format %q", s.Format)
	}
	if s.SampleRate < 0 || s.Channels < 0 || s.BitrateK < 0 {
		return fmt.Errorf("tts: negative encode setting")
	}
	return nil
}

// args builds the ffmpeg argument list to encode raw PCM read on stdin
// into `out`. Pass "-" for stdout. This is the ONLY place that knows
// ffmpeg's flag syntax.
func (s EncodeSettings) args(in RawAudio, out string) []string {
	a := []string{"-hide_banner", "-loglevel", "error"}
	a = append(a, in.inputArgs()...) // -f s16le -ar -ac, must precede -i
	a = append(a, "-i", "-", "-c:a", s.Format.codec())
	if s.SampleRate > 0 {
		a = append(a, "-ar", strconv.Itoa(s.SampleRate))
	}
	if s.Channels > 0 {
		a = append(a, "-ac", strconv.Itoa(s.Channels))
	}
	if s.BitrateK > 0 {
		a = append(a, "-b:a", fmt.Sprintf("%dk", s.BitrateK))
	}
	// -f is mandatory when out is a pipe; harmless for real files.
	a = append(a, "-f", s.Format.muxer(), out)
	return a
}
