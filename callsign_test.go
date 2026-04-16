package radix

import "testing"

func TestCallSignBase40(t *testing.T) {
	tests := []struct {
		text  string
		value int64
	}{
		{"A", 14},
		{"ANONYMOUS", 96291631790192},
		{"N0CALL", 2776087425},
		{"K/TEST", 2467422113},
	}

	for _, tt := range tests {
		got, err := EncodeCallSign(tt.text)
		if err != nil {
			t.Fatalf("EncodeCallSign(%q): %v", tt.text, err)
		}
		if got != tt.value {
			t.Errorf("EncodeCallSign(%q) = %d, want %d", tt.text, got, tt.value)
		}
		decoded, err := DecodeCallSign(got, len(tt.text))
		if err != nil {
			t.Fatalf("DecodeCallSign(%d, %d): %v", got, len(tt.text), err)
		}
		if decoded != tt.text {
			t.Errorf("DecodeCallSign(%d, %d) = %q, want %q", got, len(tt.text), decoded, tt.text)
		}
	}
}

func TestCallSignRejectsUnsupportedValues(t *testing.T) {
	if _, err := EncodeCallSign("NO-DASH"); err == nil {
		t.Fatal("EncodeCallSign accepted unsupported character")
	}
	if _, err := EncodeCallSign("          "); err == nil {
		t.Fatal("EncodeCallSign accepted zero value")
	}
	if _, err := DecodeCallSign(40, 1); err == nil {
		t.Fatal("DecodeCallSign accepted value that does not fit")
	}
}
