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
	"log"
)

const (
	maxPath     = 8192
	nround      = 20
	edgebits    = 27
	counterbits = 27
	proofSize   = 42
	nedge       = 1 << edgebits
	ncounter    = 1 << counterbits
	edgemask    = nedge - 1
	countermask = 1<<counterbits - 1
	nnode       = 2 * nedge
	easiness    = nnode * 0.5
	deadbit     = 0x80000000
)

type counter [(ncounter * 2) >> 3]byte

func (c *counter) clear() {
	for i := range c {
		c[i] = 0
	}
}
func (c *counter) incr(v uint) {
	lbyte := (v * 2) >> 3
	lbit0 := ((v * 2) & 7)
	lbit1 := (((v * 2) & 7) + 1)

	if (c[lbyte]>>lbit1)&1 == 1 {
		return
	}
	if (c[lbyte]>>lbit0)&1 == 1 {
		c[lbyte] |= 1 << lbit1 //bit1=1
		return
	}
	c[lbyte] |= 1 << lbit0 //bit0=1
}
func (c *counter) isLeaf(v uint) bool {
	lbyte := (v * 2) >> 3
	lbit1 := (((v * 2) & 7) + 1)

	bit1 := (c[lbyte] >> lbit1) & 1

	if bit1 == 0 {
		return true
	}
	return false
}

type sip struct {
	k0 uint64
	k1 uint64
	v0 uint64
	v1 uint64
	v2 uint64
	v3 uint64
}

func newsip(k []byte) *sip {
	s := &sip{
		k0: binary.LittleEndian.Uint64(k[:8]),
		k1: binary.LittleEndian.Uint64(k[8:]),
	}
	s.v0 = s.k0 ^ 0x736f6d6570736575
	s.v1 = s.k1 ^ 0x646f72616e646f6d
	s.v2 = s.k0 ^ 0x6c7967656e657261
	s.v3 = s.k1 ^ 0x7465646279746573
	return s
}

type set map[uint64]struct{}

func (s set) add(u, v uint32) {
	s[uint64(u)<<16|uint64(v)] = struct{}{}
}
func (s set) exist(u, v uint32) bool {
	_, exist := s[uint64(u)<<32|uint64(v)]
	return exist
}

type edge uint32

type cuckoo struct {
	matrix [easiness]edge
	sip    *sip
}

func newCuckoo(n []byte) *cuckoo {
	return &cuckoo{
		sip: newsip(n),
	}
}
func (c *cuckoo) trimU() {
	var bits counter
	cnt := 0
	for i := uint(0); i < easiness; i++ {
		if c.matrix[i]&deadbit != 0 || c.matrix[i]&1 == 1 {
			continue
		}
		v := uint(c.matrix[i]>>1) & countermask
		bits.incr(v)
	}
	for i := uint(0); i < easiness; i++ {
		if c.matrix[i]&deadbit != 0 || c.matrix[i]&1 == 1 {
			continue
		}
		v := uint(c.matrix[i]>>1) & countermask
		if bits.isLeaf(v) {
			c.matrix[i] |= deadbit
		} else {
			cnt++
		}
	}
	log.Println("unodes:", cnt, "/", easiness)
}
func (c *cuckoo) trimV() {
	var bits counter
	idxu := 0
	cnt := 0
	for i := uint(0); i < easiness; i++ {
		if c.matrix[i]&1 == 0 {
			continue
		}
		for ; c.matrix[idxu]&1 == 1; idxu++ {
			log.Println(idxu)
		}
		if c.matrix[idxu]&deadbit != 0 {
			idxu++
			continue
		}
		shiftv := uint(c.matrix[i]>>1) & countermask
		bits.incr(shiftv)
		idxu++
	}
	idxu = 0
	for i := uint(0); i < easiness; i++ {
		if c.matrix[i]&1 == 0 {
			continue
		}
		for ; c.matrix[idxu]&1 == 1; idxu++ {
		}
		if c.matrix[idxu]&deadbit != 0 {
			idxu++
			continue
		}
		shiftv := uint(c.matrix[i]>>1) & countermask
		if bits.isLeaf(shiftv) {
			c.matrix[i] |= deadbit
		} else {
			cnt++
		}
		idxu++
	}
	log.Println("vnodes:", cnt, "/", easiness)
}
func (c *cuckoo) buildUNodes() {
	var nodes [16]uint64
	var non [16]uint64
	for nonce := 0; nonce < easiness; nonce += 16 {
		for i := range non {
			non[i] = uint64(nonce + i)
		}
		siphashPRF16(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, &non, 0, &nodes)
		for i := range nodes {
			u := (nodes[i] & edgemask) << 1
			c.matrix[nonce+i] = edge(u)
		}
	}
}

