package util

import (
	"log"
	"math/rand"
	"os"
	"reflect"
	"unsafe"

	"mobile-contact-discovery/pir"
)

// https://github.com/cookieo9/go-misc/blob/master/slice/cap.go
var (
	_ShrinkCapacityInvalidType = "slice.ShrinkCapacity: argument 1 not pointer to a slice"
	_ShrinkCapacityIncrease    = "slice.ShrinkCapacity: attempt to increase capacity"
	_ShrinkCapacityNegative    = "slice.ShrinkCapacity: negative target capacity"
)

// func CreateTinyCF(elements []ck_pir.Row, cf_size uint) (cf *cuckoo.Filter) {
// 	cf = cuckoo.NewFilter(3, 32, cf_size, cuckoo.TableTypeSingle)
// 	for _, el := range elements {
// 		cf.Add(el)
// 	}
// 	return

// }

func RandomElements(src *rand.Rand, numElems uint32) []uint64 {
	elem := make([]uint64, numElems)
	for i := range elem {
		elem[i] = src.Uint64()
	}
	return elem
}

func RandomPIRElements(src *rand.Rand, numElems uint32, rowLen uint32) []pir.Row {
	elem := make([]pir.Row, numElems)
	for i := range elem {
		elem[i] = make([]byte, rowLen)
		src.Read(elem[i])
	}
	return elem
}

func Random_CFValues(src *rand.Rand, first bool, prfType string, numElems uint32, numBuckets uint32) *[][3]uint {
	cfIndices := make([][3]uint, numElems)
	for i := range cfIndices {
		var elem []byte
		if first {
			if prfType == "ECNR" {
				elem = []byte{3, 92, 104, 89, 218, 87, 92, 226, 29, 42, 163, 41, 226, 242, 100, 59, 118, 163, 115, 232, 221, 213, 255, 113, 165, 14, 75, 228, 189, 227, 145, 101, 104}
			} else if prfType == "GCAES" {
				elem = []byte{202, 221, 159, 223, 21, 65, 94, 159, 190, 34, 215, 212, 149, 216, 194, 21}
			} else if prfType == "GCLOWMC" {
				elem = []byte{158, 0, 140, 224, 49, 111, 224, 116, 204, 71, 147, 66, 153, 207, 180, 110}
			}
			first = false
		} else {
			if prfType == "ECNR" {
				elem = make([]byte, 33)
			} else {
				elem = make([]byte, 16)
			}
			src.Read(elem)
		}
		_, id0, id1, t := GenerateIndicesTagHash(elem, numBuckets)
		cfIndices[i] = [3]uint{uint(t), id0, id1}
	}
	return &cfIndices
}

func WriteBytesToFile(filename string, b []byte) {
	err := os.WriteFile(filename, b, 0777)
	if err != nil {
		log.Fatalln("Error writing bytes to file")
	}
}

func ShrinkCapacity(slicePointer interface{}, capacity int) {
	pointerValue := reflect.ValueOf(slicePointer)

	if t := pointerValue.Type(); t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
		panic(_ShrinkCapacityInvalidType)
	}

	sh := (*reflect.SliceHeader)(unsafe.Pointer(pointerValue.Pointer()))

	// Prevent increasing capacity
	if sh.Cap < capacity {
		panic(_ShrinkCapacityIncrease)
	}

	if capacity < 0 {
		panic(_ShrinkCapacityNegative)
	}

	// Enforce output len <= cap
	sh.Cap = capacity
	if sh.Len > sh.Cap {
		sh.Len = sh.Cap
	}
}
