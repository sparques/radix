package radix

type CRC16 struct {
	lut  [256]uint16
	poly uint16
	crc  uint16
}

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

func (c *CRC16) Reset(v uint16) {
	c.crc = v
}

func (c *CRC16) Sum() uint16 {
	return c.crc
}

func (c *CRC16) UpdateBit(data bool) uint16 {
	c.crc = c.update(c.crc, data)
	return c.crc
}

func (c *CRC16) UpdateByte(data byte) uint16 {
	tmp := c.crc ^ uint16(data)
	c.crc = (c.crc >> 8) ^ c.lut[tmp&255]
	return c.crc
}

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

type CRC32 struct {
	lut  [256]uint32
	poly uint32
	crc  uint32
}

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

func (c *CRC32) Reset(v uint32) {
	c.crc = v
}

func (c *CRC32) Sum() uint32 {
	return c.crc
}

func (c *CRC32) UpdateBit(data bool) uint32 {
	c.crc = c.update(c.crc, data)
	return c.crc
}

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
