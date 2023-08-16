package test

import (
	"contact-discovery/oprf_c"
	"fmt"
	"reflect"
	"testing"
)

// Requires local GCAES-OPRF server on port 50051
func Test_PRF_GCAES(t *testing.T) {
	fmt.Println("===== Test_PRF_GCAES =====")
	prfOutClient := oprf_c.OPRF(1, "GCAES", true, "127.0.01", 50051)
	prfOutServer := oprf_c.PRF(1, "GCAES", true, 1)

	first_elem := []byte{202, 221, 159, 223, 21, 65, 94, 159, 190, 34, 215, 212, 149, 216, 194, 21}

	if !reflect.DeepEqual((*prfOutServer)[0], (*prfOutClient)[0]) {
		fmt.Println("Server:\t", (*prfOutServer)[0])
		fmt.Println("Client:\t", (*prfOutClient)[0])
		t.Errorf("GCAES PRF Outputs of client and server are not equal!")
	} else {
		if !reflect.DeepEqual(first_elem, (*prfOutClient)[0]) {
			fmt.Println("first_elem:\t", first_elem)
			fmt.Println("PRF:\t", (*prfOutClient)[0])
			t.Errorf("GCAES PRF Output and first elem are not equal!")
		} else {
			fmt.Println((*prfOutServer)[0])
			fmt.Println("All GCAES PRF outputs are equal!")
		}
	}

}
