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
)

const (
	maxPath   = 8192
	nround    = 20
	edgebits  = 27
	proofSize = 42
	nedge     = 1 << edgebits
	edgemask  = nedge - 1
	nnode     = 2 * nedge
	easiness  = nnode * 0.5

	xbits  = 6
	ybits  = 6
	zbits  = edgebits - xbits - ybits
	yzbits = ybits + zbits
	nx     = 1 << xbits
	ny     = 1 << ybits
	nz     = 1 << zbits

	rename0  = 15
	rename1  = 9
	nrename0 = 1 << rename0
	nrename1 = 1 << rename1

	narray = (easiness / (nx * ny)) * 67 / 64
)

func xmask(v int) int {
	return (v & ((nx - 1) << yzbits)) >> yzbits
}
func ymask(v int) int {
	return (v & ((ny - 1) << zbits)) >> zbits
}

//truncate last bits
func zmask(v int) int {
	return (v & (nz - 2)) >> 1
}

//truncate last bits
func yzmask(v int) int {
	return (v & 0x3ffffe) >> 1
}
func xyz(x, y, z, uorv int) int {
	return ((x<<yzbits)|(y<<zbits)|z)<<1 | uorv
}

func yz(y, z, uorv int) int {
	return ((y<<zbits)|z)<<1 | uorv
}

type counter []byte

