package stackmurmur3

import (
	"encoding/binary"
	"math/rand"
	"runtime"
	"strconv"
	"testing"
	"testing/quick"

	murmur3 "github.com/m3db/stackmurmur3/v2"
	"github.com/m3db/stackmurmur3/v2/testdata"
	"github.com/stretchr/testify/assert"
)

var (
	quickcheckConfig = &quick.Config{MaxCount: 50000}
	DoNotOptimize32  uint32
	DoNotOptimize128 [2]uint64
)

func TestIncremental(t *testing.T) {
	for _, elem := range testdata.ReferenceHashes {
		h32 := New32()
		h128 := New128()
		for i, j, k := 0, 0, len(elem.S); i < k; i = j {
			j = 2*i + 3
			if j > k {
				j = k
			}
			s := elem.S[i:j]
			h32.Write([]byte(s))
			h128.Write([]byte(s))
		}
		if v := h32.Sum32(); v != elem.H32 {
			t.Errorf("'%s': 0x%x (want 0x%x)\n", elem.S, v, elem.H32)
		}
		if v1, v2 := h128.Sum128(); v1 != elem.H64_1 || v2 != elem.H64_2 {
			t.Errorf("'%s': 0x%x-0x%x (want 0x%x-0x%x)\n", elem.S, v1, v2, elem.H64_1, elem.H64_2)
		}
	}
}

func TestQuickSum32(t *testing.T) {
	f := func(data []byte) bool {
		goh1 := murmur3.Sum32(data)
		goh2 := murmur3.StringSum32(string(data))
		cpph1 := goh1
		if testdata.IsLittleEndian {
			cpph1 = testdata.SeedSum32(0, data)
		}
		return goh1 == goh2 && goh1 == cpph1
	}
	if err := quick.Check(f, quickcheckConfig); err != nil {
		t.Error(err)
	}
}

func TestQuickSeedSum32(t *testing.T) {
	f := func(seed uint32, data []byte) bool {
		goh1 := murmur3.SeedSum32(seed, data)
		goh2 := murmur3.SeedStringSum32(seed, string(data))
		goh3 := func() uint32 {
			h := New32WithSeed(seed)
			if len(data) > 0 {
				split := int(seed % uint32(len(data)))
				h.Write(data[:split])
				h.Write(data[split:])
			} else {
				h.Write(data)
			}
			return binary.BigEndian.Uint32(h.Sum(nil))
		}()
		cpph1 := goh1
		if testdata.IsLittleEndian {
			cpph1 = testdata.SeedSum32(seed, data)
		}
		return goh1 == goh2 && goh1 == goh3 && goh1 == cpph1
	}
	if err := quick.Check(f, quickcheckConfig); err != nil {
		t.Error(err)
	}
}

func TestQuickSum64(t *testing.T) {
	f := func(data []byte) bool {
		goh1 := murmur3.Sum64(data)
		goh2 := murmur3.StringSum64(string(data))
		cpph1 := goh1
		if testdata.IsLittleEndian {
			cpph1 = testdata.SeedSum64(0, data)
		}
		return goh1 == goh2 && goh1 == cpph1
	}
	if err := quick.Check(f, quickcheckConfig); err != nil {
		t.Error(err)
	}
}

func TestQuickSeedSum64(t *testing.T) {
	f := func(seed uint32, data []byte) bool {
		goh1 := murmur3.SeedSum64(uint64(seed), data)
		goh2 := murmur3.SeedStringSum64(uint64(seed), string(data))
		goh3 := func() uint64 {
			h := New64WithSeed(uint64(seed))
			h.Write(data)
			return binary.BigEndian.Uint64(h.Sum(nil))
		}()
		cpph1 := goh1
		if testdata.IsLittleEndian {
			cpph1 = testdata.SeedSum64(seed, data)
		}
		return goh1 == goh2 && goh1 == goh3 && goh1 == cpph1
	}
	if err := quick.Check(f, quickcheckConfig); err != nil {
		t.Error(err)
	}
}

func TestQuickSum128(t *testing.T) {
	f := func(data []byte) bool {
		goh1, goh2 := murmur3.Sum128(data)
		goh3, goh4 := murmur3.StringSum128(string(data))
		cpph1, cpph2 := goh1, goh2
		if testdata.IsLittleEndian {
			cpph1, cpph2 = testdata.SeedSum128(0, data)
		}
		return goh1 == goh3 && goh2 == goh4 && goh1 == cpph1 && goh2 == cpph2
	}
	if err := quick.Check(f, quickcheckConfig); err != nil {
		t.Error(err)
	}
}

func TestQuickSeedSum128(t *testing.T) {
	f := func(seed uint32, data []byte) bool {
		goh1, goh2 := murmur3.SeedSum128(uint64(seed), uint64(seed), data)
		goh3, goh4 := murmur3.SeedStringSum128(uint64(seed), uint64(seed), string(data))
		goh5, goh6 := func() (uint64, uint64) {
			h := New128WithSeed(uint64(seed), uint64(seed))
			if len(data) > 0 {
				split := int(seed % uint32(len(data)))
				h.Write(data[:split])
				h.Write(data[split:])
			} else {
				h.Write(data)
			}
			sum := h.Sum(nil)
			return binary.BigEndian.Uint64(sum), binary.BigEndian.Uint64(sum[8:])
		}()
		cpph1, cpph2 := goh1, goh2
		if testdata.IsLittleEndian {
			testdata.SeedSum128(seed, data)
		}
		return goh1 == goh3 && goh2 == goh4 &&
			goh1 == goh5 && goh2 == goh6 &&
			goh1 == cpph1 && goh2 == cpph2
	}
	if err := quick.Check(f, quickcheckConfig); err != nil {
		t.Error(err)
	}
}

