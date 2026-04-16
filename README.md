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
