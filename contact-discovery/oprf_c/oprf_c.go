package oprf_c

/*
#cgo linux CXXFLAGS: -I${SRCDIR}/../
#cgo linux LDFLAGS: -L../ -ldroidcrypto -lstdc++
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include "../../mobile_psi_cpp/droidCrypto/psi/oprf/prf.h"
#include "../../mobile_psi_cpp/droidCrypto/psi/oprf/oprf.h"
*/
import "C"
import (
	"contact-discovery/util"
	"fmt"
	"log"
	"math"
	"time"
	"unsafe"

	"github.com/linvon/cuckoo-filter"
)

func CF_PRF(cf *cuckoo.Filter, prfType string, numElements int, first bool, num_threads int) {
	var cfTime time.Duration = 0
	var prfTime time.Duration = 0
	var elemSize int
	var maxElem int

	if prfType == "ECNR" {
		elemSize = 33
		maxElem = 1048576 //2^20
	} else if prfType == "GCAES" || prfType == "GCLOWMC" {
		elemSize = 16
		maxElem = 67108864 //2^26
	} else {
		log.Fatalln("Please select valid prf type")
	}

	var iterations int
	var numIterElems int
	var firstRunElems int
	if numElements > maxElem {
		// multiple PRF gens iterations
		iterations = int(math.Ceil(float64(numElements / maxElem)))
		firstRunElems = maxElem
		if numElements%maxElem != 0 {
			firstRunElems = numElements % maxElem
		}
		numIterElems = maxElem
	} else {
		iterations = 1
		numIterElems, firstRunElems = numElements, numElements
	}
	results := C.malloc(C.ulong(numIterElems * elemSize))
	var numRunElements int
	fmt.Println("interations: ", iterations)
	for i := 0; i < iterations; i++ {
		fmt.Println("interation: ", i)
		if i > 0 {
			first = false
			numRunElements = numIterElems
		} else {
			numRunElements = firstRunElems
		}
		startPRF := time.Now()
		if prfType == "ECNR" {
			C.getECNR_PRF(C.int(numRunElements), C.bool(first), C.int(num_threads), (*C.uchar)(results))
		} else if prfType == "GCAES" {
			C.getGCAES_PRF(C.int(numRunElements), C.bool(first), C.int(num_threads), (*C.uchar)(results))
		} else if prfType == "GCLOWMC" {
			C.getGCLowMC_PRF(C.int(numRunElements), C.bool(first), C.int(num_threads), (*C.uchar)(results))
		}
		prfTime += time.Since(startPRF)
		prfBytes := C.GoBytes(results, C.int(numRunElements*elemSize))

		startCF := time.Now()
		for j := 0; j < int(numRunElements); j++ {
			res := cf.Add(prfBytes[j*int(elemSize) : (j+1)*int(elemSize)])
			if !res {
				log.Fatalln("CF\tInsertion into CF failed!")
			}
		}
		cfTime += time.Since(startCF)
	}
	C.free(unsafe.Pointer(results))
	log.Println("CF\tPRF Time:\t", prfTime)
	log.Println("CF\tCF Add Time:\t", cfTime)
	log.Println("CF\tTotal Time:\t", cfTime+prfTime)
}

func PRF(numElements int, prfType string, first bool, num_threads int) *[][]byte {
	var elemSize int

	if prfType == "ECNR" {
		elemSize = 33
	} else if prfType == "GCAES" || prfType == "GCLOWMC" {
		elemSize = 16
	} else {
		log.Fatalln("Please select valid prf type")
	}

	prfOut := make([][]byte, numElements)

	results := C.malloc(C.ulong(numElements * elemSize))
	// Free C memory when prf results are no longer needed
	defer C.free(results)
	if prfType == "ECNR" {
		C.getECNR_PRF(C.int(numElements), C.bool(first), C.int(num_threads), (*C.uchar)(results))
	} else if prfType == "GCAES" {
		C.getGCAES_PRF(C.int(numElements), C.bool(first), C.int(num_threads), (*C.uchar)(results))
	} else if prfType == "GCLOWMC" {
		C.getGCLowMC_PRF(C.int(numElements), C.bool(first), C.int(num_threads), (*C.uchar)(results))
	}

	for i := 0; i < int(numElements); i++ {
		prfOut[i] = C.GoBytes(results, C.int(elemSize))
		results = unsafe.Pointer(uintptr(results) + uintptr(elemSize))
	}
	return &prfOut
}

func OPRF(numElements int, prfType string, first bool, s_addr string, s_port int) *[][]byte {
	var elemSize int

	if prfType == "ECNR" {
		elemSize = 33
	} else if prfType == "GCAES" || prfType == "GCLOWMC" {
		elemSize = 16
	} else {
		log.Fatalln("Please select valid prf type")
	}
	// Free C memory when prf results are no longer needed
	results := C.malloc(C.ulong(numElements * elemSize))
	defer C.free(results)
	prfOut := make([][]byte, numElements)
	if prfType == "ECNR" {
		C.doECNR_OPRF(C.int(numElements), C.bool(first), C.CString(s_addr), C.int(s_port), (*C.uchar)(results))
	} else if prfType == "GCAES" {
		C.doGCAES_OPRF(C.int(numElements), C.bool(first), C.CString(s_addr), C.int(s_port), (*C.uchar)(results))
	} else if prfType == "GCLOWMC" {
		C.doGCLowMC_OPRF(C.int(numElements), C.bool(first), C.CString(s_addr), C.int(s_port), (*C.uchar)(results))
	}
	for i := 0; i < numElements; i++ {
		prfOut[i] = C.GoBytes(results, C.int(elemSize))
		results = unsafe.Pointer(uintptr(results) + uintptr(elemSize))
	}
	return &prfOut
}

func OPRF_CFValues(num_elements int, prfType string, first bool, s_addr string, s_port int, num_buckets uint32) *[][3]uint {
	var elemSize int
	cfIndices := make([][3]uint, num_elements)
	if prfType == "ECNR" {
		elemSize = 33
	} else if prfType == "GCAES" || prfType == "GCLOWMC" {
		elemSize = 16
	} else {
		log.Fatalln("Please select valid prf type")
	}
	results := C.malloc(C.ulong(num_elements * elemSize))
	// Free C memory when prf results are no longer needed
	defer C.free(results)

	if prfType == "ECNR" {
		C.doECNR_OPRF(C.int(num_elements), C.bool(first), C.CString(s_addr), C.int(s_port), (*C.uchar)(results))
	} else if prfType == "GCAES" {
		C.doGCAES_OPRF(C.int(num_elements), C.bool(first), C.CString(s_addr), C.int(s_port), (*C.uchar)(results))
	} else if prfType == "GCLOWMC" {
		C.doGCLowMC_OPRF(C.int(num_elements), C.bool(first), C.CString(s_addr), C.int(s_port), (*C.uchar)(results))
	}
	for i := 0; i < num_elements; i++ {
		_, id0, id1, t := util.GenerateIndicesTagHash(C.GoBytes(results, C.int(elemSize)), num_buckets)
		results = unsafe.Pointer(uintptr(results) + uintptr(elemSize))
		cfIndices[i] = [3]uint{uint(t), id0, id1}
	}
	return &cfIndices
}
