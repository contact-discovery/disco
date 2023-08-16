package test

import (
	"bytes"
	"contact-discovery/pir"
	"contact-discovery/util"
	"fmt"
	"log"
	"math"
	"reflect"
	"testing"
)

func doPir(queryRow int, db *pir.StaticDB) pir.Row {

	// ===== OFFLINE PHASE =====
	//    Client asks for offline hint
	offlineReq := pir.NewPuncHintReq()

	//    Server responds with hint
	offlineResp, err := offlineReq.Process(*db)
	if err != nil {
		log.Fatal("Offline hint generation failed")
	}

	// Initialize the client state
	client := offlineResp.InitClient(pir.RandSource(), 0.99)

	// ===== ONLINE PHASE =====
	//    Client generates queries for servers
	queries, recon := client.Query(queryRow)

	//    Servers answer queries
	answers := make([]interface{}, len(*queries))
	for i := 0; i < len(*queries); i++ {
		answers[i], err = (*queries)[i].Process(*db)
		if err != nil {
			log.Fatal("Error answering query")
		}
	}

	//    Client reconstructs
	row, err := recon(answers)
	if err != nil {
		log.Fatal("Could not reconstruct")
	}
	return row
}

// Test Write CF to File and Read from File
func TestPIROnSegment(t *testing.T) {
	fmt.Println("===== TESTPIROnSegment =====")
	dbSize := 512 // 2**9
	elements := util.RandomPIRElements(pir.RandSource(), uint32(dbSize), 4)
	// Check if segments match CF Table
	cf := util.CreateTinyCF(elements, uint(dbSize)) //only 4 buckets

	all_buckets := cf.Buckets()
	half := cf.NumBuckets() / 2
	var segments [2]*pir.StaticDB
	segments[0] = util.CFSegment(cf, 0, uint32(half))
	segments[1] = util.CFSegment(cf, 1, uint32(half))
	util.ShrinkCapacity(&segments[0].FlatDb, len(segments[0].FlatDb))
	util.ShrinkCapacity(&segments[1].FlatDb, len(segments[1].FlatDb))
	h0 := (all_buckets)[0 : 12*half]
	h1 := (all_buckets)[12*half:]
	util.ShrinkCapacity(&h0, len(h0))
	util.ShrinkCapacity(&h1, len(h1))
	if !reflect.DeepEqual(h0, segments[0].FlatDb) {
		t.Errorf("Expected seg1 to equal the first half of cf table")
		return
	}
	if !reflect.DeepEqual(h1, segments[1].FlatDb) {
		t.Errorf("Expected seg2 to equal the second half of cf table")
		return
	}
	// Query a segment to test if PIR is working
	toQuery := []int{110, 138}
	for _, rIdx := range toQuery {

		queryRow := rIdx % int(half)
		numSeg := int(math.Floor(float64(rIdx) / float64(half)))

		row := doPir(int(queryRow), segments[numSeg])

		if !bytes.Equal(row, segments[numSeg].Row(queryRow)) {
			log.Fatal("Incorrect answer returned")
		} else {
			fmt.Println("Success: ")
			fmt.Println("DB row:\t", segments[numSeg].Row(queryRow))
			fmt.Println("queried row:\t", row)
		}
	}
}
