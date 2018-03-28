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

import "encoding/binary"

type sip struct {
	k0 uint64
	k1 uint64
	v  [4]uint64
}

func newsip(h []byte) *sip {
	s := &sip{
		k0: binary.LittleEndian.Uint64(h[:]),
		k1: binary.LittleEndian.Uint64(h[8:]),
	}
	s.v[0] = s.k0 ^ 0x736f6d6570736575
	s.v[1] = s.k1 ^ 0x646f72616e646f6d
	s.v[2] = s.k0 ^ 0x6c7967656e657261
	s.v[3] = s.k1 ^ 0x7465646279746573
	return s
}
