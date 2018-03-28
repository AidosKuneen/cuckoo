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
X4:tmp
X5:ff
X6:nonce
X7:rotate16
X8:X0
X9:X1
X10:X2
X11:X3
X12:uorv
*/

DATA rotate16<>+0x00(SB)/8, $0x0504030201000706
DATA rotate16<>+0x08(SB)/8, $0x0D0C0B0A09080F0E
GLOBL rotate16<>(SB), (NOPTR+RODATA), $16

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


#define SIP1 \
	XOR(X3, X6) \
	SIPROUND  \
	SIPROUND  \
	XOR(X0, X6) \
	PXOR	X5, X2 \
	SIPROUND \
	SIPROUND \
	SIPROUND \
	SIPROUND \
	XOR(X0, X1) \
	XOR(X2, X3) \
	XOR(X0, X2)

#define NSIP \
	MOVOU	(CX),X6 \
	PSLLQ	$1, X6 \
	POR		X12,X6 \
	ADDQ	$16, CX \
	MOVOA	X8,X0 \ 
	MOVOA	X9,X1 \ 
	MOVOA	X10,X2 \
	MOVOA	X11,X3 \
	SIP1 \
	MOVOU X0, (BX) \
	ADDQ $16, BX

//func siphashPRF16(v *[4]uint64, nonce *[16]uint64, uorv uint64, result *[16]uint64)
TEXT ·siphashPRF16(SB), NOSPLIT, $0
	MOVQ	$0xff, AX
	MOVQ	AX, X5
	MOVLHPS	X5, X5
	MOVOA	rotate16<>(SB),X7

	MOVQ	v+0(FP), CX
	MOVQ	(CX), X8
	ADDQ	$8, CX
	MOVQ	(CX), X9
	ADDQ	$8, CX
	MOVQ	(CX), X10
	ADDQ	$8, CX
	MOVQ	(CX), X11
    MOVQ	nonce+8(FP), CX
	MOVQ	uorv+16(FP), X12
	MOVQ	result+24(FP), BX
	MOVLHPS	X8, X8
	MOVLHPS	X9, X9 
	MOVLHPS	X10, X10 
	MOVLHPS	X11, X11
	MOVLHPS	X12, X12

	NSIP //0
	NSIP //1
	NSIP //2
	NSIP //3
	NSIP //4
	NSIP //5
	NSIP //6
	NSIP //7
	RET

#define NSIP_SEQ \
	MOVOA	X8,X0 \ 
	MOVOA	X9,X1 \ 
	MOVOA	X10,X2 \
	MOVOA	X11,X3 \
	SIP1 \
	MOVOU X0, (BX) \
	ADDQ $16, BX \
	MOVQ	$4, AX \
	MOVQ	AX, X4 \
	MOVLHPS	X4, X4 \
	PADDQ	X4, X6 \


//func siphashPRF16Seq(v *[4]uint64, nonce uint64, uorv uint64, result *[16]uint64)
TEXT ·siphashPRF16Seq(SB), NOSPLIT, $0
	MOVQ	$0xff, AX
	MOVQ	AX, X5
	MOVLHPS	X5, X5
	MOVOA	rotate16<>(SB),X7

	MOVQ	v+0(FP), CX
	MOVQ	(CX), X8
	ADDQ	$8, CX
	MOVQ	(CX), X9
	ADDQ	$8, CX
	MOVQ	(CX), X10
	ADDQ	$8, CX
	MOVQ	(CX), X11
    MOVQ	nonce+8(FP), X6
	MOVQ	uorv+16(FP), X12
	MOVQ	result+24(FP), BX
	MOVLHPS	X8, X8
	MOVLHPS	X9, X9 
	MOVLHPS	X10, X10 
	MOVLHPS	X11, X11
	MOVLHPS	X12, X12

	MOVLHPS	X6, X6
	MOVQ	$1, AX
	MOVQ	AX, X4
	PSHUFD	$0x4e, X4, X4 //0x01001110   
	PADDQ	X4, X6
	PSLLQ	$1, X6
	POR		X12, X6

	NSIP_SEQ //0
	NSIP_SEQ //1
	NSIP_SEQ //2
	NSIP_SEQ //3
	NSIP_SEQ //4
	NSIP_SEQ //5
	NSIP_SEQ //6
	NSIP_SEQ //7
	RET

