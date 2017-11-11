package cuckoo

type cuckoo struct {
	cuckoo [nnode + 1]uint32
	sip    *sip
}

func newCuckoo(header []byte) *cuckoo {
	return &cuckoo{
		sip: newsip(header),
	}
}

func (c *cuckoo) path(u uint32, us *[maxpath]uint32) (int, bool) {
	nu := 0
	for ; u != 0; nu++ {
		if nu >= maxpath {
			return 0, false
		}
		us[nu] = u
		u = c.cuckoo[u]
	}
	return nu, true
}

type set map[uint64]struct{}

func (s set) add(u, v uint32) {
	s[uint64(u)<<32+uint64(v)] = struct{}{}
}
func (s set) exist(u, v uint32) bool {
	_, exist := s[uint64(u)<<32+uint64(v)]
	return exist
}

func (c *cuckoo) solution(us *[maxpath]uint32, sizeU int, vs *[maxpath]uint32, sizeV int) (*[proofSize]uint32, bool) {
	nu := int32(sizeU - 1)
	nv := int32(sizeV - 1)
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
	if l != proofSize {
		return nil, false
	}

	cycle := make(set)
	cycle.add(us[0], vs[0])
	for nu--; nu >= 0; nu-- {
		cycle.add(us[(nu+1)&0x1ffffffe], us[nu|1])
	}
	for nv--; nv >= 0; nv-- {
		cycle.add(vs[nv|1], vs[(nv+1)&0x1ffffffe])
	}
	var answer [proofSize]uint32
	var nodesU [16]uint64
	var nodesV [16]uint64
	idx := 0
	for nonce := uint64(0); nonce < easiness; nonce += 16 {
		siphashPRF16Seq(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, nonce, 0, &nodesU)
		siphashPRF16Seq(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, nonce, 1, &nodesV)
		for i := uint64(0); i < 16; i++ {
			u0 := uint32(nodesU[i]&edgemask) << 1
			v0 := uint32((nodesV[i]&edgemask)<<1) | 1
			if cycle.exist(u0, v0) {
				answer[idx] = uint32(nonce + i)
				idx++
			}
		}
	}
	return &answer, true
}

func (c *cuckoo) worker() (*[proofSize]uint32, bool) {
	var nodesU, nodesV [16]uint64
	var us, vs [maxpath]uint32

	for nonce := uint64(0); nonce < easiness; nonce += 16 {
		siphashPRF16Seq(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, nonce, 0, &nodesU)
		siphashPRF16Seq(c.sip.v0, c.sip.v1, c.sip.v2, c.sip.v3, nonce, 1, &nodesV)
		for i := range nodesU {
			u0 := uint32(nodesU[i]&edgemask) << 1
			if u0 == 0 {
				continue
			}
			v0 := uint32((nodesV[i]&edgemask)<<1) | 1
			var sizeU, sizeV int
			var okU, okV bool
			sizeU, okU = c.path(u0, &us)
			sizeV, okV = c.path(v0, &vs)
			if !okU || !okV {
				continue
			}
			if us[sizeU-1] == vs[sizeV-1] {
				if nonce+uint64(i) >= minnonce {
					if ans, ok := c.solution(&us, sizeU, &vs, sizeV); ok {
						return ans, true
					}
				}
				continue
			}
			if sizeU < sizeV {
				for nu := sizeU - 2; nu >= 0; nu-- {
					c.cuckoo[us[nu+1]] = us[nu]
				}
				c.cuckoo[u0] = v0
			} else {
				for nv := sizeV - 2; nv >= 0; nv-- {
					c.cuckoo[vs[nv+1]] = vs[nv]
				}
				c.cuckoo[v0] = u0
			}
		}
	}
	return nil, false
}
