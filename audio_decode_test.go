package radix

import (
	"bytes"
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