func (c counter) incr(v uint) {
	bit0 := c[(v*2)>>3] >> ((v * 2) % 8) & 1
	bit1 := c[(v*2)>>3] >> (((v * 2) % 8) + 1) & 1
	if bit1 == 1 {
		return
	}
	if bit0 == 1 {
		// bit0 = 0
		c[(v*2)>>3] |= 1 << (((v * 2) % 8) + 1) //bit1=1
		return
	}
	c[(v*2)>>3] |= 1 << ((v * 2) % 8) //bit0=1
}
func (c counter) isLeaf(v uint) bool {
	bit1 := c[(v*2)>>8] >> (((v * 2) % 8) + 1) & 1
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

func (s *sip) nodes(nonce uint32) (uint32, uint32) {
	n0, n1 := siphashPRF(s.v0, s.v1, s.v2, s.v3,
		uint64(nonce)<<1, uint64(nonce)<<1|1)
	return uint32(n0&edgemask) << 1, uint32(n1&edgemask)<<1 | 1
}

type cell [5]byte

// func newcell(uxyz, vxyz uint32) cell {
// 	uyz := (uint64(uxyz) & 0x1ffffe) << (ybits + zbits - 1)
// 	vyz := (uint64(vxyz) & 0x1ffffe) >> 1
// 	b := make([]byte, 8)
// 	binary.LittleEndian.PutUint64(b, uyz|vyz)
// 	var c cell
// 	copy(c[:], b)

// 	return c
// }

// func (ce cell) y(uorv int) uint8 {
// 	b := make([]byte, 8)
// 	copy(b, ce[:])
// 	v := binary.LittleEndian.Uint64(b)
// 	bits := (ybits + zbits - 1) * (1 - uint(uorv))
// 	return uint8(v>>(bits+zbits-1)) & 0x3f
// }

// //truncate last bits
// func (ce cell) z(uorv int) uint16 {
// 	b := make([]byte, 8)
// 	copy(b, ce[:])
// 	v := binary.LittleEndian.Uint64(b)
// 	bits := (ybits + zbits - 1) * (1 - uint(uorv))
// 	return uint16(v>>bits) & 0x4fff
// }

// //truncate last bits
// func (ce cell) yz(uorv int) uint32 {
// 	b := make([]byte, 8)
// 	copy(b, ce[:])
// 	v := binary.LittleEndian.Uint64(b)
// 	bits := (ybits + zbits - 1) * (1 - uint(uorv))
// 	return uint32(v>>bits) & 0xfffff
// }

type edgemap struct {
	index uint16
	orig  uint32
}

type cuckoo struct {
	size   [nx][ny]int
	cuckoo []edgemap
	sip    *sip
	dead   [nx][ny][narray]bool
	matrix [nx][ny][narray][5]byte
	// original [nx][nx][]cell //never append or would breaks matrix
}

func newCuckoo(k []byte) *cuckoo {
	c := &cuckoo{
		cuckoo: make([]edgemap, 1<<16),
		sip:    newsip(k),
	}
	return c
}

/*
func (c *cuckoo) buildUnodes() {
	var nodes [16]uint64
	for nonce := 0; nonce < easiness; nonce += 16 {
		siphashPRF64(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, uint64(nonce), 0, &nodes)
		for i, n := range nodes {
			n0 := int(n)
			u0 := (n0 & edgemask) << 1
			u0x := xmask(u0)
			u0yz := (u0 & 0x1ffffe) << (ybits + zbits - 1)
			non := (nonce + i) << 1
			y := (non >> (edgebits - ybits + 1))
			v0 := (non&0xfffff)<<(ybits+zbits-1) | u0yz
			size0 := c.size[u0x][y]
			c.matrix[u0x][y][size0][0] = byte(v0 & 0xff)
			c.matrix[u0x][y][size0][1] = byte((v0 >> 8) & 0xff)
			c.matrix[u0x][y][size0][2] = byte((v0 >> 16) & 0xff)
			c.matrix[u0x][y][size0][3] = byte((v0 >> 24) & 0xff)
			c.matrix[u0x][y][size0][4] = byte((v0 >> 32) & 0xff)
			c.size[u0x][y]++
		}
	}
}
*/
// func (c *cuckoo) trim0() {
// 	ctr := make(counter, 2*(1<<(xbits+ybits-1)))
// 	vs := make([]uint32, 0, nedge/(int(nx)))
// 	for x := uint8(0); x < nx; x++ {
// 		for y := uint8(0); y < ny; y++ {
// 			index := uint16(x)<<xbits + uint16(y)
// 			for i := uint16(0); i < c.counter[index]; i++ {
// 				var v uint32
// 				idx := (uint64(x)*uint64(ny)+uint64(y))*5 + uint64(i)
// 				v |= uint32(c.matrix[idx])
// 				v |= uint32(c.matrix[idx+1]) << 8
// 				v |= uint32(c.matrix[idx+2]) << 16
// 				v |= uint32(c.matrix[idx+3]&0xf) << 24
// 				ctr.incr(v)
// 				vs = append(vs, v)
// 			}
// 		}
// 		for y := uint8(0); y < ny; y++ {
// 			index := uint16(x)<<xbits + uint16(y)
// 			for i := uint16(0); i < c.counter[index]; i++ {

// 			}
// 		}
// 	}

func (c *cuckoo) buildNodes() {
	var nodesU [16]uint64
	var nodesV [16]uint64
	for nonce := 0; nonce < easiness; nonce += 16 {
		siphashPRF64(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, uint64(nonce), 0, &nodesU)
		siphashPRF64(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, uint64(nonce), 1, &nodesV)
		for i := range nodesU {
			u0 := int(nodesU[i]&edgemask) << 1
			u0x := xmask(u0)
			u0yz := yzmask(u0)

			v0 := int(nodesV[i]&edgemask) << 1
			v0x := xmask(v0)
			v0yz := yzmask(v0)

			val := uint(u0yz)<<20 | uint(v0yz)

			idx := c.size[u0x][v0x]
			// dw := (*[5]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&c.matrix[0][0][0][0])) +
			// 	uintptr(((u0x*ny+v0x)*narray+idx)*5)))

			// dw[0] = byte(val & 0xff)
			// dw[1] = byte((val >> 8) & 0xff)
			// dw[2] = byte((val >> 16) & 0xff)
			// dw[3] = byte((val >> 24) & 0xff)
			// dw[4] = byte((val >> 32) & 0xff)
			// log.Println(val&0xff, c.matrix[u0x][v0x][idx][0])
			// log.Println((val>>8)&0xff, c.matrix[u0x][v0x][idx][1])
			// log.Println((val>>16)&0xff, c.matrix[u0x][v0x][idx][2])
			// log.Println((val>>24)&0xff, c.matrix[u0x][v0x][idx][3])
			// log.Println((val>>32)&0xff, c.matrix[u0x][v0x][idx][4])
			c.matrix[u0x][v0x][idx][1] = byte((val >> 0) & 0xff)
			c.matrix[u0x][v0x][idx][1] = byte((val >> 8) & 0xff)
			c.matrix[u0x][v0x][idx][2] = byte((val >> 16) & 0xff)
			c.matrix[u0x][v0x][idx][3] = byte((val >> 24) & 0xff)
			c.matrix[u0x][v0x][idx][4] = byte((val >> 32) & 0xff)
			c.size[u0x][v0x]++
		}
	}
}

/*
func (c *cuckoo) trimZ(zsort []cell, uorv int, n uint64) {
	counter := make(counter, n*2/8+1)

	for _, ce := range zsort {
		uz := ce.z(uorv)
		counter.incr(uz)
	}
	for _, ce := range zsort {
		uz := ce.z(uorv)
		if counter.isLeaf(uz) {
			ce.kill()
		}
	}
}

func (c *cuckoo) trimYZ1(uorv int, n uint64) {
	ysort := make([]cell, 0, n)
	for base := uint8(0); base < nx; base++ {
		for enum := uint8(0); enum < nx; enum++ {
			cells := c.matrix[base][enum]
			if uorv == 1 {
				cells = c.matrix[enum][base]
			}
			for _, ce := range cells {
				if !ce.alive() {
					continue
				}
				ysort = append(ysort, ce)
			}
		}
		c.trimZ(ysort, uorv, n)
		ysort = ysort[:0]
	}
}

func (c *cuckoo) trimYZ0(uorv int) {
	const nbucket = (nedge / (uint64(nx) * uint64(ny)))

	ysort := make([][]cell, ny)
	for i := range ysort {
		ysort[i] = make([]cell, 0, nbucket)
	}
	for base := uint8(0); base < nx; base++ {
		for enum := uint8(0); enum < nx; enum++ {
			cells := c.matrix[base][enum]
			if uorv == 1 {
				cells = c.matrix[enum][base]
			}
			for _, ce := range cells {
				if !ce.alive() {
					continue
				}
				uy := ce.y(uorv)
				ysort[uy] = append(ysort[uy], ce)
			}
		}
		for _, y := range ysort {
			c.trimZ(y, uorv, uint64(nz))
		}
		for i := range ysort {
			ysort[i] = ysort[i][:0]
		}
	}
}
*/
// func (c *cuckoo) compaction() {
// 	for ux := range c.matrix {
// 		for vx, cs := range c.matrix[ux] {
// 			i := 0
// 			for _, c := range cs {
// 				if c.alive() {
// 					cs[i] = c
// 					i++
// 				}
// 			}
// 			c.matrix[ux][vx], c.original[ux][vx] = cs[i:], cs[:i]
// 			l := i - len(c.matrix[ux][vx])
// 			for ; l > 0; l-- {
// 				c.matrix[ux][vx] = append(c.matrix[ux][vx], 0)
// 			}
// 			c.matrix[ux][vx] = c.matrix[ux][vx][:i]
// 		}
// 	}
// }

// func (c *cuckoo) rename(uorv int) {
// 	for base := uint8(0); base < nx; base++ {
// 		var imax uint16
// 		names := make(map[uint32]uint16)
// 		for enum := uint8(0); enum < nx; enum++ {
// 			var mat, orig []cell
// 			if uorv == 0 {
// 				mat = c.matrix[base][enum]
// 				orig = c.original[base][enum]
// 			} else {
// 				mat = c.matrix[enum][base]
// 				orig = c.original[enum][base]
// 			}
// 			for i, ce := range orig {
// 				yz := ce.catYZ(uorv)
// 				mat[i].yz[uorv].y = 0
// 				if v, e := names[yz]; e {
// 					mat[i].yz[uorv].z = v
// 				} else {
// 					mat[i].yz[uorv].z = imax
// 					imax++
// 				}
// 			}
// 		}
// 	}
// }

// type set map[uint64]struct{}

// func (s set) add(u, v uint32) {
// 	s[uint64(u)<<16|uint64(v)] = struct{}{}
// }
// func (s set) exist(u, v uint32) bool {
// 	_, exist := s[uint64(u)<<32|uint64(v)]
// 	return exist
// }

// func (c *cuckoo) max() uint32 {
// 	var max uint32
// 	for base := uint8(0); base < nx; base++ {
// 		var cnt0, cnt1 uint32
// 		for enum := uint8(0); enum < nx; enum++ {
// 			for _, ce := range c.matrix[base][enum] {
// 				if ce.alive() {
// 					cnt0++
// 				}
// 			}
// 			for _, ce := range c.matrix[enum][base] {
// 				if ce.alive() {
// 					cnt1++
// 				}
// 			}
// 		}
// 		if max < cnt0 {
// 			max = cnt0
// 		}
// 		if max < cnt1 {
// 			max = cnt1
// 		}
// 	}
// 	return max
// }

// const (
// 	first = iota
// 	second
// 	third
// )

func (c *cuckoo) worker() ([]uint32, bool) {
	// status := first
	c.buildNodes()
	return nil, true
}

// 	log.Println("built matrix")
// 	for i := 0; i < nround; i++ {
// 		switch status {
// 		case first:
// 			c.trimYZ0(0)
// 			c.trimYZ0(1)
// 			max := c.max()
// 			log.Println(max)
// 			if max < nrename0 {
// 				c.compaction()
// 				c.rename(0)
// 				c.rename(1)
// 				status = second
// 			}
// 		case second:
// 			c.trimYZ1(0, uint64(nrename0))
// 			c.trimYZ1(1, uint64(nrename0))
// 			max := c.max()
// 			if max < nrename1 {
// 				c.rename(0)
// 				c.rename(1)
// 				status = third
// 			}
// 		case third:
// 			c.trimYZ1(0, uint64(nrename1))
// 			c.trimYZ1(1, uint64(nrename1))
// 		}
// 	}
// 	if status != third {
// 		panic("number of edges are not below 15")
// 	}
// 	return c.findCycle()
// }

// func (c *cuckoo) solution(us []edgemap, vs []edgemap) ([]uint32, bool) {
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
// 	cycle.add(us[0].orig, vs[0].orig)
// 	for nu--; nu > 0; nu-- {
// 		cycle.add(us[(nu+1)&(^uint16(1))].orig, us[nu|1].orig)
// 	}
// 	for nv--; nv > 0; nv-- {
// 		cycle.add(vs[(nv|1)].orig, vs[(nv+1)&(^uint16(1))].orig)
// 	}
// 	answer := make([]uint32, 0, proofSize)
// 	for nonce := uint32(0); nonce < easiness; nonce++ {
// 		e0, e1 := c.sip.nodes(nonce)
// 		if cycle.exist(e0, e1) {
// 			answer = append(answer, nonce)
// 		}
// 	}
// 	return answer, true
// }

// func (c *cuckoo) path(u edgemap) ([]edgemap, bool) {
// 	us := make([]edgemap, proofSize)
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
