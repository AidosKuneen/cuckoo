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

// This mean solver is a rewrite (from C++ to go) of the mean solver of github.com/tromp/cuckoo,
// which is covered by the FAIR MINING license.

package cuckoo

import (
	"errors"
	"sort"
	"sync"

	"github.com/AidosKuneen/numcpu"
)

type bucket []uint64

const (
	xbits       = 5
	comp0bits   = 32 - edgebits
	comp1bits   = 6
	xmask       = 0x1f
	zmask       = 0x3fff
	comp0mask   = 0xff
	ycomp0mask  = 0x1fff  // 5+8
	xymask      = 0x3ff   // 5+5
	xycomp1mask = 0x1ffff //5+5+6+1

	nx     = 1 << xbits
	zbits  = edgebits - 2*xbits
	nz     = 1 << zbits
	bigeps = nz + nz*3/64
)

//Cuckoo is   struct for cuckoo miner.
type Cuckoo struct {
	cuckoo []uint32
	sip    *sip
	matrix [nx][nx]bucket
	ncpu   int
	m2     [][nx]bucket
}

//NewCuckoo reeturns Cuckoo struct to do PoW.
func NewCuckoo() *Cuckoo {
	c := &Cuckoo{
		cuckoo: make([]uint32, 1<<17+1),
		ncpu:   numcpu.NumCPU(),
	}
	if c.ncpu > 32 {
		c.ncpu = 32
	}
	c.m2 = make([][nx]bucket, c.ncpu)
	for i := 0; i < c.ncpu; i++ {
		for x := 0; x < nx; x++ {
			c.m2[i][x] = make(bucket, 0, bigeps)
		}
	}
	for x := 0; x < nx; x++ {
		for y := 0; y < nx; y++ {
			c.matrix[x][y] = make([]uint64, 0, bigeps)
		}
	}
	return c
}

var pathPool = sync.Pool{
	New: func() interface{} {
		sl := make([]uint32, 0, maxpath)
		return &sl
	},
}

func (c *Cuckoo) path(u uint32) ([]uint32, error) {
	us := *(pathPool.Get().(*[]uint32))
	us = us[:0]
	nu := 0
	for ; u != 0; nu++ {
		if nu >= maxpath {
			return nil, errors.New("more than maxpath")
		}
		us = append(us, u)
		u = c.cuckoo[u&xycomp1mask]
	}
	return us, nil
}

type edges struct {
	edge   []uint64
	uxymap []bool
}

func newedges() *edges {
	return &edges{
		edge:   make([]uint64, 0, ProofSize),
		uxymap: make([]bool, 1<<(2*xbits)),
	}
}

