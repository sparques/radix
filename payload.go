package radix

import "fmt"

func ScramblePayload(data []byte) {
	scrambler := NewXorShift32(0)
	for i := range data {
		data[i] ^= byte(scrambler.Next())
	}
}

func ScrambledPayload(data []byte) []byte {
	out := append([]byte(nil), data...)
	ScramblePayload(out)
	return out
}

func PayloadCRC(data []byte) uint32 {
	crc := NewCRC32(0x8F6E37A0)
	for _, b := range data {
		crc.UpdateByte(b)
	}
	return crc.Sum()
}

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
