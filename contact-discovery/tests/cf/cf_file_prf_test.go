package test

import (
	"contact-discovery/oprf_c"
	"contact-discovery/util"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/linvon/cuckoo-filter"
)

// Test Write CF to File and Read from File
func TestCuckooFilePRF(t *testing.T) {
	fmt.Println("===== TESTCuckooFilePRF =====")
	filename := "cftest"

	exp := 10
	var prfOut *[][]byte
	prfOut = oprf_c.PRF(int(1)<<exp, "ECNR", true, 1)
	fmt.Println((*prfOut)[0])

	dbSize := uint32(int(1) << exp)
	// Create CF
	cf := cuckoo.NewFilter(util.TagsPerBucket, util.BitsPerItem,
		uint(dbSize), cuckoo.TableTypeSingle)

	for _, item := range *prfOut {
		res := cf.Add(item)
		if !res {
			log.Fatalln("[PRF]\tInsertion into CF failed!")
		}
	}
	fmt.Println("[PRF]\tCF with ", dbSize, " PRF elements")

	util.CFToFile(cf, filename)

	cf_file := util.CFfromFile(filename)

	if !reflect.DeepEqual(cf, cf_file) {
		t.Errorf("Expected epual cf")
		return
	}

	fmt.Println("CF:\t", cf.Info())
	fmt.Println("CF from file:\t", cf_file.Info())

}
