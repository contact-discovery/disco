package test

import (
	"contact-discovery/util"
	"fmt"
	"reflect"
	"testing"

	"github.com/linvon/cuckoo-filter"
)

// Test Write CF to File and Read from File
func Testutil(t *testing.T) {
	fmt.Println("===== TESTCuckooFile =====")
	filename := "cftest"

	cf := cuckoo.NewFilter(3, 32, 10000, cuckoo.TableTypeSingle)

	a := []byte("A")
	cf.Add(a)

	util.CFToFile(cf, filename)

	cf_file := util.CFfromFile(filename)

	if !reflect.DeepEqual(cf, cf_file) {
		t.Errorf("Expected epual cf")
		return
	}

	fmt.Println("CF:\t", cf.Info())
	fmt.Println("CF from file:\t", cf_file.Info())

}
