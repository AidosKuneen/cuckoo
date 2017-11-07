package cuckoo

import (
	"encoding/binary"
	"log"

	"golang.org/x/crypto/blake2b"
)

const (
	edgebits  = 27
	proofSize = 42
	nedge     = 1 << edgebits
	edgemask  = nedge - 1
	nnode     = 2 * nedge
	easipct   = 0.5
)

type sip struct {
	k0 uint64
	k1 uint64
	v0 uint64
	v1 uint64
	v2 uint64
	v3 uint64
}

func newsip(header []byte) *sip {
	h := blake2b.Sum256(header)
	s := &sip{
		k0: binary.LittleEndian.Uint64(h[:]),
		k1: binary.LittleEndian.Uint64(h[8:]),
	}
	s.v0 = s.k0 ^ 0x736f6d6570736575
	s.v1 = s.k1 ^ 0x646f72616e646f6d
	s.v2 = s.k0 ^ 0x6c7967656e657261
	s.v3 = s.k1 ^ 0x7465646279746573
	return s
}

func (s *sip) nodes(nonce uint32) (uint32, uint32) {
	n0, n1 := siphashPRF(s.v0, s.v1, s.v2, s.v3,
		uint64(nonce<<1), uint64(nonce<<1+1))
	return uint32(n0 & edgemask), uint32(n1 & edgemask)
}

type cuckoo struct {
	easiness uint32
	cuckoo   []uint32
	sip      *sip
}

func newCuckoo(header []byte) *cuckoo {
	return &cuckoo{
		easiness: uint32(easipct * nnode),
		cuckoo:   make([]uint32, 1+nnode),
		sip:      newsip(header),
	}
}

func (c *cuckoo) path(u uint32, u0 uint32, us []uint32) bool {
	for nu := 1; u != 0; u = c.cuckoo[u] {
		if nu >= proofSize {
			return false
			// for nu > 0 && us[nu] != u {
			// 	nu--
			// }
			// if nu == 0 {
			// 	return errors.New("exceeds maximum path")
			// }
			// return fmt.Errorf("illegal %d cycle", proofSize-nu)
		}
		us[nu] = u
	}
	return true
}

type set map[uint64]struct{}

func (s set) add(u, v uint32) {
	s[uint64(u)<<32+uint64(v)] = struct{}{}
}
func (s set) exist(u, v uint32) bool {
	_, exist := s[uint64(u)<<32+uint64(v)]
	return exist
}

func (c *cuckoo) solution(us []uint32, nu uint32, vs []uint32, nv uint32) {
	cycle := make(set)
	cycle.add(us[0], vs[0])
	for ; nu > 0; nu-- {
		cycle.add(us[(nu+1)&(^uint32(1))], us[nu|1])
	}
	for ; nv > 0; nv-- {
		cycle.add(vs[(nv|1)], vs[(nv+1)&(^uint32(1))])
	}
	for nonce := uint32(0); nonce < c.easiness; nonce++ {
		e0, e1 := c.sip.nodes(nonce)
		if cycle.exist(e0, e1) {
			log.Println("nonce", nonce)
		}
	}
}

func (c *cuckoo) worker() {
	for nonce := uint32(0); nonce < c.easiness; nonce++ {
		us := make([]uint32, proofSize)
		vs := make([]uint32, proofSize)
		u0, v0 := c.sip.nodes(nonce)
		us[0] = u0
		vs[0] = u0
		if u0 == 0 {
			continue
		}
		u := c.cuckoo[u0]
		v := c.cuckoo[v0]

		if !c.path(u, u0, us) {
			continue
		}
		if !c.path(v, v0, vs) {
			continue
		}
		nu := len(us) - 1
		nv := len(vs) - 1
		if us[nu] == vs[nv] {
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
			if l == proofSize {
				log.Printf("%v cycle found", l)
				c.solution(us, uint32(nu), vs, uint32(nv))
				return
			}
			continue
		}
		if nv < nv {
			for nu > 0 {
				nu--
				c.cuckoo[us[nu+1]] = us[nu]
			}
			c.cuckoo[u0] = v0
		} else {
			for nv > 0 {
				nv--
				c.cuckoo[vs[nv+1]] = vs[nv]
			}
			c.cuckoo[v0] = u0
		}
	}
}

// func main() {
// 	fmt.Printf("Looking for %d-cycle on cuckoo%d with %v edges\n", proofSize, edgebits+1, easipct)
// 	c := newCuckoo([]byte("This is a test for cuckoo PoW."))
// 	c.worker()
// }
