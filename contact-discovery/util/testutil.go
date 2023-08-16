package util

import (
	pir "contact-discovery/pir"
	"log"
	"math/rand"
	"os"
	"reflect"
	"unsafe"

	"github.com/linvon/cuckoo-filter"
)

// https://github.com/cookieo9/go-misc/blob/master/slice/cap.go
var (
	_ShrinkCapacityInvalidType = "slice.ShrinkCapacity: argument 1 not pointer to a slice"
	_ShrinkCapacityIncrease    = "slice.ShrinkCapacity: attempt to increase capacity"
	_ShrinkCapacityNegative    = "slice.ShrinkCapacity: negative target capacity"
)

func CreateTinyCF(elements []pir.Row, cf_size uint) (cf *cuckoo.Filter) {
	cf = cuckoo.NewFilter(3, 32, cf_size, cuckoo.TableTypeSingle)
	for _, el := range elements {
		cf.Add(el)
	}
	return
}

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

func WriteBytesToFile(filename string, b []byte) {
	err := os.WriteFile(filename, b, 0777)
	if err != nil {
		log.Fatalln("Error writing bytes to file")
	}
}

// https://github.com/cookieo9/go-misc/blob/master/slice/cap.go
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
