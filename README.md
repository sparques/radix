# Radix

Radix is a Go package port of [`aicodix/modem`](https://github.com/aicodix/modem), a simple OFDM modem for transceiving datagrams.

This repository is library-first. It does not provide a standalone encoder or decoder command. The first slice ports the mode metadata and PSK/QAM constellation primitives that later encoder and decoder code will build on.

```go
mode, err := radix.NewMode(radix.QAM16, radix.RateHalf, radix.ShortFrame)
if err != nil {
	panic(err)
}

cfg, err := radix.Setup(mode)
if err != nil {
	panic(err)
}

constellation, err := radix.NewConstellation(cfg.Modulation)
if err != nil {
	panic(err)
}

symbol, err := constellation.Map([]float64{1, -1, 1, -1})
if err != nil {
	panic(err)
}
_ = symbol
```

Helpers are also available for the upstream encoder parameters:

```go
callSign, err := radix.EncodeCallSign("ANONYMOUS")
if err != nil {
	panic(err)
}

if err := radix.ValidateFrequencyOffset(48000, 1, 1500); err != nil {
	panic(err)
}

toneOffset, err := radix.ToneOffset(48000, 1500)
if err != nil {
	panic(err)
}

plan, err := radix.BuildTonePlan(cfg)
if err != nil {
	panic(err)
}

metadata, err := radix.EncodeMetadata(callSign, mode)
if err != nil {
	panic(err)
}

payload := radix.ScrambledPayload([]byte("hello"))
crc := radix.PayloadCRC(payload)

payloadCode, err := radix.EncodePayload(cfg, []byte("hello"))
if err != nil {
	panic(err)
}

toneFrames, err := radix.BuildToneFrames(cfg, metadata, payloadCode)
if err != nil {
	panic(err)
}

_, _, _, _ = toneOffset, plan, crc, toneFrames
```

Audio encoding uses complex analytic float32 samples. Front-ends can write that
directly, adapt it to stereo IQ float32, or take the real component for mono
float32 when using an offset valid for real audio:

```go
samples, err := radix.EncodeComplex(radix.EncoderConfig{
	Audio: radix.AudioConfig{
		SampleRate:      48000,
		FrequencyOffset: 1500,
	},
	Mode:     mode,
	CallSign: callSign,
}, []byte("hello"))
if err != nil {
	panic(err)
}

iq := radix.ComplexToInterleavedFloat32(samples)
mono := radix.ComplexToMonoFloat32(samples)

_, _ = iq, mono
```

For receive-side plumbing, callers can read little-endian complex64 streams and
run aligned OFDM analysis when timing and mode are already known:

```go
frames, err := radix.AnalyzeComplexAlignedFrom(reader, radix.AlignedDecoderConfig{
	Audio: radix.AudioConfig{
		SampleRate:      48000,
		FrequencyOffset: 1500,
	},
	Mode: mode,
})
if err != nil {
	panic(err)
}

_ = frames
```

For clean streams where symbol timing and mode are already known, aligned decode
can recover metadata and the padded mode payload:

```go
metadata, payload, err := radix.DecodeAligned(radix.AlignedDecoderConfig{
	Audio: radix.AudioConfig{
		SampleRate:      48000,
		FrequencyOffset: 1500,
	},
	Mode: mode,
}, samples)
if err != nil {
	panic(err)
}

_, _ = metadata, payload
```

Captured receive can search an arbitrary complex sample buffer for the repeated
Schmidl-Cox preamble, estimate residual carrier frequency and IQ phase, align to
the first data symbol, apply seed-tone equalization, and recover the metadata and
padded payload. The mode is still supplied by the caller:

```go
metadata, payload, acquisition, err := radix.DecodeCaptured(radix.AlignedDecoderConfig{
	Audio: radix.AudioConfig{
		SampleRate:      48000,
		FrequencyOffset: 1500,
	},
	Mode: mode,
}, captured)
if err != nil {
	panic(err)
}

_, _, _ = metadata, payload, acquisition
```

Reader helpers are available for little-endian complex64, stereo float32 IQ,
mono float32, stereo int16 IQ, and mono int16 captures:

```go
metadata, payload, acquisition, err = radix.DecodeInterleavedFloat32CapturedFrom(reader, radix.AlignedDecoderConfig{
	Audio: radix.AudioConfig{
		SampleRate:      48000,
		FrequencyOffset: 1500,
	},
	Mode: mode,
})
```

Real captures still need more receiver work for difficult channels: sample-rate
tracking, stronger channel interpolation, and soft/noisy polar decoding.
