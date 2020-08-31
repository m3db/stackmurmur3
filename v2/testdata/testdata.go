package testdata

import (
	"math/rand"
	"unsafe"
)

var IsLittleEndian = func() bool {
	i := uint16(1)
	return (*(*[2]byte)(unsafe.Pointer(&i)))[0] == 1
}()

var ReferenceHashes = []struct {
	H32   uint32
	H64_1 uint64
	H64_2 uint64
	S     string
}{
	{0x00000000, 0x0000000000000000, 0x0000000000000000, ""},
	{0x248bfa47, 0xcbd8a7b341bd9b02, 0x5b1e906a48ae1d19, "hello"},
	{0x149bbb7f, 0x342fac623a5ebc8e, 0x4cdcbc079642414d, "hello, world"},
	{0xe31e8a70, 0xb89e5988b737affc, 0x664fc2950231b2cb, "19 Jan 2038 at 3:14:07 AM"},
	{0xd5c48bfc, 0xcd99481f9ee902c9, 0x695da1a38987b6e7, "The quick brown fox jumps over the lazy dog."},
}

func RandBytes(n int) []byte {
	var (
		rnd = rand.New(rand.NewSource(0))
		b   = make([]byte, n)
	)
	for i := range b {
		b[i] = byte(rnd.Intn(256))
	}
	return b
}
