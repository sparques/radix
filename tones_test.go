package radix

import (
	"math/cmplx"
	"testing"
)

func TestBuildToneFramesShapeAndSeeds(t *testing.T) {
	mode, err := NewMode(QAM16, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Setup(mode)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("ANONYMOUS")
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
	frames, err := BuildToneFrames(cfg, meta, payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != cfg.SymbolCount+1 {
		t.Fatalf("len(frames) = %d, want %d", len(frames), cfg.SymbolCount+1)
	}
	for j, frame := range frames {
		if len(frame) != ToneCount {
			t.Fatalf("len(frames[%d]) = %d, want %d", j, len(frame), ToneCount)
		}
	}

	seed := NewMLS(MLS1Poly, 0)
	plan, err := BuildTonePlan(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, symbolPlan := range plan {
		for _, tone := range symbolPlan.Tones {
			if tone.Kind != SeedTone {
				continue
			}
			want := complex(float64(NRZ(seed.NextBit())), 0)
			if frames[symbolPlan.Index][tone.Index] != want {
				t.Fatalf("seed frame %d tone %d = %v, want %v", symbolPlan.Index, tone.Index, frames[symbolPlan.Index][tone.Index], want)
			}
		}
	}
}

func TestBuildToneFramesMapsMetadataAndPayload(t *testing.T) {
	mode, err := NewMode(QPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Setup(mode)
	if err != nil {
		t.Fatal(err)
	}
	meta := make([]int8, MetadataCodeBits)
	for i := range meta {
		meta[i] = 1
	}
	payload := make([]int8, 1<<cfg.CodeOrder)
	for i := range payload {
		payload[i] = 1
	}
	frames, err := BuildToneFrames(cfg, meta, payload)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := BuildTonePlan(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, tone := range plan[0].Tones {
		if tone.Kind == MetaTone && frames[0][tone.Index] != complex(1, 0) {
			t.Fatalf("metadata tone %d = %v, want 1+0i", tone.Index, frames[0][tone.Index])
		}
	}
	for _, tone := range plan[1].Tones {
		if tone.Kind == DataTone {
			want := complex(rcpSqrt2, rcpSqrt2)
			if cmplx.Abs(frames[1][tone.Index]-want) > 1e-12 {
				t.Fatalf("payload tone %d = %v, want %v", tone.Index, frames[1][tone.Index], want)
			}
			return
		}
	}
	t.Fatal("no data tone found")
}

func TestDecodeToneFramesRoundTrip(t *testing.T) {
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
	frames, err := BuildToneFrames(cfg, meta, payload)
	if err != nil {
		t.Fatal(err)
	}
	gotMeta, gotPayload, err := DecodeToneFrames(cfg, frames)
	if err != nil {
		t.Fatal(err)
	}
	for i := range meta {
		if gotMeta[i] != meta[i] {
			t.Fatalf("meta[%d] = %d, want %d", i, gotMeta[i], meta[i])
		}
	}
	for i := range payload {
		if gotPayload[i] != payload[i] {
			t.Fatalf("payload[%d] = %d, want %d", i, gotPayload[i], payload[i])
		}
	}
}

func TestDecodeFramesForModeEqualizesSeedToneGain(t *testing.T) {
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
	wantPayload := []byte("equalize")
	payloadCode, err := EncodePayload(cfg, wantPayload)
	if err != nil {
		t.Fatal(err)
	}
	frames, err := BuildToneFrames(cfg, meta, payloadCode)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := BuildTonePlan(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for frameIdx, symbolPlan := range plan {
		estimates := make([]seedGainEstimate, 0, SeedTones)
		for _, tone := range symbolPlan.Tones {
			if tone.Kind != SeedTone {
				continue
			}
			amplitude := 0.35 + 0.001*float64((tone.Index+frameIdx)%50)
			phase := 0.6 + 0.003*float64(tone.Index) + 0.01*float64(frameIdx%3)
			estimates = append(estimates, seedGainEstimate{
				tone: tone.Index,
				gain: cmplx.Rect(amplitude, phase),
			})
		}
		for toneIdx := range frames[frameIdx] {
			gain, ok := nearestSeedGain(toneIdx, estimates)
			if !ok {
				t.Fatal("no seed gain estimate")
			}
			frames[frameIdx][toneIdx] *= gain
		}
	}

	metadata, payload, err := decodeFramesForMode(mode, cfg, frames)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.CallSignValue != call {
		t.Fatalf("call = %d, want %d", metadata.CallSignValue, call)
	}
	if string(payload[:len(wantPayload)]) != string(wantPayload) {
		t.Fatalf("payload prefix = %q, want %q", payload[:len(wantPayload)], wantPayload)
	}
}