func (e *edges) add(u, v uint32) {
	u >>= 1
	uz := u >> (2*xbits + comp1bits)
	uxy := (u >> comp1bits) & xymask
	ru := (uxy << zbits) | uz
	e.uxymap[uxy] = true
	v >>= 1
	vz := v >> (2*xbits + comp1bits)
	vxy := (v >> comp1bits) & xymask
	rv := (vxy << zbits) | vz
	e.edge = append(e.edge, (uint64(ru)<<32)|uint64(rv))
}
func (e *edges) find(uv uint64, min, max int) bool {
	if max < min {
		return false
	}
	mid := (min + max) / 2
	if e.edge[mid] > uv {
		return e.find(uv, min, mid-1)
	}
	if e.edge[mid] < uv {
		return e.find(uv, mid+1, max)
	}
	return true
}
func (c *Cuckoo) solution(us []uint32, vs []uint32) ([]uint32, bool) {
	nu := int32(len(us) - 1)
	nv := int32(len(vs) - 1)
	min := nu
	if min > nv {
		min = nv
	}
	nv -= min
	nu -= min
	for us[nu] != vs[nv] {
		nu++
		nv++
	}
	l := nu + nv + 1
	if l != ProofSize {
		return nil, false
	}

	es := newedges()
	es.add(us[0], vs[0])
	for nu--; nu >= 0; nu-- {
		es.add(us[(nu+1)&^1], us[nu|1])
	}
	for nv--; nv >= 0; nv-- {
		es.add(vs[nv|1], vs[(nv+1)&^1])
	}
	sort.Slice(es.edge, func(i, j int) bool {
		return es.edge[i] < es.edge[j]
	})
	answer := make([]uint32, 0, ProofSize)
	steps := easiness / c.ncpu
	remain := easiness - steps*c.ncpu
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for j := 0; j < c.ncpu; j++ {
		wg.Add(1)
		go func(j int) {
			var nodesU [8192]uint64
			last := uint64(steps * (j + 1))
			if j == c.ncpu-1 {
				last += uint64(remain)
			}
		loop:
			for nonce := uint64(steps * j); nonce < last; nonce += 8192 {
				siphashPRF8192Seq(&c.sip.v, nonce, 0, &nodesU)
				for i := uint64(0); i < 8192; i++ {
					u0 := nodesU[i] & edgemask
					if es.uxymap[(u0>>zbits)&xymask] {
						v0 := siphashPRF(&c.sip.v, ((nonce+i)<<1)|1) & edgemask
						if es.find((u0<<32)|v0, 0, len(es.edge)-1) {
							mutex.Lock()
							answer = append(answer, uint32(nonce+i))
							if len(answer) >= ProofSize {
								mutex.Unlock()
								break loop
							}
							mutex.Unlock()
						}
					}
				}
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	sort.Slice(answer, func(i, j int) bool {
		return answer[i] < answer[j]
	})
	return answer, true
}

func (c *Cuckoo) index(isU bool, x uint32) [nx]*bucket {
	var indexer [nx]*bucket
	if isU {
		for i := 0; i < nx; i++ {
			indexer[i] = &c.matrix[x][i]
		}
		return indexer
	}
	for i := 0; i < nx; i++ {
		indexer[i] = &c.matrix[i][x]
	}
	return indexer
}

func (c *Cuckoo) buildU() {
	var wg sync.WaitGroup
	matrix := make([][nx][nx]bucket, c.ncpu)
	for j := 0; j < c.ncpu; j++ {
		for x := 0; x < nx; x++ {
			for y := 0; y < nx; y++ {
				matrix[j][x][y] = make(bucket, 0, bigeps/c.ncpu)
			}
		}
	}
	steps := easiness / c.ncpu
	remain := easiness - steps*c.ncpu
	for j := 0; j < c.ncpu; j++ {
		wg.Add(1)
		go func(j int) {
			last := uint64((j + 1) * steps)
			if j == c.ncpu-1 {
				last += uint64(remain)
			}
			var nodesU [8192]uint64
			for nonce := uint64(steps * j); nonce < last; nonce += 8192 {
				siphashPRF8192Seq(&c.sip.v, nonce, 0, &nodesU)
				for i := range nodesU {
					u := nodesU[i] & edgemask
					if u == 0 {
						continue
					}
					ux := (u >> (edgebits - xbits)) & xmask
					uy := (u >> (edgebits - 2*xbits)) & xmask
					v := ((nonce + uint64(i)) << 32) | u
					matrix[j][ux][uy] = append(matrix[j][ux][uy], v)
				}
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	for j := 0; j < c.ncpu; j++ {
		for x := 0; x < nx; x++ {
			for y := 0; y < nx; y++ {
				c.matrix[x][y] = append(c.matrix[x][y], matrix[j][x][y]...)
			}
		}
	}
}

func (c *Cuckoo) buildV() int {
	var wg sync.WaitGroup
	num := make([]int, c.ncpu)
	steps := nx / c.ncpu
	remain := nx - steps*c.ncpu
	for j := 0; j < c.ncpu; j++ {
		wg.Add(1)
		go func(j int) {
			var nodesV [8192]uint64
			var nonces [8192]uint64
			var us [8192]uint64
			var m2 [nx]bucket
			for i := range m2 {
				m2[i] = make([]uint64, 0, bigeps)
			}
			last := (j + 1) * steps
			if j == c.ncpu-1 {
				last += remain
			}
			for ux := j * steps; ux < last; ux++ {
				mu := c.matrix[ux]
				nsip := 0
				for _, m := range mu {
					var cnt [nz]byte
					for _, nu := range m {
						cnt[nu&zmask]++
					}
					for _, nu := range m {
						if cnt[nu&zmask] == 1 {
							continue
						}
						num[j]++
						nonces[nsip] = nu >> 32
						us[nsip] = nu << 32
						if nsip++; nsip == 8192 {
							nsip = 0
							siphashPRF8192(&c.sip.v, &nonces, 1, &nodesV)
							for i, v := range nodesV {
								v &= edgemask
								vx := (v >> (edgebits - xbits)) & xmask
								m2[vx] = append(m2[vx], us[i]|v)
							}
						}
					}
				}
				siphashPRF8192(&c.sip.v, &nonces, 1, &nodesV)
				for i := 0; i < nsip; i++ {
					v := nodesV[i] & edgemask
					vx := (v >> (edgebits - xbits)) & xmask
					m2[vx] = append(m2[vx], us[i]|v)
				}
				c.matrix[ux], m2 = m2, c.matrix[ux]
				for i := range m2 {
					m2[i] = m2[i][:0]
				}
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	number := 0
	for j := 0; j < c.ncpu; j++ {
		number += num[j]
	}
	return number
}

func (c *Cuckoo) trim(isU bool) (int, int) {
	var wg sync.WaitGroup
	num := make([]int, c.ncpu)
	maxbucket := make([]int, c.ncpu)
	steps := nx / c.ncpu
	remain := nx - steps*c.ncpu
	for j := 0; j < c.ncpu; j++ {
		wg.Add(1)
		go func(j int) {
			last := uint32((j + 1) * steps)
			if j == c.ncpu-1 {
				last += uint32(remain)
			}
			for ux := uint32(j * steps); ux < last; ux++ {
				indexer := c.index(isU, ux)
				for vx := uint32(0); vx < nx; vx++ {
					m := indexer[vx]
					for _, uv := range *m {
						y := (uv >> (edgebits - 2*xbits)) & xmask
						c.m2[j][y] = append(c.m2[j][y], uv)
					}
					*m = (*m)[:0]
				}
				for i, m2y := range c.m2[j] {
					var cnt [nz]byte
					for _, uv := range m2y {
						cnt[uv&zmask]++
					}
					nbucket := 0
					for _, uv := range m2y {
						if cnt[uv&zmask] == 1 {
							continue
						}
						nbucket++
						num[j]++
						vu := uv >> 32
						vux := (vu >> (edgebits - xbits)) & xmask
						ruv := (uv << 32) | vu
						m := indexer[vux]
						*m = append(*m, ruv)
					}
					c.m2[j][i] = c.m2[j][i][:0]
					if maxbucket[j] < nbucket {
						maxbucket[j] = nbucket
					}
				}
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	number := 0
	maxb := 0
	for j := 0; j < c.ncpu; j++ {
		number += num[j]
		if maxb < maxbucket[j] {
			maxb = maxbucket[j]
		}
	}
	return number, maxb
}

func (c *Cuckoo) trimrename0(isU bool) int {
	num := 0
	for ux := uint32(0); ux < nx; ux++ {
		indexer := c.index(isU, ux)
		for vx := uint32(0); vx < nx; vx++ {
			m := indexer[vx]
			for _, uv := range *m {
				y := (uv >> (edgebits - 2*xbits)) & xmask
				c.m2[0][y] = append(c.m2[0][y], uv)
			}
			*m = (*m)[:0]
		}
		for i, m2y := range c.m2[0] {
			var nodeid byte
			var cnt [nz]byte
			for _, uv := range m2y {
				cnt[uv&zmask]++
			}
			for _, uv := range m2y {
				cntv := cnt[uv&zmask]
				if cntv == 1 {
					continue
				}
				num++
				var myid byte
				if cntv >= 32 {
					myid = cntv - 32
				} else {
					myid = nodeid
					cnt[uv&zmask] = 32 + nodeid
					nodeid++
				}
				newuv := uv & 0xffffffff
				newuv >>= zbits
				newuv |= (uv & zmask) << (2 * xbits)
				newuv <<= comp0bits
				newuv |= uint64(myid)
				vu := uv >> 32
				allbits := uint(edgebits)
				if isU {
					allbits = 2*xbits + comp0bits
				}
				vux := (vu >> (allbits - xbits)) & xmask
				ruv := (newuv << 32) | vu
				m := indexer[vux]
				*m = append(*m, ruv)
			}
			c.m2[0][i] = c.m2[0][i][:0]
		}
	}
	return num
}

func (c *Cuckoo) trim2(isU bool) int {
	num := 0
	m2 := make([]uint64, 0, bigeps)
	for ux := uint32(0); ux < nx; ux++ {
		var cnt [1 << (xbits + comp0bits)]byte
		indexer := c.index(isU, ux)
		for vx := uint32(0); vx < nx; vx++ {
			m := indexer[vx]
			for _, uv := range *m {
				cnt[uv&ycomp0mask]++
			}
		}
		for vx := uint32(0); vx < nx; vx++ {
			m := indexer[vx]
			for i := len(*m) - 1; i >= 0; i-- {
				uv := (*m)[i]
				if cnt[uv&ycomp0mask] == 1 {
					continue
				}
				num++
				m2 = append(m2, (uv<<32)|(uv>>32))
			}
			*m, m2 = m2, *m
			m2 = m2[:0]
		}
	}
	return num
}

func (c *Cuckoo) trimrename1(isU bool) int {
	num := 0
	for ux := uint32(0); ux < nx; ux++ {
		indexer := c.index(isU, ux)
		for vx := uint32(0); vx < nx; vx++ {
			m := indexer[vx]
			for _, uv := range *m {
				y := (uv >> comp0bits) & xmask
				c.m2[0][y] = append(c.m2[0][y], uv)
			}
			*m = (*m)[:0]
		}
		for i, m2y := range c.m2[0] {
			var nodeid byte
			var cnt [nz]byte
			for _, uv := range m2y {
				cnt[uv&comp0mask]++
			}
			for _, uv := range m2y {
				cntv := cnt[uv&comp0mask]
				if cntv == 1 {
					continue
				}
				num++
				var myid byte
				if cntv >= 32 {
					myid = cntv - 32
				} else {
					myid = nodeid
					cnt[uv&comp0mask] = 32 + nodeid
					nodeid++
				}
				newuv := uv & 0xffffffff
				newuv >>= comp0bits
				newuv <<= comp1bits
				newuv |= uint64(myid)
				vu := uv >> 32
				nbits := uint(comp0bits)
				if isU {
					nbits = comp1bits
				}
				vux := (vu >> (nbits + xbits)) & xmask
				ruv := (newuv << 32) | vu
				m := indexer[vux]
				*m = append(*m, ruv)
			}
			c.m2[0][i] = c.m2[0][i][:0]
		}
	}
	return num
}

func (c *Cuckoo) trimmimng() {
	var i int
	c.buildU()
	c.buildV()
	_, maxv := c.trim(false)
	_, maxu := c.trim(true)
	for i = 3; maxu > 1<<(comp0bits+1) || maxv > 1<<(comp0bits+1); i += 2 {
		_, maxv = c.trim(false)
		_, maxu = c.trim(true)
	}
	c.trimrename0(false)
	c.trimrename0(true)
	for i += 2; i < 65; i += 2 {
		c.trim2(false)
		c.trim2(true)
	}
	c.trimrename1(false)
	c.trimrename1(true)
}

//PoW does PoW with hash, which is the key for siphash.
func (c *Cuckoo) PoW(sipkey []byte) ([]uint32, bool) {
	for i := 0; i < c.ncpu; i++ {
		for x := 0; x < nx; x++ {
			c.m2[i][x] = c.m2[i][x][:0]
		}
	}
	for x := 0; x < nx; x++ {
		for y := 0; y < nx; y++ {
			c.matrix[x][y] = c.matrix[x][y][:0]
		}
	}
	for i := range c.cuckoo {
		c.cuckoo[i] = 0
	}

	c.sip = newsip(sipkey)
	c.trimmimng()

	for _, ux := range c.matrix {
		for _, m := range ux {
			for _, uv := range m {
				u := uint32(uv>>32) << 1
				v := (uint32(uv) << 1) | 1
				us, err1 := c.path(u)
				vs, err2 := c.path(v)
				if err1 != nil || err2 != nil {
					continue
				}
				if us[len(us)-1] == vs[len(vs)-1] {
					if ans, ok := c.solution(us, vs); ok {
						return ans, true
					}
					continue
				}
				if len(us) < len(vs) {
					for nu := len(us) - 2; nu >= 0; nu-- {
						c.cuckoo[us[nu+1]&xycomp1mask] = us[nu]
					}
					c.cuckoo[u&xycomp1mask] = v
				} else {
					for nv := len(vs) - 2; nv >= 0; nv-- {
						c.cuckoo[vs[nv+1]&xycomp1mask] = vs[nv]
					}
					c.cuckoo[v&xycomp1mask] = u
				}
				pathPool.Put(&us)
				pathPool.Put(&vs)
			}
		}
	}
	return nil, false
}
