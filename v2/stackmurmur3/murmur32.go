package stackmurmur3

import (
	"math/bits"
)

const (
	c1_32 uint32 = 0xcc9e2d51
	c2_32 uint32 = 0x1b873593
)

// Digest32 represents a partial evaluation of a 32 bit hash.
type Digest32 struct {
	tailBuf [4]byte
	tailIdx int    // Length of tail stored in tailBuf.
	clen    int    // Digested input cumulative length.
	h1      uint32 // running hash.
}

// New32WithSeed returns a Digest32 for streaming 32 bit sums with its internal
// digest initialized to seed.
//
// This reads and processes the data in chunks of little endian uint32s;
// thus, the returned hash is portable across architectures.
func New32WithSeed(seed uint32) *Digest32 {
	return &Digest32{h1: seed}
}

// New32 returns a hash.Hash32 for streaming 32 bit sums.
func New32() *Digest32 {
	return New32WithSeed(0)
}

func (d *Digest32) Write(p []byte) {
	n := len(p)
	d.clen += n

	// If tail is not empty, must process it before rest of payload
	if d.tailIdx > 0 {
		// Stick back pending bytes.
		nfree := len(d.tailBuf) - d.tailIdx // nfree ∈ [1, len(d.tailBuf)-1].
		if nfree > n {
			// Tail + payload size smaller than a block, can't perform bmix
			d.tailIdx += copy(d.tailBuf[d.tailIdx:], p)
			return
		}
		// Expanded tail to one full block
		copy(d.tailBuf[d.tailIdx:], p[:nfree])
		_ = d.bmix32(d.tailBuf[:])
		// Process rest of the payload
		p = p[nfree:]
		d.tailIdx = 0
	}

	p = d.bmix32(p)
	// Keep own copy of the 0 to Size()-1 pending bytes.
	d.tailIdx += copy(d.tailBuf[d.tailIdx:], p)

	return
}

// Sum finalizes the hash and writes it out to a byte slice
func (d Digest32) Sum(b []byte) []byte {
	h := d.Sum32()
	return append(b, byte(h>>24), byte(h>>16), byte(h>>8), byte(h))
}

// Digest all blocks, return the tail
func (d *Digest32) bmix32(p []byte) []byte {
	var h1 = d.h1
	for len(p) >= 4 {
		k1 := uint32(p[0]) | uint32(p[1])<<8 | uint32(p[2])<<16 | uint32(p[3])<<24
		p = p[4:]

		k1 *= c1_32
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2_32

		h1 ^= k1
		h1 = bits.RotateLeft32(h1, 13)
		h1 = h1*5 + 0xe6546b64
	}
	d.h1 = h1
	return p
}

// Sum32 finalizes the hash
func (d Digest32) Sum32() (h1 uint32) {
	h1 = d.h1
	var k1 uint32
	switch d.tailIdx & 3 {
	case 3:
		k1 ^= uint32(d.tailBuf[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(d.tailBuf[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(d.tailBuf[0])
		k1 *= c1_32
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2_32
		h1 ^= k1
	}

	h1 ^= uint32(d.clen)

	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}
