package stackmurmur3

import (
	"math/bits"
)

const (
	c1_128 = 0x87c37b91114253d5
	c2_128 = 0x4cf5ad432745937f
)

// Digest128 represents a partial evaluation of a 128 bit hash.
type Digest128 struct {
	tailBuf [16]byte
	tailIdx int    // Length of tail stored in tailBuf.
	clen    int    // Digested input cumulative length.
	h1      uint64 // running hash part 1.
	h2      uint64 // running hash part 2.
}

// New128WithSeed returns a Digest128 for streaming 128 bit sums with its internal
// digests initialized to seed1 and seed2.
//
// The canonical implementation allows one only uint32 seed; to imitate that
// behavior, use the same, uint32-max seed for seed1 and seed2.
func New128WithSeed(seed1, seed2 uint64) *Digest128 {
	return &Digest128{h1: seed1, h2: seed2}
}

// New128 returns a Digest128 for streaming 128 bit sums.
func New128() *Digest128 {
	return New128WithSeed(0, 0)
}

func (d *Digest128) Write(p []byte) {
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
		_ = d.bmix128(d.tailBuf[:])
		// Process rest of the payload
		p = p[nfree:]
		d.tailIdx = 0
	}

	p = d.bmix128(p)
	// Keep own copy of the 0 to Size()-1 pending bytes.
	d.tailIdx += copy(d.tailBuf[d.tailIdx:], p)

	return
}

// Sum finalizes the hash and writes it out to a byte slice
func (d Digest128) Sum(b []byte) []byte {
	h1, h2 := d.Sum128()
	return append(b,
		byte(h1>>56), byte(h1>>48), byte(h1>>40), byte(h1>>32),
		byte(h1>>24), byte(h1>>16), byte(h1>>8), byte(h1),

		byte(h2>>56), byte(h2>>48), byte(h2>>40), byte(h2>>32),
		byte(h2>>24), byte(h2>>16), byte(h2>>8), byte(h2),
	)
}

// Digest all blocks, return the tail
func (d *Digest128) bmix128(p []byte) []byte {
	var h1, h2 = d.h1, d.h2

	for len(p) >= 16 {
		k1 := uint64(p[0]) | uint64(p[1])<<8 | uint64(p[2])<<16 | uint64(p[3])<<24 | uint64(p[4])<<32 | uint64(p[5])<<40 | uint64(p[6])<<48 | uint64(p[7])<<56
		k2 := uint64(p[8]) | uint64(p[9])<<8 | uint64(p[10])<<16 | uint64(p[11])<<24 | uint64(p[12])<<32 | uint64(p[13])<<40 | uint64(p[14])<<48 | uint64(p[15])<<56
		p = p[16:]

		k1 *= c1_128
		k1 = bits.RotateLeft64(k1, 31)
		k1 *= c2_128
		h1 ^= k1

		h1 = bits.RotateLeft64(h1, 27)
		h1 += h2
		h1 = h1*5 + 0x52dce729

		k2 *= c2_128
		k2 = bits.RotateLeft64(k2, 33)
		k2 *= c1_128
		h2 ^= k2

		h2 = bits.RotateLeft64(h2, 31)
		h2 += h1
		h2 = h2*5 + 0x38495ab5
	}

	d.h1, d.h2 = h1, h2
	return p
}

// Sum128 finalizes the hash
func (d Digest128) Sum128() (h1, h2 uint64) {
	h1, h2 = d.h1, d.h2

	var k1, k2 uint64
	switch d.tailIdx & 15 {
	case 15:
		k2 ^= uint64(d.tailBuf[14]) << 48
		fallthrough
	case 14:
		k2 ^= uint64(d.tailBuf[13]) << 40
		fallthrough
	case 13:
		k2 ^= uint64(d.tailBuf[12]) << 32
		fallthrough
	case 12:
		k2 ^= uint64(d.tailBuf[11]) << 24
		fallthrough
	case 11:
		k2 ^= uint64(d.tailBuf[10]) << 16
		fallthrough
	case 10:
		k2 ^= uint64(d.tailBuf[9]) << 8
		fallthrough
	case 9:
		k2 ^= uint64(d.tailBuf[8]) << 0

		k2 *= c2_128
		k2 = bits.RotateLeft64(k2, 33)
		k2 *= c1_128
		h2 ^= k2

		fallthrough

	case 8:
		k1 ^= uint64(d.tailBuf[7]) << 56
		fallthrough
	case 7:
		k1 ^= uint64(d.tailBuf[6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(d.tailBuf[5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(d.tailBuf[4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(d.tailBuf[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(d.tailBuf[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(d.tailBuf[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint64(d.tailBuf[0]) << 0
		k1 *= c1_128
		k1 = bits.RotateLeft64(k1, 31)
		k1 *= c2_128
		h1 ^= k1
	}

	cl := uint64(d.clen)
	h1 ^= cl
	h2 ^= cl

	h1 += h2
	h2 += h1

	h1 = fmix64(h1)
	h2 = fmix64(h2)

	h1 += h2
	h2 += h1

	return h1, h2
}

func fmix64(k uint64) uint64 {
	k ^= k >> 33
	k *= 0xff51afd7ed558ccd
	k ^= k >> 33
	k *= 0xc4ceb9fe1a85ec53
	k ^= k >> 33
	return k
}
