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

// int main(int argc, char** argv)
// {
//     int i = 0;
//     uint64_t k0 = 0x0011223344556677;
//     uint64_t k1 = 0x8899aabbccddeeff;
//     uint64_t v[4];
//     v[0] = k0 ^ 0x736f6d6570736575;
//     v[1] = k1 ^ 0x646f72616e646f6d;
//     v[2] = k0 ^ 0x6c7967656e657261;
//     v[3] = k1 ^ 0x7465646279746573;
//     uint64_t uorv = 1;
//     uint64_t ts[16];
//     uint64_t nonce[] = {
//         5577006791947779410u,
//         8674665223082153551u,
//         6129484611666145821u,
//         4037200794235010051u,
//         3916589616287113937u,
//         6334824724549167320u,
//         605394647632969758u,
//         1443635317331776148u,
//         894385949183117216u,
//         2775422040480279449u,
//         4751997750760398084u,
//         7504504064263669287u,
//         1976235410884491574u,
//         3510942875414458836u,
//         2933568871211445515u,
//         4324745483838182873u
//     };
//     uint64_t answers[] = {
//         15792995720401100103u,
//         9170369451364337636u,
//         824572431404338964u,
//         15761619652798949465u,
//         12996180346644554812u,
//         725343453931019011u,
//         17158853600731801478u,
//         17546331119738431785u,
//         12830094850130523973u,
//         13939840424219265429u,
//         6322588667715040360u,
//         5038487420505719201u,
//         1936843290018707214u,
//         14911679484758948273u,
//         16011335585547651053u,
//         15805936675293436437u
//     };
//     siphashPRF16(v, nonce, uorv, ts);
//     for (i = 0; i < 16; i++) {
//         printf("%d  %" PRIu64 " %" PRIu64 "\n", i, ts[i], answers[i]);
//         if (ts[i] != answers[i]) {
//             printf("incorrect\n");
//         }
//     }
//     uint64_t answers2[] = {
//         16902871805249102149u,
//         509710775399566316u,
//         17460957912471295606u,
//         8004697006529892058u,
//         1484206524365537172u,
//         3472809066092780505u,
//         1286860054142601603u,
//         9560556364460450277u,
//         8215212750239263445u,
//         5178032223764063101u,
//         11724744714162270015u,
//         10714585442073777882u,
//         16376165729377320285u,
//         1037271618166486051u,
//         17073545850867311974u,
//         8555370355296464024u
//     };
//     siphashPRF16Seq(v, 2610529275472644968u, uorv, ts);
//     for (i = 0; i < 16; i++) {
//         printf("%d  %" PRIu64 " %" PRIu64 "\n", i, ts[i], answers2[i]);
//         if (ts[i] != answers2[i]) {
//             printf("incorrect\n");
//         }
//     }

//     return 0;
// }
*/
import "C"
import "unsafe"

func siphashPRF8192(v *[4]uint64, nonce *[8192]uint64, uorv uint64, result *[8192]uint64) {
	C.siphashPRF8192((*C.ulonglong)(unsafe.Pointer(&v[0])), (*C.ulonglong)(unsafe.Pointer(&nonce[0])), C.ulonglong(uorv), (*C.ulonglong)(unsafe.Pointer(&result[0])))
}
func siphashPRF8192Seq(v *[4]uint64, nonce uint64, uorv uint64, result *[8192]uint64) {
	C.siphashPRF8192Seq((*C.ulonglong)(unsafe.Pointer(&v[0])), C.ulonglong(nonce), C.ulonglong(uorv), (*C.ulonglong)(unsafe.Pointer(&result[0])))
}
