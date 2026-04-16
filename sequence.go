package radix

// XorShiftMask is a small deterministic pseudo-random sequence generator
// limited to a fixed number of low bits.
type XorShiftMask struct {
	bits        uint
	first       uint
	second      uint
	third       uint
	defaultSeed uint32
	mask        uint32
	state       uint32
}

// NewXorShiftMask constructs a masked xorshift sequence generator.
func NewXorShiftMask(bits, first, second, third uint, seed uint32) XorShiftMask {
	mask := uint32(1<<bits) - 1
	return XorShiftMask{
		bits:        bits,
		first:       first,
		second:      second,
		third:       third,
		defaultSeed: seed,
		mask:        mask,
		state:       seed,
	}
}

// Reset returns the generator to seed. Passing zero uses the default seed.
func (x *XorShiftMask) Reset(seed uint32) {
	if seed == 0 {
		seed = x.defaultSeed
	}
	x.state = seed
}

// Next advances the generator and returns the next masked value.
func (x *XorShiftMask) Next() uint32 {
	x.state ^= x.state << x.first
	x.state &= x.mask
	x.state ^= x.state >> x.second
	x.state ^= x.state << x.third
	x.state &= x.mask
	return x.state
}

// XorShift32 is the byte scrambler generator used for payload whitening.
type XorShift32 struct {
	state uint32
}

// NewXorShift32 constructs the payload scrambler generator.
// Passing zero uses the protocol default seed.
func NewXorShift32(seed uint32) XorShift32 {
	if seed == 0 {
		seed = 2463534242
	}
	return XorShift32{state: seed}
}

// Reset returns the generator to seed. Passing zero uses the protocol default.
func (x *XorShift32) Reset(seed uint32) {
	if seed == 0 {
		seed = 2463534242
	}
	x.state = seed
}

// Next advances the generator and returns the next 32-bit value.
func (x *XorShift32) Next() uint32 {
	x.state ^= x.state << 13
	x.state ^= x.state >> 17
	x.state ^= x.state << 5
	return x.state
}

// MLS is a maximal-length sequence generator. Radix uses MLS sequences for
// preamble, seed, and noise-like reference patterns.
type MLS struct {
	poly int
	test int
	reg  int
}

// NewMLS constructs a maximal-length sequence generator.
// Passing zero for poly or reg uses the protocol defaults.
func NewMLS(poly int, reg int) MLS {
	if poly == 0 {
		poly = 0b100000000000000001001
	}
	if reg == 0 {
		reg = 1
	}
	return MLS{poly: poly, test: hiBit(poly) >> 1, reg: reg}
}

// Reset returns the MLS generator to reg. Passing zero uses state 1.
func (m *MLS) Reset(reg int) {
	if reg == 0 {
		reg = 1
	}
	m.reg = reg
}

// Length returns the number of register bits in the MLS polynomial.
func (m *MLS) Length() int {
	return hiBit(m.poly) - 1
}

// Bad reports whether reg appears to produce a short or invalid MLS cycle.
func (m *MLS) Bad(reg int) bool {
	if reg == 0 {
		reg = 1
	}
	m.reg = reg
	length := m.Length()
	for i := 1; i < length; i++ {
		m.NextBit()
		if m.reg == reg {
			return true
		}
	}
	m.NextBit()
	return m.reg != reg
}

// Next advances the MLS generator and returns the new register state.
func (m *MLS) Next() int {
	m.NextBit()
	return m.reg
}

// NextBit advances the MLS generator and returns the output bit.
func (m *MLS) NextBit() bool {
	fb := m.reg&m.test != 0
	m.reg <<= 1
	if fb {
		m.reg ^= m.poly
	}
	return fb
}

func hiBit(n int) int {
	u := uint(n)
	u |= u >> 1
	u |= u >> 2
	u |= u >> 4
	u |= u >> 8
	u |= u >> 16
	if ^uint(0)>>32 != 0 {
		u |= u >> 32
	}
	return int(u ^ (u >> 1))
}
