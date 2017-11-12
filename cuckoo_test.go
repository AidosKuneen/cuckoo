// Copyright (c) 2017 Aidos Developer

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
	"encoding/binary"
	"math/rand"
	"testing"
)

/*
Looking for 20-cycle on cuckoo25("") with 100% edges
33554433 4
k0 a34c6a2bdaa03a14 k1 d736650ae53eee9e
  20-cycle found at 72% 24161625
Solution,0xbec0d,0x19d70c,0x20d56b,0x2ae3eb,0x2d6ded,0x3f3535,0x5af313,0x65d7af,0x67e5d4,0x79d5c5,0x7e950e,0x801ce6,0x844274,0xb4ee35,0xcac488,0xe18224,0xe54eeb,0xe9205b,0xf1bd7c,0x170ad59

real	1m24.845s
user	1m24.365s
sys	0m0.051s
*/
func TestCuckoo(t *testing.T) {
	var k0 uint64 = 0xa34c6a2bdaa03a14
	var k1 uint64 = 0xd736650ae53eee9e
	no := []uint32{
		0xbec0d, 0x19d70c, 0x20d56b, 0x2ae3eb, 0x2d6ded, 0x3f3535, 0x5af313,
		0x65d7af, 0x67e5d4, 0x79d5c5, 0x7e950e, 0x801ce6, 0x844274, 0xb4ee35,
		0xcac488, 0xe18224, 0xe54eeb, 0xe9205b, 0xf1bd7c, 0x170ad59,
	}

	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, k0)
	binary.LittleEndian.PutUint64(b[8:], k1)
	ans, found := PoW(b, func(nonces *[ProofSize]uint32) bool {
		return true
	})
	if !found {
		t.Fatalf("should be found")
	}
	for i, v := range ans {
		if v != no[i] {
			t.Error("nonce is incorrect")
		}
	}
}

func TestVerify(t *testing.T) {
	var k0 uint64 = 0xa34c6a2bdaa03a14
	var k1 uint64 = 0xd736650ae53eee9e
	no := [ProofSize]uint32{
		0xbec0d, 0x19d70c, 0x20d56b, 0x2ae3eb, 0x2d6ded, 0x3f3535, 0x5af313,
		0x65d7af, 0x67e5d4, 0x79d5c5, 0x7e950e, 0x801ce6, 0x844274, 0xb4ee35,
		0xcac488, 0xe18224, 0xe54eeb, 0xe9205b, 0xf1bd7c, 0x170ad59,
	}
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, k0)
	binary.LittleEndian.PutUint64(b[8:], k1)
	if err := Verify(b, &no); err != nil {
		t.Error("should be legit, but", err)
	}
}

func TestCuckoo2(t *testing.T) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	ans, found := PoW(b, func(nonces *[ProofSize]uint32) bool {
		return true
	})
	if !found {
		t.Fatalf("should be found")
	}
	if err := Verify(b, ans); err != nil {
		t.Fatal("nonces are not correct", err)
	}
}
