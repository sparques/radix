package radix

import "testing"

func TestAudioConfigLengths(t *testing.T) {
	cfg := AudioConfig{SampleRate: 48000, FrequencyOffset: 1500}
	guardLen, err := cfg.GuardLen()
	if err != nil {
		t.Fatal(err)
	}
	if guardLen != 160 {
		t.Fatalf("GuardLen = %d, want 160", guardLen)
	}
	symbolLen, err := cfg.SymbolLen()
	if err != nil {
		t.Fatal(err)
	}
	if symbolLen != 6400 {
		t.Fatalf("SymbolLen = %d, want 6400", symbolLen)
	}
	toneOffset, err := cfg.ToneOffset()
	if err != nil {
		t.Fatal(err)
	}
	if toneOffset != 40 {
		t.Fatalf("ToneOffset = %d, want 40", toneOffset)
	}
}

func TestEncodeComplexProducesBoundedSamples(t *testing.T) {
	mode, err := NewMode(BPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("N0CALL")
	if err != nil {
		t.Fatal(err)
	}
	audio := AudioConfig{SampleRate: 44100, FrequencyOffset: 1500}
	samples, err := EncodeComplex(EncoderConfig{
		Audio:    audio,
		Mode:     mode,
		CallSign: call,
	}, []byte("hi"))
	if err != nil {
		t.Fatal(err)
	}
	guardLen, _ := audio.GuardLen()
	symbolLen, _ := audio.SymbolLen()
	cfg, _ := Setup(mode)
	wantLen := (cfg.SymbolCount + 4) * (guardLen + symbolLen)
	if len(samples) != wantLen {
		t.Fatalf("len(samples) = %d, want %d", len(samples), wantLen)
	}
	var nonzero bool
	for i, sample := range samples {
		if real(sample) < -1 || real(sample) > 1 || imag(sample) < -1 || imag(sample) > 1 {
			t.Fatalf("sample %d = %v outside clipped range", i, sample)
		}
		if sample != 0 {
			nonzero = true
		}
	}
	if !nonzero {
		t.Fatal("EncodeComplex returned all zero samples")
	}
}

func TestFloat32Adapters(t *testing.T) {
	samples := []complex64{complex(1, -1), complex(0.5, 0.25)}
	stereo := ComplexToInterleavedFloat32(samples)
	if len(stereo) != 4 || stereo[0] != 1 || stereo[1] != -1 || stereo[2] != 0.5 || stereo[3] != 0.25 {
		t.Fatalf("ComplexToInterleavedFloat32 = %v", stereo)
	}
	mono := ComplexToMonoFloat32(samples)
	if len(mono) != 2 || mono[0] != 1 || mono[1] != 0.5 {
		t.Fatalf("ComplexToMonoFloat32 = %v", mono)
	}
}
