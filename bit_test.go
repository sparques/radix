package radix

import "testing"

func TestBitAccessors(t *testing.T) {
	buf := []byte{0, 0}
	SetBEBit(buf, 0, true)
	SetBEBit(buf, 9, true)
	if buf[0] != 0x80 || buf[1] != 0x40 {
		t.Fatalf("SetBEBit produced %08b %08b", buf[0], buf[1])
	}
	if !GetBEBit(buf, 0) || !GetBEBit(buf, 9) || GetBEBit(buf, 8) {
		t.Fatal("GetBEBit returned unexpected values")
	}
	XorBEBit(buf, 9, true)
	if GetBEBit(buf, 9) {
		t.Fatal("XorBEBit did not clear bit")
	}

	buf = []byte{0, 0}
	SetLEBit(buf, 0, true)
	SetLEBit(buf, 9, true)
	if buf[0] != 0x01 || buf[1] != 0x02 {
		t.Fatalf("SetLEBit produced %08b %08b", buf[0], buf[1])
	}
	if !GetLEBit(buf, 0) || !GetLEBit(buf, 9) || GetLEBit(buf, 8) {
		t.Fatal("GetLEBit returned unexpected values")
	}
	XorLEBit(buf, 9, true)
	if GetLEBit(buf, 9) {
		t.Fatal("XorLEBit did not clear bit")
	}
}
