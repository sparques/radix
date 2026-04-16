package radix

import (
	"bytes"
	"math"
	"testing"
)

func TestAnalyzeComplexAlignedRoundTripsToneFrames(t *testing.T) {
	mode, err := NewMode(QPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Setup(mode)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("N0CALL")
	if err != nil {
		t.Fatal(err)
	}
	meta, err := EncodeMetadata(call, mode)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := EncodePayload(cfg, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	wantFrames, err := BuildToneFrames(cfg, meta, payload)
	if err != nil {
		t.Fatal(err)
	}

	audio := AudioConfig{SampleRate: 44100, FrequencyOffset: 1500}
	samples, err := EncodeComplex(EncoderConfig{Audio: audio, Mode: mode, CallSign: call}, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	gotFrames, err := AnalyzeComplexAligned(AlignedDecoderConfig{Audio: audio, Mode: mode}, samples)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotFrames) != len(wantFrames) {
		t.Fatalf("frames = %d, want %d", len(gotFrames), len(wantFrames))
	}

	plan, err := BuildTonePlan(cfg)
	if err != nil {
		t.Fatal(err)
	}
	constellation, err := NewConstellation(QPSK)
	if err != nil {
		t.Fatal(err)
	}
	for frameIdx := range wantFrames {
		for _, tone := range plan[frameIdx].Tones {
			got := gotFrames[frameIdx][tone.Index]
			want := wantFrames[frameIdx][tone.Index]
			switch tone.Kind {
			case SeedTone, MetaTone:
				if NearestSignedTone(got) != NearestSignedTone(want) {
					t.Fatalf("frame %d tone %d hard = %d, want %d from %v", frameIdx, tone.Index, NearestSignedTone(got), NearestSignedTone(want), got)
				}
			case DataTone:
				if tone.Bits != 2 {
					t.Fatalf("test expected QPSK data tones, got %d bits", tone.Bits)
				}
				gotBits := constellation.Hard(got)
				wantBits := constellation.Hard(want)
				for i := range wantBits {
					if gotBits[i] != wantBits[i] {
						t.Fatalf("frame %d tone %d hard bits = %v, want %v from %v", frameIdx, tone.Index, gotBits, wantBits, got)
					}
				}
			}
		}
	}
}

func TestAnalyzeComplexAlignedRejectsShortInput(t *testing.T) {
	mode, err := NewMode(BPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	_, err = AnalyzeComplexAligned(AlignedDecoderConfig{
		Audio: AudioConfig{SampleRate: 44100, FrequencyOffset: 1500},
		Mode:  mode,
	}, nil)
	if err == nil {
		t.Fatal("AnalyzeComplexAligned accepted empty input")
	}
}

func TestAnalyzeComplexAlignedFrom(t *testing.T) {
	mode, err := NewMode(BPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("N0CALL")
	if err != nil {
		t.Fatal(err)
	}
	audio := AudioConfig{SampleRate: 44100, FrequencyOffset: 1500}
	var buf bytes.Buffer
	if err := EncodeComplexTo(&buf, EncoderConfig{Audio: audio, Mode: mode, CallSign: call}, []byte("hi")); err != nil {
		t.Fatal(err)
	}
	frames, err := AnalyzeComplexAlignedFrom(&buf, AlignedDecoderConfig{Audio: audio, Mode: mode})
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Setup(mode)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != cfg.SymbolCount+1 {
		t.Fatalf("frames = %d, want %d", len(frames), cfg.SymbolCount+1)
	}
}

func TestDecodeAlignedEndToEnd(t *testing.T) {
	mode, err := NewMode(QPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("N0CALL")
	if err != nil {
		t.Fatal(err)
	}
	audio := AudioConfig{SampleRate: 44100, FrequencyOffset: 1500}
	wantPayload := []byte("hello")
	samples, err := EncodeComplex(EncoderConfig{Audio: audio, Mode: mode, CallSign: call}, wantPayload)
	if err != nil {
		t.Fatal(err)
	}
	metadata, payload, err := DecodeAligned(AlignedDecoderConfig{Audio: audio, Mode: mode}, samples)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.CallSignValue != call {
		t.Fatalf("call = %d, want %d", metadata.CallSignValue, call)
	}
	if metadata.Mode != mode {
		t.Fatalf("mode = %d, want %d", metadata.Mode, mode)
	}
	if string(payload[:len(wantPayload)]) != string(wantPayload) {
		t.Fatalf("payload prefix = %q, want %q", payload[:len(wantPayload)], wantPayload)
	}
}

func TestDecodeAlignedFrom(t *testing.T) {
	mode, err := NewMode(BPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("N0CALL")
	if err != nil {
		t.Fatal(err)
	}
	audio := AudioConfig{SampleRate: 44100, FrequencyOffset: 1500}
	var buf bytes.Buffer
	if err := EncodeComplexTo(&buf, EncoderConfig{Audio: audio, Mode: mode, CallSign: call}, []byte("hi")); err != nil {
		t.Fatal(err)
	}
	metadata, payload, err := DecodeAlignedFrom(&buf, AlignedDecoderConfig{Audio: audio, Mode: mode})
	if err != nil {
		t.Fatal(err)
	}
	if metadata.CallSignValue != call {
		t.Fatalf("call = %d, want %d", metadata.CallSignValue, call)
	}
	if string(payload[:2]) != "hi" {
		t.Fatalf("payload prefix = %q, want hi", payload[:2])
	}
}

func TestDecodeCapturedEndToEndWithLeadingSamples(t *testing.T) {
	mode, err := NewMode(QPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("N0CALL")
	if err != nil {
		t.Fatal(err)
	}
	audio := AudioConfig{SampleRate: 44100, FrequencyOffset: 1500}
	wantPayload := []byte("captured")
	samples, err := EncodeComplex(EncoderConfig{Audio: audio, Mode: mode, CallSign: call}, wantPayload)
	if err != nil {
		t.Fatal(err)
	}

	const leading = 321
	captured := make([]complex64, leading+len(samples)+123)
	for i := 0; i < leading; i++ {
		captured[i] = complex(float32(i%7)*0.0001, float32(i%5)*-0.0001)
	}
	copy(captured[leading:], samples)

	decoder := AlignedDecoderConfig{Audio: audio, Mode: mode}
	metadata, payload, acquisition, err := DecodeCaptured(decoder, captured)
	if err != nil {
		t.Fatal(err)
	}
	if acquisition.PreambleStart != leading {
		t.Fatalf("preamble start = %d, want %d", acquisition.PreambleStart, leading)
	}
	if metadata.CallSignValue != call {
		t.Fatalf("call = %d, want %d", metadata.CallSignValue, call)
	}
	if string(payload[:len(wantPayload)]) != string(wantPayload) {
		t.Fatalf("payload prefix = %q, want %q", payload[:len(wantPayload)], wantPayload)
	}
}

func TestDecodeInterleavedFloat32CapturedFrom(t *testing.T) {
	mode, err := NewMode(BPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("N0CALL")
	if err != nil {
		t.Fatal(err)
	}
	audio := AudioConfig{SampleRate: 44100, FrequencyOffset: 1500}
	samples, err := EncodeComplex(EncoderConfig{Audio: audio, Mode: mode, CallSign: call}, []byte("iq"))
	if err != nil {
		t.Fatal(err)
	}
	captured := make([]complex64, 77+len(samples))
	copy(captured[77:], samples)

	var buf bytes.Buffer
	if err := WriteInterleavedFloat32LE(&buf, captured); err != nil {
		t.Fatal(err)
	}
	metadata, payload, acquisition, err := DecodeInterleavedFloat32CapturedFrom(&buf, AlignedDecoderConfig{Audio: audio, Mode: mode})
	if err != nil {
		t.Fatal(err)
	}
	if acquisition.PreambleStart != 77 {
		t.Fatalf("preamble start = %d, want 77", acquisition.PreambleStart)
	}
	if metadata.CallSignValue != call {
		t.Fatalf("call = %d, want %d", metadata.CallSignValue, call)
	}
	if string(payload[:2]) != "iq" {
		t.Fatalf("payload prefix = %q, want iq", payload[:2])
	}
}

func TestDecodeInterleavedInt16CapturedFrom(t *testing.T) {
	mode, err := NewMode(BPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("N0CALL")
	if err != nil {
		t.Fatal(err)
	}
	audio := AudioConfig{SampleRate: 44100, FrequencyOffset: 1500}
	samples, err := EncodeComplex(EncoderConfig{Audio: audio, Mode: mode, CallSign: call}, []byte("pcm"))
	if err != nil {
		t.Fatal(err)
	}
	captured := make([]complex64, 105+len(samples))
	copy(captured[105:], samples)

	var buf bytes.Buffer
	if err := WriteInterleavedInt16LE(&buf, captured); err != nil {
		t.Fatal(err)
	}
	metadata, payload, acquisition, err := DecodeInterleavedInt16CapturedFrom(&buf, AlignedDecoderConfig{Audio: audio, Mode: mode})
	if err != nil {
		t.Fatal(err)
	}
	if acquisition.PreambleStart != 105 {
		t.Fatalf("preamble start = %d, want 105", acquisition.PreambleStart)
	}
	if metadata.CallSignValue != call {
		t.Fatalf("call = %d, want %d", metadata.CallSignValue, call)
	}
	if string(payload[:3]) != "pcm" {
		t.Fatalf("payload prefix = %q, want pcm", payload[:3])
	}
}

func TestDecodeCapturedCorrectsResidualCarrier(t *testing.T) {
	mode, err := NewMode(QPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("N0CALL")
	if err != nil {
		t.Fatal(err)
	}
	audio := AudioConfig{SampleRate: 44100, FrequencyOffset: 1500}
	wantPayload := []byte("carrier")
	samples, err := EncodeComplex(EncoderConfig{Audio: audio, Mode: mode, CallSign: call}, wantPayload)
	if err != nil {
		t.Fatal(err)
	}

	const leading = 211
	captured := make([]complex64, leading+len(samples)+95)
	copy(captured[leading:], samples)
	rotateSamples(captured, audio.SampleRate, 1.25, 1.1)

	metadata, payload, acquisition, err := DecodeCaptured(AlignedDecoderConfig{Audio: audio, Mode: mode}, captured)
	if err != nil {
		t.Fatal(err)
	}
	if acquisition.PreambleStart != leading {
		t.Fatalf("preamble start = %d, want %d", acquisition.PreambleStart, leading)
	}
	if math.Abs(acquisition.ResidualFrequencyHz-1.25) > 0.02 {
		t.Fatalf("residual frequency = %.6f, want 1.25", acquisition.ResidualFrequencyHz)
	}
	if metadata.CallSignValue != call {
		t.Fatalf("call = %d, want %d", metadata.CallSignValue, call)
	}
	if string(payload[:len(wantPayload)]) != string(wantPayload) {
		t.Fatalf("payload prefix = %q, want %q", payload[:len(wantPayload)], wantPayload)
	}
}

func rotateSamples(samples []complex64, sampleRate int, frequencyHz, phaseRadians float64) {
	omega := 2 * math.Pi * frequencyHz / float64(sampleRate)
	for i, sample := range samples {
		phase := omega*float64(i) + phaseRadians
		rot := complex64(complex(math.Cos(phase), math.Sin(phase)))
		samples[i] = sample * rot
	}
}
