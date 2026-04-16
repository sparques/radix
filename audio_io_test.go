package radix

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestWriteComplex64LE(t *testing.T) {
	var buf bytes.Buffer
	err := WriteComplex64LE(&buf, []complex64{complex(1, -1), complex(0.5, 0.25)})
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 16 {
		t.Fatalf("buffer length = %d, want 16", buf.Len())
	}
	got := buf.Bytes()
	values := []float32{
		math.Float32frombits(binary.LittleEndian.Uint32(got[0:4])),
		math.Float32frombits(binary.LittleEndian.Uint32(got[4:8])),
		math.Float32frombits(binary.LittleEndian.Uint32(got[8:12])),
		math.Float32frombits(binary.LittleEndian.Uint32(got[12:16])),
	}
	want := []float32{1, -1, 0.5, 0.25}
	for i := range want {
		if values[i] != want[i] {
			t.Fatalf("value %d = %g, want %g", i, values[i], want[i])
		}
	}
}

func TestEncodeComplexTo(t *testing.T) {
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
	err = EncodeComplexTo(&buf, EncoderConfig{
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
	wantSamples := (cfg.SymbolCount + 4) * (guardLen + symbolLen)
	if buf.Len() != wantSamples*8 {
		t.Fatalf("buffer length = %d, want %d", buf.Len(), wantSamples*8)
	}
}
