package radix

import "testing"

func TestScramblePayloadIsSelfInverse(t *testing.T) {
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	orig := append([]byte(nil), data...)
	ScramblePayload(data)
	if string(data) == string(orig) {
		t.Fatal("ScramblePayload did not change payload")
	}
	ScramblePayload(data)
	if string(data) != string(orig) {
		t.Fatalf("second ScramblePayload = %v, want %v", data, orig)
	}
}

func TestScramblePayloadMatchesXorShift32LowByte(t *testing.T) {
	got := ScrambledPayload([]byte{0, 0, 0, 0, 0})
	want := []byte{99, 122, 160, 126, 225}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("scrambled[%d] = %#x, want %#x", i, got[i], want[i])
		}
	}
}

func TestPayloadCRCResidue(t *testing.T) {
	data := ScrambledPayload([]byte{1, 2, 3, 4, 5})
	sum := PayloadCRC(data)

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

func TestEncodePayloadShape(t *testing.T) {
	mode, err := NewMode(QAM16, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Setup(mode)
	if err != nil {
		t.Fatal(err)
	}
	code, err := EncodePayload(cfg, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 1<<cfg.CodeOrder {
		t.Fatalf("EncodePayload len = %d, want %d", len(code), 1<<cfg.CodeOrder)
	}
	for i, bit := range code {
		if bit != -1 && bit != 1 {
			t.Fatalf("code[%d] = %d, want -1 or 1", i, bit)
		}
	}
}

func TestEncodePayloadRejectsOversizedData(t *testing.T) {
	mode, err := NewMode(BPSK, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Setup(mode)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := EncodePayload(cfg, make([]byte, cfg.DataBytes+1)); err == nil {
		t.Fatal("EncodePayload accepted oversized data")
	}
}
