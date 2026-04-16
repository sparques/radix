package radix

import "fmt"

const (
	MetadataBits      = 56
	MetadataCRCBits   = 16
	MetadataCodeBits  = 256
	MetadataPolarBits = MetadataBits + MetadataCRCBits
)

func MetadataWord(callSign int64, mode Mode) (uint64, error) {
	if callSign <= 0 || callSign >= MaxCallSign {
		return 0, fmt.Errorf("unsupported call sign value %d", callSign)
	}
	if _, err := Setup(mode); err != nil {
		return 0, err
	}
	return (uint64(callSign) << 8) | uint64(mode), nil
}

func EncodeMetadata(callSign int64, mode Mode) ([]int8, error) {
	word, err := MetadataWord(callSign, mode)
	if err != nil {
		return nil, err
	}

	message := make([]int8, MetadataPolarBits)
	for i := 0; i < MetadataBits; i++ {
		message[i] = NRZ((word>>i)&1 != 0)
	}

	crc := NewCRC16(0xA8F4)
	crc.UpdateUint64(word << 8)
	sum := crc.Sum()
	for i := 0; i < MetadataCRCBits; i++ {
		message[MetadataBits+i] = NRZ((sum>>i)&1 != 0)
	}

	code, err := PolarEncode(message, Frozen256_72, 8)
	if err != nil {
		return nil, err
	}
	return InterleaveEncode(code, 8)
}
