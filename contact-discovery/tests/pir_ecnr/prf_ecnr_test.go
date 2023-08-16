package test

import (
	"contact-discovery/oprf_c"
	"fmt"
	"reflect"
	"testing"
)

// Requires local ECNR-OPRF server on port 50051
func Test_PRF_ECNR(t *testing.T) {
	fmt.Println("===== Test_PRF_ECNR =====")

	/****** EC-NR*****/
	// Read CF from file
	prfOutServer := oprf_c.PRF(20, "ECNR", true, 1)
	prfOutClient := oprf_c.OPRF(1, "ECNR", true, "127.0.0.1", 50051)

	first_elem := []byte{3, 92, 104, 89, 218, 87, 92, 226, 29, 42, 163, 41, 226, 242, 100, 59, 118, 163, 115, 232, 221, 213, 255, 113, 165, 14, 75, 228, 189, 227, 145, 101, 104}

	if !reflect.DeepEqual((*prfOutServer)[0], (*prfOutClient)[0]) {
		fmt.Println("Server:\t", (*prfOutServer)[0])
		fmt.Println("Client:\t", (*prfOutClient)[0])
		t.Errorf("ECNR PRF Outputs of client and server are not equal!")
	} else {
		if !reflect.DeepEqual(first_elem, (*prfOutClient)[0]) {
			fmt.Println("first_elem:\t", first_elem)
			fmt.Println("PRF:\t", (*prfOutClient)[0])
			t.Errorf("ECNR PRF Output and first elem are not equal!")
		} else {
			fmt.Println((*prfOutServer)[0])
			fmt.Println("All ECNR PRF outputs are equal")
		}
	}
}
