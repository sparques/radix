package radix

import "fmt"

// ScramblePayload XORs payload bytes with the protocol scrambler in place.
// Scrambling makes long runs of zeros or ones look noise-like on the air.
func ScramblePayload(data []byte) {
	scrambler := NewXorShift32(0)
	for i := range data {
		data[i] ^= byte(scrambler.Next())
	}
}

// ScrambledPayload returns a scrambled copy of data, leaving data unchanged.
func ScrambledPayload(data []byte) []byte {
	out := append([]byte(nil), data...)
	ScramblePayload(out)
	return out
}

// PayloadCRC computes the CRC used to detect payload decode errors.
func PayloadCRC(data []byte) uint32 {
	crc := NewCRC32(0x8F6E37A0)
	for _, b := range data {
		crc.UpdateByte(b)
	}
	return crc.Sum()
}

// EncodePayload pads, scrambles, CRC-protects, polar-encodes, and interleaves
// payload bytes for the supplied mode configuration.
func EncodePayload(cfg Config, payload []byte) ([]int8, error) {
	if len(cfg.FrozenBits) == 0 {
		return nil, fmt.Errorf("config has no frozen payload table")
	}
	if len(payload) > cfg.DataBytes {
		return nil, fmt.Errorf("payload has %d bytes, mode accepts %d", len(payload), cfg.DataBytes)
	}

	data := make([]byte, cfg.DataBytes)
	copy(data, payload)
	ScramblePayload(data)

	message := make([]int8, cfg.DataBits+32)
	for i := 0; i < cfg.DataBits; i++ {
		message[i] = NRZ(GetLEBit(data, i))
	}
	sum := PayloadCRC(data)
	for i := 0; i < 32; i++ {
		message[cfg.DataBits+i] = NRZ((sum>>i)&1 != 0)
	}

	code, err := PolarEncode(message, cfg.FrozenBits, cfg.CodeOrder)
	if err != nil {
		return nil, err
	}
	return InterleaveEncode(code, cfg.CodeOrder)
}

// DecodePayload reverses EncodePayload and verifies the payload CRC.
// The returned slice is always cfg.DataBytes long, so callers that know the
// original message length should trim padding themselves.
func DecodePayload(cfg Config, code []int8) ([]byte, error) {
	codeBits := 1 << cfg.CodeOrder
	if len(code) != codeBits {
		return nil, fmt.Errorf("payload code has %d symbols, want %d", len(code), codeBits)
	}
	if len(cfg.FrozenBits) == 0 {
		return nil, fmt.Errorf("config has no frozen payload table")
	}
	deinterleaved, err := InterleaveDecode(code, cfg.CodeOrder)
	if err != nil {
		return nil, err
	}
	message, err := PolarDecodeHard(deinterleaved, cfg.FrozenBits, cfg.CodeOrder)
	if err != nil {
		return nil, err
	}
	if len(message) != cfg.DataBits+32 {
		return nil, fmt.Errorf("payload polar message has %d symbols, want %d", len(message), cfg.DataBits+32)
	}

	crc := NewCRC32(0x8F6E37A0)
	data := make([]byte, cfg.DataBytes)
	for i := 0; i < cfg.DataBits; i++ {
		bit := message[i] < 0
		SetLEBit(data, i, bit)
		crc.UpdateBit(bit)
	}
	for i := 0; i < 32; i++ {
		crc.UpdateBit(message[cfg.DataBits+i] < 0)
	}
	if crc.Sum() != 0 {
		return nil, fmt.Errorf("payload CRC mismatch")
	}
	ScramblePayload(data)
	return data, nil
}
