//+build amd64 !noasm !appengine
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

#include "textflag.h"

/*
X0=v0
X1=v1
X2=v2
X3=v3
X7:rotate16
*/

DATA rotate16<>+0x00(SB)/8, $0x0504030201000706
DATA rotate16<>+0x08(SB)/8, $0x0D0C0B0A09080F0E
GLOBL rotate16<>(SB), (NOPTR+RODATA), $16

DATA ff<>+0x00(SB)/8, $0x00000000000000FF
DATA ff<>+0x08(SB)/8, $0x00000000000000FF
GLOBL ff<>(SB), (NOPTR+RODATA), $16

DATA init0<>+0x00(SB)/8, $0x736f6d6570736575
DATA init0<>+0x08(SB)/8, $0x646f72616e646f6d
DATA init1<>+0x00(SB)/8, $0x6c7967656e657261
DATA init1<>+0x08(SB)/8, $0x7465646279746573
GLOBL init0<>(SB), (NOPTR+RODATA), $16
GLOBL init1<>(SB), (NOPTR+RODATA), $16


#define ADD(a,b)\
	PADDQ	b, a

#define XOR(a, b)\
	PXOR	b, a

#define ROT(x,n) \
	MOVOA	x, X4 \
	PSLLQ	$n, X4 \ 
	PSRLQ	$(64-n), x \
	POR	X4, x 

#define ROT16(x) \
	PSHUFB	X7, x

#define ROT32(x) \
	PSHUFD	$0xB1, x, x          //X5 = X5[1,0,3,2]


#define SIPROUND \
    ADD(X0, X1) \
	ADD(X2, X3) \
	ROT(X1, 13) \
    ROT16(X3)\
	XOR(X1, X0)\
	XOR(X3, X2) \
    ROT32(X0) \
	ADD(X2, X1) \ 
	ADD(X0, X3) \
    ROT(X1, 17) \
	ROT(X3, 21) \
    XOR(X1, X2) \
	XOR(X3, X0) \
	ROT32(X2)

// func siphash(k0, k1,b0,b1 uint64) (uint64,uint64)
TEXT ·siphash(SB), NOSPLIT, $0
	MOVOU	k0+0(FP), X0
	MOVOA	X0, X2
	PXOR	init0<>(SB), X0
	PXOR	init1<>(SB), X2
	MOVOA	X0, X1
	MOVHLPS	X1, X1              
	MOVLHPS	X0, X0
	MOVOA	X2, X3
	MOVHLPS	X3, X3              
	MOVLHPS	X2, X2
	MOVOA	rotate16<>(SB),X7


    MOVOU   b0+16(FP), X6
	XOR(X3, X6)
	SIPROUND
	SIPROUND
	XOR(X0, X6)
	PXOR	ff<>(SB), X2
	SIPROUND
	SIPROUND
	SIPROUND
	SIPROUND
	XOR(X0, X1)
	XOR(X2, X3)
	XOR(X0, X2)
	MOVOU	X0, ret+32(FP)
	RET



// func siphashPRF(v0, v1,v2,v3,b0,b1 uint64) (uint64,uint64)
TEXT ·siphashPRF(SB), NOSPLIT, $0
	MOVLPD	v0+0(FP), X0
	MOVLPD	v1+8(FP), X1
	MOVLPD	v2+16(FP), X2
	MOVLPD	v3+24(FP), X3
	MOVLHPS	X0, X0
	MOVLHPS	X1, X1
	MOVLHPS	X2, X2
	MOVLHPS	X3, X3
	MOVOA	rotate16<>(SB),X7

    MOVOU   b0+32(FP), X6
	XOR(X3, X6)
	SIPROUND
	SIPROUND
	XOR(X0, X6)
	PXOR	ff<>(SB), X2
	SIPROUND
	SIPROUND
	SIPROUND
	SIPROUND
	XOR(X0, X1)
	XOR(X2, X3)
	XOR(X0, X2)
	MOVOU	X0, ret+48(FP)
	RET

