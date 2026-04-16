package radix

import "testing"

func TestFrozenPayloadBitsCoversAllModes(t *testing.T) {
	for _, rate := range []CodeRate{RateHalf, RateTwoThirds, RateThreeQuarters, RateFiveSixths} {
		for order := 11; order <= 16; order++ {
			frozen, err := FrozenPayloadBits(order, rate)
			if err != nil {
				t.Fatalf("FrozenPayloadBits(%d, %s): %v", order, rate, err)
			}
			if len(frozen) != 1<<(order-5) {
				t.Fatalf("FrozenPayloadBits(%d, %s) len = %d, want %d", order, rate, len(frozen), 1<<(order-5))
			}
			msgBits, err := PayloadMessageBits(order, rate)
			if err != nil {
				t.Fatalf("PayloadMessageBits(%d, %s): %v", order, rate, err)
			}
			if got := countInformationBits(frozen, order); got != msgBits {
				t.Fatalf("FrozenPayloadBits(%d, %s) information bits = %d, want %d", order, rate, got, msgBits)
			}
		}
	}
}

func TestSetupIncludesFrozenPayloadBits(t *testing.T) {
	mode, err := NewMode(QAM16, RateHalf, ShortFrame)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Setup(mode)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.FrozenBits) != len(Frozen4096_2080) {
		t.Fatalf("cfg.FrozenBits len = %d, want %d", len(cfg.FrozenBits), len(Frozen4096_2080))
	}
	if &cfg.FrozenBits[0] != &Frozen4096_2080[0] {
		t.Fatal("cfg.FrozenBits does not reference the QAM16 1/2 short table")
	}
}

func TestMetadataFrozenTableInformationBits(t *testing.T) {
	if got := countInformationBits(Frozen256_72, 8); got != MetadataPolarBits {
		t.Fatalf("Frozen256_72 information bits = %d, want %d", got, MetadataPolarBits)
	}
}

func countInformationBits(frozen []uint32, order int) int {
	var count int
	for i := 0; i < 1<<order; i++ {
		if !frozenBit(frozen, i) {
			count++
		}
	}
	return count
}
