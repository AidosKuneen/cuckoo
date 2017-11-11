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

// Original license from https://github.com/dchest/siphash
// Written in 2012 by Dmitry Chestnykh.
//
// To the extent possible under law, the author have dedicated all copyright
// and related and neighboring rights to this software to the public domain
// worldwide. This software is distributed without any warranty.
// http://creativecommons.org/publicdomain/zero/1.0/

package cuckoo

import "errors"

func verify(nonces [proofSize]uint32, header []byte) error {
	sip := newsip(header)
	var uvs [2 * proofSize]uint32
	var xor0, xor1 uint32
	var ns, resultU, resultV [16]uint64
	for n := 0; n < proofSize; n += 16 {
		for i := 0; i < 16; i++ {
			if nonces[n+i] > edgemask {
				return errors.New("nonce is too big")
			}
			if n+i > 0 && nonces[n+i] <= nonces[n+i-1] {
				return errors.New("nonces are not in order")
			}
			ns[i] = uint64(nonces[n+i])
		}
		siphashPRF16(sip.v0, sip.v1, sip.v2, sip.v3, &ns, 0, &resultU)
		siphashPRF16(sip.v0, sip.v1, sip.v2, sip.v3, &ns, 1, &resultV)
		for i := 0; i < 16; i++ {
			u0:=uint32(resultU[i]) <<1
			xor0 ^= u0
			uvs[2*(n+i)] = u0
			v0:=(uint32(resultU[i]) <<1)|1
			xor1 ^= v0
			uvs[2*(n+i)+1] = v0
		}
	}
	for n:=0;n<proofSize
	if xor0|xor1 != 0 {
		return errors.New("endpoinsts don't match")
	}
	return nil
}
