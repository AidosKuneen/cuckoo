// Copyright (c) 2018 Aidos Developer

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cuckoo

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"runtime"
	"testing"

	"github.com/AidosKuneen/numcpu"
)

/*
Looking for 20-cycle on cuckoo25("",77) with 50% edges
Using 99MB bucket memory at 7f0155310010,
1x2810KB thread memory at 7f0155071010,
1-way siphash, and 32 buckets.
k0 k1 f4956dc403730b01 e6d45de39c2a5a3e
k0:73006e74 k1:f24e3553  k2:6d167960 k3:e55e3f4d kk0:3730b01,kk1:9c2a5a3e %kk2:3730b01 %kk3:9c2a5a3e
nonce 77
genUnodes round  0 size 16777216 rdtsc: 601295224
genVnodes round  1 size 10603535 rdtsc: 668146100
trimedges id 0 round  3 size 2941774 rdtsc: 152348180
trimedges id 0 round  7 size 825426 rdtsc: 28922872
trimrename id 0 round 14 size 253624 rdtsc: 12262032 maxnnid 4006
trimrename id 0 round 15 size 224569 rdtsc: 10361964 maxnnid 3559
trimedges1 id 0 round 31 size 59912 rdtsc: 657168
trimedges1 id 0 round 63 size 15863 rdtsc: 257240
trimrename1 id 0 round 66 size 14526 rdtsc: 609200 maxnnid 248
trimrename1 id 0 round 67 size 14123 rdtsc: 552332 maxnnid 240
  20-cycle found
Nodes (1d24412,1eed093): 0x968240 (4ede8a,18d901): 0xb6b419 (91de84,18d901): 0x7aaaf (91de84,8da58d): 0xcb572a (1d6719c,8da58d): 0xad8119 (1d6719c,f8d67f): 0xf2898f (dec466,f8d67f): 0xa28796 (dec466,82b97d): 0xbd2765 (d2ba5a,82b97d): 0x6d31e (d2ba5a,19f63d): 0x7fc4fb (1d24412,19f63d): 0xeea5a5 (4ede8a,7861a1): 0x18cdb9 (e75f5a,7861a1): 0x1ffaef (e75f5a,16c579d): 0x72b0e (a92c4c,16c579d): 0x28b919 (a92c4c,1a95dbd): 0xfa27c0 (1a61184,1a95dbd): 0x43d8fa (1a61184,8fa649): 0x134522 (1a4e7f4,8fa649): 0xbbddd6 (1a4e7f4,1eed093): 0xe090d9
  24-cycle found
 148-cycle found
 142-cycle found
  14-cycle found
findcycles rdtsc: 5786836
Time: 756 ms
Solution 0x6d31e, 0x72b0e, 0x7aaaf, 0x134522, 0x18cdb9, 0x1ffaef, 0x28b919, 0x43d8fa, 0x7fc4fb, 0x968240, 0xa28796, 0xad8119, 0xb6b419, 0xbbddd6, 0xbd2765, 0xcb572a, 0xe090d9, 0xeea5a5, 0xf2898f, 0xfa27c0,
Verified with cyclehash f01baf5b437888fefb579aab9d492ce20fb59efcbf01561320ef102a62538b1c
1 total solutions
*/
func TestCuckoo(t *testing.T) {
	n := numcpu.NumCPU()
	p := runtime.GOMAXPROCS(n)
	var k0 uint64 = 0xf4956dc403730b01
	var k1 uint64 = 0xe6d45de39c2a5a3e
	no := []uint32{
		0x6d31e, 0x72b0e, 0x7aaaf, 0x134522, 0x18cdb9,
		0x1ffaef, 0x28b919, 0x43d8fa, 0x7fc4fb, 0x968240,
		0xa28796, 0xad8119, 0xb6b419, 0xbbddd6, 0xbd2765,
		0xcb572a, 0xe090d9, 0xeea5a5, 0xf2898f, 0xfa27c0,
	}

	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, k0)
	binary.LittleEndian.PutUint64(b[8:], k1)
	c := NewCuckoo()
	ans, found := c.PoW(b)
	if !found {
		t.Fatalf("should be found")
	}
	for i, v := range ans {
		if v != no[i] {
			t.Error("nonce is incorrect")
		}
	}
	if err := Verify(b, ans); err != nil {
		t.Error("should be legit, but", err)
	}
	runtime.GOMAXPROCS(p)
}

func BenchmarkCuckoo2(b *testing.B) {
	n := numcpu.NumCPU()
	p := runtime.GOMAXPROCS(n)
	r := make([]byte, 16)
	c := NewCuckoo()

	for i := 0; i < b.N; i++ {
		if _, err := rand.Read(r); err != nil {
			b.Fatal(err)
		}
		c.PoW(r)
	}
	runtime.GOMAXPROCS(p)
}

func BenchmarkCuckooProb(b *testing.B) {
	n := numcpu.NumCPU()
	p := runtime.GOMAXPROCS(n)
	r := make([]byte, 16)
	c := NewCuckoo()

	for i := 0; i < 1000000; i++ {
		if _, err := rand.Read(r); err != nil {
			b.Fatal(err)
		}
		_, ok := c.PoW(r)
		if ok {
			fmt.Println(i)
		}
	}
	runtime.GOMAXPROCS(p)
}
