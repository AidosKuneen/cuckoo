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
	"testing"
)

/*
nonce 18 k0 k1 6085f431d6490443 79e6ecd104d543f8
genUnodes round  0 size 134217728 rdtsc: 5217694472
genVnodes round  1 size 84839301 rdtsc: 5715265860
trimedges round  3 size 23533857 rdtsc: 1331876588
trimedges round  7 size 6625317 rdtsc: 244714824
trimrename round 10 size 3665171 rdtsc: 180670088 maxnnid 28436
trimrename round 11 size 3113682 rdtsc: 150380252 maxnnid 24308
trimedges1 round 15 size 1809392 rdtsc: 22943644
trimedges1 round 31 size 482777 rdtsc: 6897076
trimedges1 round 63 size 127102 rdtsc: 2382744
trimrename1 round 66 size 116195 rdtsc: 5551112 maxnnid 1048
trimrename1 round 67 size 112833 rdtsc: 6683168 maxnnid 980
   2-cycle found
   6-cycle found
   2-cycle found
   6-cycle found
  24-cycle found
  88-cycle found
  42-cycle found
  Solution 31f5df 77fe19 c239a6 e58c4d 12c9dee 1300a26 144a355
   1453457 17a1aeb 19242d4 1c25cce 21cc945 23cd7a7 25dbd1c 276b507
   2fa0c1f 3027d5e 30404a4 3763527 3e86a15 442daaf 481b337 4ff3a50
	5202c32 55b6400 5ca8446 5e80656 5efe6b8 5f25886 6454270 667bdf9
	69a5716 76f180b 779fbf7 793fb75 7b077d3 7c82af5 7d10242 7d50826
	7d6e66b 7d83329 7ee1dc7
findcycles rdtsc: 34112464
Time: 6462 ms
*/

func TestCuckoo(t *testing.T) {
	var k0 uint64 = 0x6085f431d6490443
	var k1 uint64 = 0x79e6ecd104d543f8
	_ = []uint32{
		0x31f5df, 0x77fe19, 0xc239a6, 0xe58c4d, 0x12c9dee, 0x1300a26, 0x144a355,
		0x1453457, 0x17a1aeb, 0x19242d4, 0x1c25cce, 0x21cc945, 0x23cd7a7, 0x25dbd1c, 0x276b507,
		0x2fa0c1f, 0x3027d5e, 0x30404a4, 0x3763527, 0x3e86a15, 0x442daaf, 0x481b337, 0x4ff3a50,
		0x5202c32, 0x55b6400, 0x5ca8446, 0x5e80656, 0x5efe6b8, 0x5f25886, 0x6454270, 0x667bdf9,
		0x69a5716, 0x76f180b, 0x779fbf7, 0x793fb75, 0x7b077d3, 0x7c82af5, 0x7d10242, 0x7d50826,
		0x7d6e66b, 0x7d83329, 0x7ee1dc7,
	}

	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, k0)
	binary.LittleEndian.PutUint64(b[8:], k1)
	cu := newCuckoo(b)
	cu.worker()
	// if !found {
	// 	t.Error("should found")
	// }
	// if len(nonces) != len(no) {
	// 	t.Error("invalid number of nonces")
	// }
	// for i, v := range nonces {
	// 	if v != no[i] {
	// 		t.Error("nonce is incorrect")
	// 	}
	// }
}
