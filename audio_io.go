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
