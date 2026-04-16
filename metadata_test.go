package radix

import "testing"

func TestEncodeMetadataShapeAndValues(t *testing.T) {
	mode, err := NewMode(QAM16, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	call, err := EncodeCallSign("ANONYMOUS")
	if err != nil {
		t.Fatal(err)
	}
	got, err := EncodeMetadata(call, mode)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != MetadataCodeBits {
		t.Fatalf("len(EncodeMetadata) = %d, want %d", len(got), MetadataCodeBits)
	}
	for i, v := range got {
		if v != -1 && v != 1 {
			t.Fatalf("metadata[%d] = %d, want -1 or 1", i, v)
		}
	}

	wantPrefix := []int8{1, 1, -1, 1, -1, -1, 1, 1, -1, 1, 1, -1, -1, -1, 1, -1}
	for i, want := range wantPrefix {
		if got[i] != want {
			t.Fatalf("metadata prefix[%d] = %d, want %d", i, got[i], want)
		}
	}
}

func TestMetadataWordRejectsInvalidInputs(t *testing.T) {
	if _, err := MetadataWord(0, 0); err == nil {
		t.Fatal("MetadataWord accepted zero call sign")
	}
	if _, err := MetadataWord(1, 128); err == nil {
		t.Fatal("MetadataWord accepted analog mode")
	}
}
