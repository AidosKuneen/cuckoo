[![Build Status](https://travis-ci.org/AidosKuneen/cuckoo.svg?branch=master)](https://travis-ci.org/AidosKuneen/cuckoo)
[![GoDoc](https://godoc.org/github.com/AidosKuneen/cuckoo?status.svg)](https://godoc.org/github.com/AidosKuneen/cuckoo)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/AidosKuneen/cuckoo/master/LICENSE)


Cuckoo Cycle
=====

## Overview

This library is [Cockoo Cycle](https://github.com/tromp/cuckoo) implementation in Go 
mentiond in [this paper](https://github.com/tromp/cuckoo/blob/master/doc/cuckoo.pdf).

In short, 

"Cuckoo Cycle is the first graph-theoretic proof-of-work, and the most memory bound, yet with instant verification"


to prevent ASICs to get half of hash power easier in network, without heavy work for proover.

This library uses below parameters.

* cycle = 20 to minimize the size impact to transactions.
* bits of nodes (log2(#nodes)) = 25, memory usage should be around 128MB.
* easiness = 50% , with edge trimming.

PoW should be finished in around 800mS if it doesn't find solution, 
and 1 second with solution. So hash based PoW should be added with the PoW.

## Requirements

* git
* go 1.9+

are required to compile.


## Install
    $ go get -u github.com/AidosKuneen/cuckoo


## Usage

```go
	import "github.com/AidosKuneen/cuckoo"
	hash :=[]byte{"some data, which should be 32 bytes"}
	nonces, found := cuckoo.PoW(hash, func(nonces *[ProofSize]uint32) bool {
		//additional PoW (e.g. hash-based PoW) with nonces
		return true //or return false
	}))
	if !found{
		//retry with another hash
	}

	if !cuckoo.Verify(hash, nonces){
		//failed to verify
	}
```
