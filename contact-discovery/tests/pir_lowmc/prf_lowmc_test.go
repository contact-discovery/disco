package test

import (
	"contact-discovery/oprf_c"
	"fmt"
	"reflect"
	"testing"
)

// Requires local GCLOWMC-OPRF server on port 50051
func Test_PRF(t *testing.T) {
	fmt.Println("===== Test_PRF LowMC=====")

	fmt.Println("Please start GC-LowMC OPRF Server on port 50051")

	/****** GC-LowMC *****/
	prfOutClient := oprf_c.OPRF(1, "GCLOWMC", true, "127.0.0.1", 50051)
	prfOutServer := oprf_c.PRF(1, "GCLOWMC", true, 1)

	first_elem := []byte{158, 0, 140, 224, 49, 111, 224, 116, 204, 71, 147, 66, 153, 207, 180, 110}

	if !reflect.DeepEqual((*prfOutServer)[0], (*prfOutClient)[0]) {
		fmt.Println("Server:\t", (*prfOutServer)[0])
		fmt.Println("Client:\t", (*prfOutClient)[0])
		t.Errorf("GCLowMc PRF Outputs of client and server are not equal!")
	} else {
		if !reflect.DeepEqual(first_elem, (*prfOutClient)[0]) {
			fmt.Println("first_elem:\t", first_elem)
			fmt.Println("PRF:\t", (*prfOutClient)[0])
			t.Errorf("GCLowMc PRF Output and first elem are not equal!")
		} else {
			fmt.Println((*prfOutServer)[0])
			fmt.Println("All GCLowMc PRF outputs are equal!")
		}
	}

}
