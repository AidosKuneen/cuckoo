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
	"errors"
)

const (
	edgebits = 25
	//ProofSize is the number of nonces and cycles
	ProofSize = 20
	nedge     = 1 << edgebits
	edgemask  = nedge - 1
	nnode     = 2 * nedge
	easiness  = nnode //* 70 / 100
	maxpath   = 8192
	minnonce  = nnode * 60 / 100
)

//Verify verifiex cockoo nonces.
func Verify(nonces [ProofSize]uint32, sipkey []byte) error {
	sip := newsip(sipkey)
	var uvs [2 * ProofSize]uint32
	var xor0, xor1 uint32

	if nonces[ProofSize-1] > easiness {
		return errors.New("nonce is too big")
	}
	if nonces[ProofSize-1] < minnonce {
		return errors.New("last nonce is too small")
	}

	for n := 0; n < ProofSize; n++ {
		if n > 0 && nonces[n] <= nonces[n-1] {
			return errors.New("nonces are not in order")
		}
		u00 := siphashPRF(&sip.v, uint64(nonces[n]<<1))
		v00 := siphashPRF(&sip.v, (uint64(nonces[n])<<1)|1)
		u0 := uint32(u00&edgemask) << 1
		xor0 ^= u0
		uvs[2*n] = u0
		v0 := (uint32(v00&edgemask) << 1) | 1
		xor1 ^= v0
		uvs[2*n+1] = v0
	}
	if xor0 != 0 {
		return errors.New("U endpoinsts don't match")
	}
	if xor1 != 0 {
		return errors.New("V endpoinsts don't match")
	}

	n := 0
	for i := 0; ; n++ {
		another := i
		for k := (i + 2) % (2 * ProofSize); k != i; k = (k + 2) % (2 * ProofSize) {
			if uvs[k] == uvs[i] {
				if another != i {
					return errors.New("there are branches in nonce")
				}
				another = k
			}
		}
		if another == i {
			return errors.New("dead end in nonce")
		}
		i = another ^ 1
		if i == 0 {
			break
		}
	}
	if n != ProofSize {
		return errors.New("cycle is too short")
	}
	return nil
}
