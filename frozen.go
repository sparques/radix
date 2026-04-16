package radix

import "fmt"

// FrozenPayloadBits returns the frozen-bit table used by the payload polar
// code. Frozen bits are positions reserved by the error-correction code rather
// than carrying your data.
func FrozenPayloadBits(codeOrder int, rate CodeRate) ([]uint32, error) {
	if codeOrder < 11 || codeOrder > 16 {
		return nil, fmt.Errorf("unsupported code order %d", codeOrder)
	}

	switch rate {
	case RateHalf:
		switch codeOrder {
		case 11:
			return Frozen2048_1056, nil
		case 12:
			return Frozen4096_2080, nil
		case 13:
			return Frozen8192_4128, nil
		case 14:
			return Frozen16384_8224, nil
		case 15:
			return Frozen32768_16416, nil
		case 16:
			return Frozen65536_32800, nil
		}
	case RateTwoThirds:
		switch codeOrder {
		case 11:
			return Frozen2048_1400, nil
		case 12:
			return Frozen4096_2768, nil
		case 13:
			return Frozen8192_5504, nil
		case 14:
			return Frozen16384_10976, nil
		case 15:
			return Frozen32768_21920, nil
		case 16:
			return Frozen65536_43808, nil
		}
	case RateThreeQuarters:
		switch codeOrder {
		case 11:
			return Frozen2048_1568, nil
		case 12:
			return Frozen4096_3104, nil
		case 13:
			return Frozen8192_6176, nil
		case 14:
			return Frozen16384_12320, nil
		case 15:
			return Frozen32768_24608, nil
		case 16:
			return Frozen65536_49184, nil
		}
	case RateFiveSixths:
		switch codeOrder {
		case 11:
			return Frozen2048_1736, nil
		case 12:
			return Frozen4096_3440, nil
		case 13:
			return Frozen8192_6848, nil
		case 14:
			return Frozen16384_13664, nil
		case 15:
			return Frozen32768_27296, nil
		case 16:
			return Frozen65536_54560, nil
		}
	default:
		return nil, fmt.Errorf("unsupported code rate %d", rate)
	}

	return nil, fmt.Errorf("unsupported payload frozen table for order %d rate %s", codeOrder, rate)
}

// PayloadMessageBits returns the number of information bits carried before
// polar encoding, including the 32 payload CRC bits.
func PayloadMessageBits(codeOrder int, rate CodeRate) (int, error) {
	dataBits, err := dataBits(codeOrder, rate)
	if err != nil {
		return 0, err
	}
	return dataBits + 32, nil
}
