package radix

import "fmt"

const (
	// MetadataBits is the number of uncoded metadata bits before the metadata
	// CRC is appended.
	MetadataBits = 56
	// MetadataCRCBits is the number of CRC bits protecting metadata.
	MetadataCRCBits = 16
	// MetadataCodeBits is the number of transmitted metadata symbols after
	// coding and interleaving.
	MetadataCodeBits = 256
	// MetadataPolarBits is the number of metadata bits fed into the polar code.
	MetadataPolarBits = MetadataBits + MetadataCRCBits
)

// MetadataWord packs a call sign value and mode into the uncoded metadata word.
func MetadataWord(callSign int64, mode Mode) (uint64, error) {
	if callSign <= 0 || callSign >= MaxCallSign {
		return 0, fmt.Errorf("unsupported call sign value %d", callSign)
	}
	if _, err := Setup(mode); err != nil {
		return 0, err
	}
	return (uint64(callSign) << 8) | uint64(mode), nil
}

// EncodeMetadata encodes the call sign and mode into signed metadata symbols
// ready for tone mapping.
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

// Metadata is the decoded frame header. It tells you who transmitted and which
// mode was used for the payload.
type Metadata struct {
	// CallSignValue is the packed call sign. Use DecodeCallSign to display it.
	CallSignValue int64
	// Mode is the mode announced by the transmitter.
	Mode Mode
	// Word is the raw decoded metadata word.
	Word uint64
}

// DecodeMetadata decodes signed metadata symbols and verifies their CRC.
func DecodeMetadata(code []int8) (Metadata, error) {
	if len(code) != MetadataCodeBits {
		return Metadata{}, fmt.Errorf("metadata has %d symbols, want %d", len(code), MetadataCodeBits)
	}
	deinterleaved, err := InterleaveDecode(code, 8)
	if err != nil {
		return Metadata{}, err
	}
	message, err := PolarDecodeHard(deinterleaved, Frozen256_72, 8)
	if err != nil {
		return Metadata{}, err
	}
	if len(message) != MetadataPolarBits {
		return Metadata{}, fmt.Errorf("metadata polar message has %d symbols, want %d", len(message), MetadataPolarBits)
	}

	var word uint64
	crc := NewCRC16(0xA8F4)
	for i := 0; i < MetadataBits; i++ {
		bit := message[i] < 0
		if bit {
			word |= 1 << i
		}
		crc.UpdateBit(bit)
	}
	for i := 0; i < MetadataCRCBits; i++ {
		crc.UpdateBit(message[MetadataBits+i] < 0)
	}
	if crc.Sum() != 0 {
		return Metadata{}, fmt.Errorf("metadata CRC mismatch")
	}

	call := int64(word >> 8)
	if call <= 0 || call >= MaxCallSign {
		return Metadata{}, fmt.Errorf("unsupported call sign value %d", call)
	}
	mode := Mode(word & 255)
	if _, err := Setup(mode); err != nil {
		return Metadata{}, err
	}
	return Metadata{CallSignValue: call, Mode: mode, Word: word}, nil
}
