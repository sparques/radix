package radix

import "testing"

func TestCRC16MetadataResidue(t *testing.T) {
	mode, err := NewMode(QAM16, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("ANONYMOUS")
	if err != nil {
		t.Fatal(err)
	}
	word, err := MetadataWord(call, mode)
	if err != nil {
		t.Fatal(err)
	}

	crc := NewCRC16(0xA8F4)
	crc.UpdateUint64(word << 8)
	sum := crc.Sum()

	check := NewCRC16(0xA8F4)
	for i := 0; i < MetadataBits; i++ {
		check.UpdateBit((word>>i)&1 != 0)
	}
	for i := 0; i < MetadataCRCBits; i++ {
		check.UpdateBit((sum>>i)&1 != 0)
	}
	if check.Sum() != 0 {
		t.Fatalf("metadata CRC residue = %#x, want 0", check.Sum())
	}
}

func TestCRC32PayloadResidue(t *testing.T) {
	data := []byte{0x10, 0x20, 0x30, 0x40}
	crc := NewCRC32(0x8F6E37A0)
	for _, b := range data {
		crc.UpdateByte(b)
	}
	sum := crc.Sum()

	check := NewCRC32(0x8F6E37A0)
	for _, b := range data {
		check.UpdateByte(b)
	}
	for i := 0; i < 32; i++ {
		check.UpdateBit((sum>>i)&1 != 0)
	}
	if check.Sum() != 0 {
		t.Fatalf("payload CRC residue = %#x, want 0", check.Sum())
	}
}
