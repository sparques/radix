package radix

import (
	"encoding/binary"
	"io"
	"math"
)

// EncodeComplexTo writes EncodeComplex output as little-endian complex64
// samples. Each sample is two float32 values: real then imaginary.
func EncodeComplexTo(w io.Writer, cfg EncoderConfig, payload []byte) error {
	samples, err := EncodeComplex(cfg, payload)
	if err != nil {
		return err
	}
	return WriteComplex64LE(w, samples)
}

// WriteComplex64LE writes complex samples as little-endian real/imag float32
// pairs. This is the simplest lossless stream format for Radix IQ samples.
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

// WriteInterleavedFloat32LE writes complex samples as little-endian stereo IQ
// float32 values: I, Q, I, Q.
func WriteInterleavedFloat32LE(w io.Writer, samples []complex64) error {
	return WriteFloat32LE(w, ComplexToInterleavedFloat32(samples))
}

// WriteMonoFloat32LE writes the real part of complex samples as little-endian
// mono float32 audio.
func WriteMonoFloat32LE(w io.Writer, samples []complex64) error {
	return WriteFloat32LE(w, ComplexToMonoFloat32(samples))
}

// WriteInterleavedInt16LE writes complex samples as little-endian stereo IQ
// signed 16-bit PCM.
func WriteInterleavedInt16LE(w io.Writer, samples []complex64) error {
	return WriteInt16LE(w, ComplexToInterleavedInt16(samples))
}

// WriteMonoInt16LE writes the real part of complex samples as little-endian mono
// signed 16-bit PCM.
func WriteMonoInt16LE(w io.Writer, samples []complex64) error {
	return WriteInt16LE(w, ComplexToMonoInt16(samples))
}

// WriteFloat32LE writes raw little-endian float32 samples.
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

// WriteInt16LE writes raw little-endian signed 16-bit samples.
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

// ReadComplex64LE reads little-endian real/imag float32 pairs into complex
// samples.
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

// ReadFloat32LE reads a raw little-endian float32 sample stream.
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

// ReadInt16LE reads a raw little-endian signed 16-bit sample stream.
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

// InterleavedFloat32ToComplex converts stereo-style float32 IQ values into
// complex samples.
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

// MonoFloat32ToComplex converts mono float32 samples into complex samples with a
// zero imaginary part.
func MonoFloat32ToComplex(samples []float32) []complex64 {
	out := make([]complex64, len(samples))
	for i, sample := range samples {
		out[i] = complex(sample, 0)
	}
	return out
}

// ComplexToInterleavedInt16 converts complex samples to stereo-style signed
// 16-bit IQ values.
func ComplexToInterleavedInt16(samples []complex64) []int16 {
	out := make([]int16, 2*len(samples))
	for i, sample := range samples {
		out[2*i] = float32ToInt16(real(sample))
		out[2*i+1] = float32ToInt16(imag(sample))
	}
	return out
}

// ComplexToMonoInt16 converts the real part of complex samples to signed 16-bit
// mono PCM.
func ComplexToMonoInt16(samples []complex64) []int16 {
	out := make([]int16, len(samples))
	for i, sample := range samples {
		out[i] = float32ToInt16(real(sample))
	}
	return out
}

// InterleavedInt16ToComplex converts stereo-style signed 16-bit IQ values into
// complex samples scaled roughly to [-1,+1].
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

// MonoInt16ToComplex converts mono signed 16-bit PCM into complex samples with a
// zero imaginary part.
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

// AnalyzeComplexAlignedFrom reads little-endian complex64 samples and analyzes
// them as an already-aligned frame.
func AnalyzeComplexAlignedFrom(r io.Reader, cfg AlignedDecoderConfig) (ToneFrames, error) {
	samples, err := ReadComplex64LE(r)
	if err != nil {
		return nil, err
	}
	return AnalyzeComplexAligned(cfg, samples)
}

// DecodeAlignedFrom reads little-endian complex64 samples and decodes an
// already-aligned frame.
func DecodeAlignedFrom(r io.Reader, cfg AlignedDecoderConfig) (Metadata, []byte, error) {
	samples, err := ReadComplex64LE(r)
	if err != nil {
		return Metadata{}, nil, err
	}
	return DecodeAligned(cfg, samples)
}

// DecodeCapturedFrom reads little-endian complex64 samples and runs captured
// frame acquisition and decode.
func DecodeCapturedFrom(r io.Reader, cfg AlignedDecoderConfig) (Metadata, []byte, Acquisition, error) {
	samples, err := ReadComplex64LE(r)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	return DecodeCaptured(cfg, samples)
}

// DecodeInterleavedFloat32CapturedFrom reads stereo-style float32 IQ samples and
// runs captured frame acquisition and decode.
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

// DecodeMonoFloat32CapturedFrom reads mono float32 audio and runs captured frame
// acquisition and decode.
func DecodeMonoFloat32CapturedFrom(r io.Reader, cfg AlignedDecoderConfig) (Metadata, []byte, Acquisition, error) {
	raw, err := ReadFloat32LE(r)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	return DecodeCaptured(cfg, MonoFloat32ToComplex(raw))
}

// DecodeInterleavedInt16CapturedFrom reads stereo-style signed 16-bit IQ samples
// and runs captured frame acquisition and decode.
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

// DecodeMonoInt16CapturedFrom reads mono signed 16-bit PCM audio and runs
// captured frame acquisition and decode.
func DecodeMonoInt16CapturedFrom(r io.Reader, cfg AlignedDecoderConfig) (Metadata, []byte, Acquisition, error) {
	raw, err := ReadInt16LE(r)
	if err != nil {
		return Metadata{}, nil, Acquisition{}, err
	}
	return DecodeCaptured(cfg, MonoInt16ToComplex(raw))
}