func (c *cuckoo) buildVNodes() {
	var nodes [16]uint64
	var non [16]uint64
	cnt := 0
	idxv := 0
	for nonce := 0; nonce < easiness; nonce++ {
		if c.matrix[nonce]&deadbit != 0 {
			continue
		}
		non[cnt] = uint64(nonce)
		cnt++
		if cnt < 16 {
			continue
		}
		cnt = 0
		siphashPRF16(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, &non, 1, &nodes)
		for i := range nodes {
			for ; c.matrix[idxv]&deadbit == 0; idxv++ {
			}
			v := edge((nodes[i]&edgemask)<<1) | 1
			c.matrix[idxv] = v
			idxv++
		}
	}
	if cnt != 0 {
		siphashPRF16(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, &non, 1, &nodes)
		for i := range nodes {
			for ; c.matrix[idxv]&1 == 0; idxv++ {
			}
			v := edge((nodes[i]&edgemask)<<1) | 1
			c.matrix[idxv] = v
			idxv++
		}
	}
}
func (c *cuckoo) worker() ([]uint32, bool) {
	c.buildUNodes()
	// log.Println("built U nodes")
	c.trimU()
	// log.Println("trim U nodes")
	// c.buildVNodes()
	// log.Println("built V nodes")
	// c.trimV()
	// log.Println("trim V nodes")
	return nil, false
}

// func (c *cuckoo) solution(us, vs []uint32) ([]uint32, bool) {
// 	nu := uint16(len(us) - 1)
// 	nv := uint16(len(vs) - 1)
// 	if us[nu] != vs[nv] {
// 		return nil, false
// 	}
// 	min := nu
// 	if min > nv {
// 		min = nv
// 	}
// 	nv -= min
// 	nu -= min
// 	for us[nu] != vs[nv] {
// 		nu++
// 		nv++
// 	}
// 	l := nu + nv + 1
// 	if l != proofSize {
// 		return nil, false
// 	}

// 	cycle := make(set)
// 	cycle.add(us[0], vs[0])
// 	for nu--; nu > 0; nu-- {
// 		cycle.add(us[(nu+1)&(^uint16(1))], us[nu|1])
// 	}
// 	for nv--; nv > 0; nv-- {
// 		cycle.add(vs[(nv|1)], vs[(nv+1)&(^uint16(1))])
// 	}
// 	answer := make([]uint32, 0, proofSize)
// 	var nodesU [16]uint64
// 	var nodesV [16]uint64
// 	var non [16]uint64
// 	for nonce := 0; nonce < easiness; nonce += 16 {
// 		for i := 0; i < 16; i++ {
// 			non[i] = uint64(nonce + i)
// 		}
// 		siphashPRF16(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, &non, 0, &nodesU)
// 		siphashPRF16(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, &non, 1, &nodesV)
// 		for i := 0; i < 16; i++ {
// 			if cycle.exist(uint32(nodesU[i]), uint32(nodesV[i])) {
// 				answer = append(answer, uint32(nonce+i))
// 			}
// 		}
// 	}
// 	return answer, true
// }

// func (c *cuckoo) path(u uint32) ([]uint32, bool) {
// 	us := make([]uint32, proofSize)
// 	for nu := 0; u.index != 0; nu++ {
// 		if nu >= maxPath {
// 			return nil, false
// 		}
// 		us[nu] = u
// 		u = c.cuckoo[u.index]
// 	}
// 	return us, true
// }

// func (c *cuckoo) findCycle() ([]uint32, bool) {
// 	for ux := uint8(0); ux < nx; ux++ {
// 		for vx := uint8(0); vx < nx; vx++ {
// 			for i, ce := range c.matrix[ux][vx] {
// 				if !ce.alive() {
// 					continue
// 				}
// 				u0 := edgemap{
// 					index: xz16(ux, ce.yz[0].z) << 1,
// 					orig:  c.original[ux][vx][i].catXYZ(0, ux),
// 				}
// 				v0 := edgemap{
// 					index: xz16(vx, ce.yz[1].z)<<1 | 1,
// 					orig:  c.original[ux][vx][i].catXYZ(1, vx),
// 				}
// 				us, oku := c.path(u0)
// 				vs, okv := c.path(v0)
// 				if !oku || !okv {
// 					continue
// 				}
// 				if us[len(us)-1] == vs[len(vs)-1] {
// 					if ans, ok := c.solution(us, vs); ok {
// 						return ans, true
// 					}
// 					continue
// 				}
// 				if len(us) < len(vs) {
// 					for nu := len(us) - 2; nu > 0; nu-- {
// 						c.cuckoo[us[nu+1].index] = us[nu]
// 					}
// 					c.cuckoo[u0.index] = v0
// 				} else {
// 					for nv := len(vs) - 2; nv > 0; nv-- {
// 						c.cuckoo[vs[nv+1].index] = vs[nv]
// 					}
// 					c.cuckoo[v0.index] = u0
// 				}
// 			}
// 		}
// 	}
// 	return nil, false
// }
