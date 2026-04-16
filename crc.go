package radix

// CRC16 is a small reflected CRC-16 calculator used by Radix metadata.
type CRC16 struct {
	lut  [256]uint16
	poly uint16
	crc  uint16
}

// NewCRC16 builds a CRC-16 calculator for the given reflected polynomial.
func NewCRC16(poly uint16) CRC16 {
	c := CRC16{poly: poly}
	for j := 0; j < 256; j++ {
		tmp := uint16(j)
		for i := 0; i < 8; i++ {
			tmp = c.update(tmp, false)
		}
		c.lut[j] = tmp
	}
	return c
}

// Reset replaces the current CRC-16 state.
func (c *CRC16) Reset(v uint16) {
	c.crc = v
}

// Sum returns the current CRC-16 value.
func (c *CRC16) Sum() uint16 {
	return c.crc
}

// UpdateBit feeds one bit into the CRC-16 and returns the updated value.
func (c *CRC16) UpdateBit(data bool) uint16 {
	c.crc = c.update(c.crc, data)
	return c.crc
}

// UpdateByte feeds one byte into the CRC-16 and returns the updated value.
func (c *CRC16) UpdateByte(data byte) uint16 {
	tmp := c.crc ^ uint16(data)
	c.crc = (c.crc >> 8) ^ c.lut[tmp&255]
	return c.crc
}

// UpdateUint64 feeds a uint64 into the CRC-16, least-significant byte first.
func (c *CRC16) UpdateUint64(data uint64) uint16 {
	for i := 0; i < 8; i++ {
		c.UpdateByte(byte(data >> (8 * i)))
	}
	return c.crc
}

func (c *CRC16) update(prev uint16, data bool) uint16 {
	tmp := prev
	if data {
		tmp ^= 1
	}
	if tmp&1 != 0 {
		return (prev >> 1) ^ c.poly
	}
	return prev >> 1
}

// CRC32 is a small reflected CRC-32 calculator used by Radix payloads.
type CRC32 struct {
	lut  [256]uint32
	poly uint32
	crc  uint32
}

// NewCRC32 builds a CRC-32 calculator for the given reflected polynomial.
func NewCRC32(poly uint32) CRC32 {
	c := CRC32{poly: poly}
	for j := 0; j < 256; j++ {
		tmp := uint32(j)
		for i := 0; i < 8; i++ {
			tmp = c.update(tmp, false)
		}
		c.lut[j] = tmp
	}
	return c
}

// Reset replaces the current CRC-32 state.
func (c *CRC32) Reset(v uint32) {
	c.crc = v
}

// Sum returns the current CRC-32 value.
func (c *CRC32) Sum() uint32 {
	return c.crc
}

// UpdateBit feeds one bit into the CRC-32 and returns the updated value.
func (c *CRC32) UpdateBit(data bool) uint32 {
	c.crc = c.update(c.crc, data)
	return c.crc
}

// UpdateByte feeds one byte into the CRC-32 and returns the updated value.
func (c *CRC32) UpdateByte(data byte) uint32 {
	tmp := c.crc ^ uint32(data)
	c.crc = (c.crc >> 8) ^ c.lut[tmp&255]
	return c.crc
}

func (c *CRC32) update(prev uint32, data bool) uint32 {
	tmp := prev
	if data {
		tmp ^= 1
	}
	if tmp&1 != 0 {
		return (prev >> 1) ^ c.poly
	}
	return prev >> 1
}
