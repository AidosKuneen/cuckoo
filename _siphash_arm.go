// Copyright (c) 2019 Aidos Developer

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

// #cgo CFLAGS: -mfloat-abi=softfp -mfpu=neon -Wall
// #cgo LDFLAGS:  -mfpu=neon
/*
#include <arm_neon.h>
#include <stdio.h>
#include <inttypes.h>

#define rotate(x, n) vorrq_u64(vshlq_n_u64(x, n), vshrq_n_u64(x, 64 - n))

static inline void sip_round(uint64x2_t x[4])
{
    uint32x4_t tmp;

    x[0] = vaddq_u64(x[0], x[1]);
    x[2] = vaddq_u64(x[2], x[3]);
    x[1] = rotate(x[1], 13);
    x[3] = rotate(x[3], 16);
    x[1] = veorq_u64(x[1], x[0]);
    x[3] = veorq_u64(x[3], x[2]);

    tmp = vreinterpretq_u32_u64(x[0]);
    tmp = vrev64q_u32(tmp);
    x[0] = vreinterpretq_u64_u32(tmp);

    x[2] = vaddq_u64(x[2], x[1]);
    x[0] = vaddq_u64(x[0], x[3]);
    x[1] = rotate(x[1], 17);
    x[3] = rotate(x[3], 21);
    x[1] = veorq_u64(x[1], x[2]);
    x[3] = veorq_u64(x[3], x[0]);

    tmp = vreinterpretq_u32_u64(x[2]);
    tmp = vrev64q_u32(tmp);
    x[2] = vreinterpretq_u64_u32(tmp);
}

static inline uint64x2_t sip_rounds(uint64x2_t x[4], uint64x2_t nonce)
{
    const uint64x2_t ff = vdupq_n_u64(0xff);

    x[3] = veorq_u64(x[3], nonce);
    sip_round(x);
    sip_round(x);
    x[0] = veorq_u64(x[0], nonce);
    x[2] = veorq_u64(x[2], ff);
    sip_round(x);
    sip_round(x);
    sip_round(x);
    sip_round(x);
    x[0] = veorq_u64(x[0], x[1]);
    x[2] = veorq_u64(x[2], x[3]);
    x[0] = veorq_u64(x[0], x[2]);
    return x[0];
}

void siphashPRF8192(uint64_t v[4], uint64_t nonce[8192], uint64_t uorv, uint64_t result[8192])
{
    uint64x2_t vs[4], vs_2[4];
    uint64_t i = 0, j = 0;
    for (i = 0; i < 4; i++) {
        vs_2[i] = vdupq_n_u64(v[i]);
    }
    uint64x2_t uv = vdupq_n_u64(uorv);
    for (i = 0; i < 8192; i += 2) {
        for (j = 0; j < 4; j++) {
            vs[j] = vs_2[j];
        }
        uint64_t ntmp[] = { nonce[i], nonce[i + 1] };
        uint64x2_t n = vld1q_u64(ntmp);
        n = vshlq_n_u64(n, 1);
        n = vorrq_u64(n, uv);
        uint64x2_t r = sip_rounds(vs, n);
        vst1q_u64(&result[i], r);
    }
}

void siphashPRF8192Seq(uint64_t v[4], uint64_t nonce, uint64_t uorv, uint64_t result[8192])
{
    uint64x2_t vs[4], vs_2[4];
    uint64_t i = 0, j = 0;
    for (i = 0; i < 4; i++) {
        vs_2[i] = vdupq_n_u64(v[i]);
    }
    uint64x2_t uv = vdupq_n_u64(uorv);
    for (i = 0; i < 8192; i += 2) {
        for (j = 0; j < 4; j++) {
            vs[j] = vs_2[j];
        }
        uint64_t ntmp[] = { nonce + i, nonce + i + 1 };
        uint64x2_t n = vld1q_u64(ntmp);
        n = vshlq_n_u64(n, 1);
        n = vorrq_u64(n, uv);
        uint64x2_t r = sip_rounds(vs, n);
        vst1q_u64(&result[i], r);
    }
}
*/
import "C"
import "unsafe"

func siphashPRF8192(v *[4]uint64, nonce *[8192]uint64, uorv uint64, result *[8192]uint64) {
	C.siphashPRF8192((*C.ulonglong)(unsafe.Pointer(&v[0])), (*C.ulonglong)(unsafe.Pointer(&nonce[0])), C.ulonglong(uorv), (*C.ulonglong)(unsafe.Pointer(&result[0])))
}
func siphashPRF8192Seq(v *[4]uint64, nonce uint64, uorv uint64, result *[8192]uint64) {
	C.siphashPRF8192Seq((*C.ulonglong)(unsafe.Pointer(&v[0])), C.ulonglong(nonce), C.ulonglong(uorv), (*C.ulonglong)(unsafe.Pointer(&result[0])))
}
