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

func WriteInterleavedInt16LE(w io.Writer, samples []complex64) error {
	return WriteInt16LE(w, ComplexToInterleavedInt16(samples))
}

func WriteMonoInt16LE(w io.Writer, samples []complex64) error {
	return WriteInt16LE(w, ComplexToMonoInt16(samples))
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

func WriteInt16LE(w io.Writer, samples []int16) error {
	var buf [2]byte
	for _, sample := range samples {
		binary.LittleEndian.PutUint16(buf[:], uint16(sample))
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

func ReadInt16LE(r io.Reader) ([]int16, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(raw)%2 != 0 {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]int16, len(raw)/2)
	for i := range out {
		out[i] = int16(binary.LittleEndian.Uint16(raw[2*i : 2*i+2]))
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

func ComplexToInterleavedInt16(samples []complex64) []int16 {
	out := make([]int16, 2*len(samples))
	for i, sample := range samples {
		out[2*i] = float32ToInt16(real(sample))
		out[2*i+1] = float32ToInt16(imag(sample))
	}
	return out
}

func ComplexToMonoInt16(samples []complex64) []int16 {
	out := make([]int16, len(samples))
	for i, sample := range samples {
		out[i] = float32ToInt16(real(sample))
	}
	return out
}

func InterleavedInt16ToComplex(samples []int16) ([]complex64, error) {
	if len(samples)%2 != 0 {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]complex64, len(samples)/2)
	for i := range out {
		out[i] = complex(int16ToFloat32(samples[2*i]), int16ToFloat32(samples[2*i+1]))
	}
	return out, nil
}

func MonoInt16ToComplex(samples []int16) []complex64 {
	out := make([]complex64, len(samples))
	for i, sample := range samples {
		out[i] = complex(int16ToFloat32(sample), 0)
	}
	return out
}

func float32ToInt16(sample float32) int16 {
	sample = float32(clamp(float64(sample), -1, 1))
	if sample < 0 {
		return int16(math.Round(float64(sample) * 32768))
	}
	return int16(math.Round(float64(sample) * 32767))
}

func int16ToFloat32(sample int16) float32 {
	return float32(sample) / 32768
}

func AnalyzeComplexAlignedFrom(r io.Reader, cfg AlignedDecoderConfig) (ToneFrames, error) {
	samples, err := ReadComplex64LE(r)
	if err != nil {
		return nil, err
	}
	return AnalyzeComplexAligned(cfg, samples)
}

func DecodeAlignedFrom(r io.Reader, cfg AlignedDecoderConfig) (Metadata, []byte, error) {
	samples, err := ReadComplex64LE(r)
	if err != nil {
		return Metadata{}, nil, err
	}
	return DecodeAligned(cfg, samples)
}

func DecodeCapturedFrom(r io.Reader, cfg AlignedDecoderConfig) (Metadata, []byte, Acquisition, error) {
	samples, err := ReadComplex64LE(r)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	return DecodeCaptured(cfg, samples)
}

func DecodeInterleavedFloat32CapturedFrom(r io.Reader, cfg AlignedDecoderConfig) (Metadata, []byte, Acquisition, error) {
	raw, err := ReadFloat32LE(r)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	samples, err := InterleavedFloat32ToComplex(raw)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	return DecodeCaptured(cfg, samples)
}

func DecodeMonoFloat32CapturedFrom(r io.Reader, cfg AlignedDecoderConfig) (Metadata, []byte, Acquisition, error) {
	raw, err := ReadFloat32LE(r)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	return DecodeCaptured(cfg, MonoFloat32ToComplex(raw))
}

func DecodeInterleavedInt16CapturedFrom(r io.Reader, cfg AlignedDecoderConfig) (Metadata, []byte, Acquisition, error) {
	raw, err := ReadInt16LE(r)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	samples, err := InterleavedInt16ToComplex(raw)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	return DecodeCaptured(cfg, samples)
}

func DecodeMonoInt16CapturedFrom(r io.Reader, cfg AlignedDecoderConfig) (Metadata, []byte, Acquisition, error) {
	raw, err := ReadInt16LE(r)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	return DecodeCaptured(cfg, MonoInt16ToComplex(raw))
}
