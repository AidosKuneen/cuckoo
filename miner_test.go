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
Looking for 20-cycle on cuckoo26("") with 70% edges
k0 a34c6a2bdaa03a14 k1 d736650ae53eee9e
  20-cycle found at 67%
Solution 187240 1fa278 28c722 2ae3d9 4bad9f 827253 d2c821 edc876 eef1c1 f40ac3 f5e44a 106a8e3 10a4393 122127f 1484233 161b565 16468d4 17a5b54 17bb3db 1e3034b
  20-cycle found at 94%
Solution 8a955 18d131 472402 576611 81c8b8 bdbc08 c619c6 db35b1 e2de6a 12b177c 132b902 138e3d8 14cb161 17a5eb3 18eb464 1b75a82 215eb16 23813d4 2849c1a 2a426d3
  20-cycle found at 95%
Solution 31543 c35e6 432718 7039d5 800a00 927a3b d70f7f e0fd27 f6e467 1118a46 1507312 16dab68 1a59f01 1b5f466 1c862d1 1dba125 1efb59e 1fbee14 2478e1e 2a9ef30
  20-cycle found at 98%
Solution 12a24f 1aa7f9 3c913e 4b9148 6df923 b95571 cde93e cfc268 109a02b 1328206 139fa99 1567473 18c9b5d 1b3507d 1e6243a 1e6d6b9 1fd56c4 2131377 2674d09 2c2da16
real	8m32.809s
user	8m31.801s
sys	0m0.353s

*/
func TestCuckoo(t *testing.T) {
	var k0 uint64 = 0xa34c6a2bdaa03a14
	var k1 uint64 = 0xd736650ae53eee9e
	no := []uint32{
		0x8a955, 0x18d131, 0x472402, 0x576611, 0x81c8b8, 0xbdbc08, 0xc619c6,
		0xdb35b1, 0xe2de6a, 0x12b177c, 0x132b902, 0x138e3d8, 0x14cb161,
		0x17a5eb3, 0x18eb464, 0x1b75a82, 0x215eb16, 0x23813d4, 0x2849c1a, 0x2a426d3,
	}

	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, k0)
	binary.LittleEndian.PutUint64(b[8:], k1)
	cu := newCuckoo(b)
	ans, found := cu.worker()
	if !found {
		t.Fatalf("should be found")
	}
	for i, v := range ans {
		if v != no[i] {
			t.Error("nonce is incorrect")
		}
	}
}
func TestCuckoo2(t *testing.T) {
	var k0 uint64 = 0xa34c632bdaa03a14
	var k1 uint64 = 0xd736a50ae53eee9e
	no := []uint32{
		0x8a955, 0x18d131, 0x472402, 0x576611, 0x81c8b8, 0xbdbc08, 0xc619c6,
		0xdb35b1, 0xe2de6a, 0x12b177c, 0x132b902, 0x138e3d8, 0x14cb161,
		0x17a5eb3, 0x18eb464, 0x1b75a82, 0x215eb16, 0x23813d4, 0x2849c1a, 0x2a426d3,
	}

	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, k0)
	binary.LittleEndian.PutUint64(b[8:], k1)
	cu := newCuckoo(b)
	ans, found := cu.worker()
	if !found {
		t.Fatalf("should be found")
	}
	for i, v := range ans {
		if v != no[i] {
			t.Error("nonce is incorrect")
		}
	}
}