// go1.14 showed that doing *(*uint32)(unsafe.Pointer(&data[i*4])) was unsafe
// due to alignment issues; this test ensures that we will always catch that.
func TestUnaligned(t *testing.T) {
	in1 := []byte("abcdefghijklmnopqrstuvwxyz")
	in2 := []byte("_abcdefghijklmnopqrstuvwxyz")

	t.Run("Digest32", func(t *testing.T) {
		sum1 := func() uint32 { n := New32(); n.Write(in1); return n.Sum32() }()
		sum2 := func() uint32 { n := New32(); n.Write(in2[1:]); return n.Sum32() }()
		assert.EqualValues(t, sum1, sum2)
	})

	t.Run("Digest64", func(t *testing.T) {
		sum1 := func() uint64 { n := New64(); n.Write(in1); return n.Sum64() }()
		sum2 := func() uint64 { n := New64(); n.Write(in2[1:]); return n.Sum64() }()
		assert.EqualValues(t, sum1, sum2)
	})

	t.Run("Digest128", func(t *testing.T) {
		sum1a, sum1b := func() (uint64, uint64) { n := New128(); n.Write(in1); return n.Sum128() }()
		sum2a, sum2b := func() (uint64, uint64) { n := New128(); n.Write(in2[1:]); return n.Sum128() }()
		assert.EqualValues(t, sum1a, sum2a)
		assert.EqualValues(t, sum1b, sum2b)
	})
}

// Our lengths force 1) the function base itself (no loop/tail), 2) remainders
// and 3) the loop itself.

func BenchmarkPartial32Branches(b *testing.B) {
	for length := 0; length <= 4; length++ {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf := make([]byte, length)
			b.SetBytes(int64(length))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hasher := New32()
				hasher.Write(buf)
				DoNotOptimize32 = hasher.Sum32()
			}
		})
	}
}

func BenchmarkPartial128Branches(b *testing.B) {
	for length := 0; length <= 16; length++ {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf := make([]byte, length)
			b.SetBytes(int64(length))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hasher := New128()
				hasher.Write(buf)
				DoNotOptimize128[0], DoNotOptimize128[1] = hasher.Sum128()
			}
		})
	}
}

// Sizes below pick up where branches left off to demonstrate speed at larger
// slice sizes.

func BenchmarkPartial32Sizes(b *testing.B) {
	buf := testdata.RandBytes(8192)
	for length := 32; length <= cap(buf); length *= 2 {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf = buf[:length]
			b.SetBytes(int64(length))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hasher := New32()
				hasher.Write(buf)
				DoNotOptimize32 = hasher.Sum32()
			}
		})
	}
}

func BenchmarkPartial64Sizes(b *testing.B) {
	buf := testdata.RandBytes(8192)
	for length := 32; length <= cap(buf); length *= 2 {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf = buf[:length]
			b.SetBytes(int64(length))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hasher := New64()
				hasher.Write(buf)
				DoNotOptimize128[0] = hasher.Sum64()
			}
		})
	}
}

func BenchmarkPartial128Sizes(b *testing.B) {
	buf := testdata.RandBytes(8192)
	for length := 32; length <= cap(buf); length *= 2 {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf = buf[:length]
			b.SetBytes(int64(length))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hasher := New128()
				hasher.Write(buf)
				DoNotOptimize128[0], DoNotOptimize128[1] = hasher.Sum128()
			}
		})
	}
}

// --

func TestDigest32ZeroAllocWrite(t *testing.T) {
	d := make([]byte, 128)
	for i := range d {
		d[i] = byte(i)
	}

	var (
		h     = New32WithSeed(uint32(rand.Int()))
		buf   = make([]byte, 4096)
		stats runtime.MemStats
	)
	runtime.ReadMemStats(&stats)
	startAllocs := stats.Mallocs

	for i := 0; i < 1000; i++ {
		n, err := rand.Read(buf)
		if err != nil {
			t.FailNow()
		}
		h.Write(buf[:n])
	}

	runtime.ReadMemStats(&stats)
	endAllocs := stats.Mallocs
	assert.Equal(t, startAllocs, endAllocs)
}

func TestDigest64ZeroAllocWrite(t *testing.T) {
	d := make([]byte, 128)
	for i := range d {
		d[i] = byte(i)
	}

	var (
		h     = New64WithSeed(uint64(rand.Int()))
		buf   = make([]byte, 4096)
		stats runtime.MemStats
	)
	runtime.ReadMemStats(&stats)
	startAllocs := stats.Mallocs

	for i := 0; i < 1000; i++ {
		n, err := rand.Read(buf)
		if err != nil {
			t.FailNow()
		}
		h.Write(buf[:n])
	}

	runtime.ReadMemStats(&stats)
	endAllocs := stats.Mallocs
	assert.Equal(t, startAllocs, endAllocs)
}

func TestDigest128ZeroAllocWrite(t *testing.T) {
	d := make([]byte, 128)
	for i := range d {
		d[i] = byte(i)
	}

	var (
		h     = New128WithSeed(uint64(rand.Int()), uint64(rand.Int()))
		buf   = make([]byte, 4096)
		stats runtime.MemStats
	)
	runtime.ReadMemStats(&stats)
	startAllocs := stats.Mallocs

	for i := 0; i < 1000; i++ {
		n, err := rand.Read(buf)
		if err != nil {
			t.FailNow()
		}
		h.Write(buf[:n])
	}

	runtime.ReadMemStats(&stats)
	endAllocs := stats.Mallocs
	assert.Equal(t, startAllocs, endAllocs)
}
