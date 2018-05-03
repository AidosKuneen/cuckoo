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

PoW should be finished in 1 second, so hash based PoW should be added with the PoW.
The probability of suceeding to find a solution is around 5% and variant is
also around 5%. So the solution should be found in 20 seconds and 3 sigma
is 60 seconds.


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

# Expected PoW Time

Using the following test environment...

```
* Compiler: go version go1.10 linux/amd64
* Kernel: Linux WS777 4.13.5-1-ARCH #1 SMP PREEMPT Fri Oct 6 09:58:47 CEST 2017 x86_64 GNU/Linux
* CPU:  Celeron(R) CPU G1840 @ 2.80GHz 
* Memory: 8 GB
```

PoW takes around 630 mS.


```
BenchmarkCuckoo2-2   	       2	 633949626 ns/op	244375408 B/op	   30199 allocs/op
PASS
```



On a cloud server:

```
* Compiler: go version go1.8.1 linux/amd64
* Kernel: Linux 4.8.0-58-generic #63~16.04.1-Ubuntu SMP Mon Jun 26 18:08:51 UTC 2017 x86_64 x86_64 x86_64 GNU/Linux
* CPU:  CAMD Ryzen 7 1700X Eight-Core Processor @ 2.20GHz (16 cores)
* Memory: 64 GB
```

PoW takes around 330 mS.

```
BenchmarkCuckoo2-16    	       5	 332289248 ns/op	293002734 B/op	   47726 allocs/op
PASS
```

On DIGNO M KYL22(Android Smartphone):



```
* Compiler: go version go1.10 linux/arm
* OS: 	Android 4.2.2
* CPU:	Qualcomm Snapdragon 800 MSM8974 2.2GHz (quad core)
* Memory: 2 GB
```

PoW takes around 2.6 seconds.


```
BenchmarkCuckoo2 	       1	2580786889 ns/op
PASS
```