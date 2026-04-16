package radix

import "fmt"

// MaxCallSign is one past the largest packed call-sign value accepted by the
// modem metadata field.
const MaxCallSign = 262144000000000

// EncodeCallSign packs a call sign into the numeric form stored in frame
// metadata. Letters are case-insensitive; supported characters are space, slash,
// digits, and A-Z.
func EncodeCallSign(s string) (int64, error) {
	var acc int64
	for _, c := range s {
		acc *= 40
		switch {
		case c == ' ':
		case c == '/':
			acc += 3
		case c >= '0' && c <= '9':
			acc += int64(c-'0') + 4
		case c >= 'A' && c <= 'Z':
			acc += int64(c-'A') + 14
		case c >= 'a' && c <= 'z':
			acc += int64(c-'a') + 14
		default:
			return 0, fmt.Errorf("unsupported call sign character %q", c)
		}
	}
	if acc <= 0 || acc >= MaxCallSign {
		return 0, fmt.Errorf("unsupported call sign %q", s)
	}
	return acc, nil
}

// DecodeCallSign unpacks a numeric call-sign value into a fixed-length string.
// Use the same length you want displayed by your application.
func DecodeCallSign(value int64, length int) (string, error) {
	if value < 0 {
		return "", fmt.Errorf("negative call sign value %d", value)
	}
	if length < 0 {
		return "", fmt.Errorf("negative call sign length %d", length)
	}

	const alphabet = "   /0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	out := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		out[i] = alphabet[value%40]
		value /= 40
	}
	if value != 0 {
		return "", fmt.Errorf("call sign value does not fit in %d characters", length)
	}
	return string(out), nil
}
