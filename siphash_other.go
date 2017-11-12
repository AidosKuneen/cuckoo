//+build !amd64
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

func siphashPRF16(v0, v1, v2, v3 uint64, nonce *[16]uint64, uorv uint64, result *[16]uint64) {
	for i := range nonce {
		b := (nonce[i] << 1) | uorv
		result[i] = siphashPRFGeneral(v0, v1, v2, v3, b)
	}
}
func siphashPRF16Seq(v0, v1, v2, v3 uint64, nonce uint64, uorv uint64, result *[16]uint64) {
	for i := uint64(0); i < 16; i++ {
		b := ((nonce + i) << 1) | uorv
		result[i] = siphashPRFGeneral(v0, v1, v2, v3, b)
	}
}
