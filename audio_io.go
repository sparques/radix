package radix

import (
	"encoding/binary"
	"io"
	"math"
)

func EncodeComplexTo(w io.Writer, cfg EncoderConfig, payload []byte) error {
	samples, err := EncodeComplex(cfg, payload)
	if err != nil {
		return err
	}
	return WriteComplex64LE(w, samples)
}

func WriteComplex64LE(w io.Writer, samples []complex64) error {
	var buf [8]byte
	for _, sample := range samples {
		binary.LittleEndian.PutUint32(buf[0:4], math.Float32bits(real(sample)))
		binary.LittleEndian.PutUint32(buf[4:8], math.Float32bits(imag(sample)))
		if _, err := w.Write(buf[:]); err != nil {
			return err
		}
	}
	return nil
}

func WriteInterleavedFloat32LE(w io.Writer, samples []complex64) error {
	return WriteFloat32LE(w, ComplexToInterleavedFloat32(samples))
}

func WriteMonoFloat32LE(w io.Writer, samples []complex64) error {
	return WriteFloat32LE(w, ComplexToMonoFloat32(samples))
}

func WriteFloat32LE(w io.Writer, samples []float32) error {
	var buf [4]byte
	for _, sample := range samples {
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(sample))
		if _, err := w.Write(buf[:]); err != nil {
			return err
		}
	}
	return nil
}

func ReadComplex64LE(r io.Reader) ([]complex64, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(raw)%8 != 0 {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]complex64, len(raw)/8)
	for i := range out {
		re := math.Float32frombits(binary.LittleEndian.Uint32(raw[8*i : 8*i+4]))
		im := math.Float32frombits(binary.LittleEndian.Uint32(raw[8*i+4 : 8*i+8]))
		out[i] = complex(re, im)
	}
	return out, nil
}

func ReadFloat32LE(r io.Reader) ([]float32, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(raw)%4 != 0 {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]float32, len(raw)/4)
	for i := range out {
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(raw[4*i : 4*i+4]))
	}
	return out, nil
}

func InterleavedFloat32ToComplex(samples []float32) ([]complex64, error) {
	if len(samples)%2 != 0 {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]complex64, len(samples)/2)
	for i := range out {
		out[i] = complex(samples[2*i], samples[2*i+1])
	}
	return out, nil
}

func MonoFloat32ToComplex(samples []float32) []complex64 {
	out := make([]complex64, len(samples))
	for i, sample := range samples {
		out[i] = complex(sample, 0)
	}
	return out
}

func AnalyzeComplexAlignedFrom(r io.Reader, cfg AlignedDecoderConfig) (ToneFrames, error) {
	samples, err := ReadComplex64LE(r)
	if err != nil {
		return nil, err
	}
	return AnalyzeComplexAligned(cfg, samples)
}
