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

func TestReadComplex64LE(t *testing.T) {
	var buf bytes.Buffer
	want := []complex64{complex(1, -1), complex(0.5, 0.25)}
	if err := WriteComplex64LE(&buf, want); err != nil {
		t.Fatal(err)
	}
	got, err := ReadComplex64LE(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sample %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestReadFloat32LEAndConverters(t *testing.T) {
	var buf bytes.Buffer
	values := []float32{1, -1, 0.5, 0.25}
	if err := WriteFloat32LE(&buf, values); err != nil {
		t.Fatal(err)
	}
	gotValues, err := ReadFloat32LE(&buf)
	if err != nil {
		t.Fatal(err)
	}
	for i := range values {
		if gotValues[i] != values[i] {
			t.Fatalf("float %d = %g, want %g", i, gotValues[i], values[i])
		}
	}
	complexSamples, err := InterleavedFloat32ToComplex(gotValues)
	if err != nil {
		t.Fatal(err)
	}
	if len(complexSamples) != 2 || complexSamples[0] != complex(1, -1) || complexSamples[1] != complex(0.5, 0.25) {
		t.Fatalf("InterleavedFloat32ToComplex = %v", complexSamples)
	}
	mono := MonoFloat32ToComplex([]float32{1, -1})
	if len(mono) != 2 || mono[0] != complex(1, 0) || mono[1] != complex(-1, 0) {
		t.Fatalf("MonoFloat32ToComplex = %v", mono)
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
