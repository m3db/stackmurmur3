package stackmurmur3

// Digest64 is half a Digest128.
type Digest64 Digest128

// New64WithSeed returns a Digest64 for streaming 64 bit sums. As the canonical
// implementation does not support Sum64, this uses New128WithSeed(seed, seed)
func New64WithSeed(seed uint64) *Digest64 {
	return (*Digest64)(New128WithSeed(seed, seed))
}

// New64 returns a Digest64 for streaming 64 bit sums.
func New64() *Digest64 {
	return New64WithSeed(0)
}

func (d *Digest64) Write(p []byte) {
	(*Digest128)(d).Write(p)
}

// Sum finalizes the hash and writes it out to a byte slice
func (d Digest64) Sum(b []byte) []byte {
	h1 := d.Sum64()
	return append(b,
		byte(h1>>56), byte(h1>>48), byte(h1>>40), byte(h1>>32),
		byte(h1>>24), byte(h1>>16), byte(h1>>8), byte(h1))
}

// Sum64 finalizes the hash
func (d Digest64) Sum64() uint64 {
	h1, _ := Digest128(d).Sum128()
	return h1
}
