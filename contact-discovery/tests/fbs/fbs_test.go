package test

import (
	fbs "contact-discovery/fbs"
	pir "contact-discovery/pir"
	psi "contact-discovery/psi"
	"contact-discovery/util"
	"fmt"
	"log"
	"reflect"
	"testing"
)

// Requires local GCAES-OPRF server on port 50051
func Test_Fbs(t *testing.T) {
	fmt.Println("===== Test_Fbs =====")

	oprfserver := "127.0.0.1:50051"
	prfType := "GCAES"
	cfFile := "../../cf_files/cf_GCAES_14.data"
	cf := util.CFfromFile(cfFile)
	segsize := int(1) << 10
	npir := int(1) << 13
	numsegs := npir / segsize
	cexp := 10
	numWorker := 4
	threshold := 0.99

	p := psi.NewProvider(cf, uint32(segsize))

	// ==== PIR OFFLINE PHASE ====
	// C generates HintRequest
	c, hintreq := psi.NewClient(segsize, numsegs, npir)
	c_builder_hintreq := psi.EncodeHintReq(hintreq)
	s_hr := fbs.GetRootAsHintReq(psi.FinishEncode(c_builder_hintreq), 0)

	// S processes HintRequest & generates HintResps
	phr := psi.DecodeHintReq(s_hr)

	if !reflect.DeepEqual(hintreq, phr) {
		t.Errorf("Flatbuffer PuncHintReq input and output are not equal")
	}

	hintresp := p.AllHintReponse(phr, numWorker, false)
	s_hresp := psi.EncodeAllHintResp(hintresp)

	// C processes HintResps
	c_hresp := fbs.GetRootAsAllHintResp(psi.FinishEncode(s_hresp), 0)
	c_hintresps := psi.DecodeAllHintResp(c_hresp, numWorker)

	h1 := (*c_hintresps)[0]
	h2 := (*hintresp)[0]
	util.ShrinkCapacity(&h1.Hints, len(h1.Hints))

	if !reflect.DeepEqual(h1, h2) {
		t.Errorf("Flatbuffer PuncHintResp input and output are not equal")
	}

	c.ProcessHintResp(c_hintresps, threshold, numWorker)
	c.RunOPRF(uint16(cexp), true, oprfserver, prfType)

	_, _, _, tag1 := util.GenerateIndicesTagHash(
		[]byte{202, 221, 159, 223, 21, 65, 94, 159, 190, 34, 215, 212, 149, 216, 194, 21},
		c.N_cf)

	if prfType == "GCAES" && (*c.CfIndices)[0][0] != uint(tag1) {
		t.Errorf("PRF Output not correct!")
	}

	for s := range c.ClientPIR {
		if s == 4 || s == 7 { // remove after debugging
			queryReqs := make([][]pir.PuncQueryReq, c.MaxIterations)
			tags := make([]uint32, c.MaxIterations)
			recons := make([]pir.ReconstructFunc, c.MaxIterations)

			i := 0
			for _, indices := range *c.CfIndices {
				for _, idx := range indices[1:] { // indices[0] contains tag
					if int(idx/uint(c.SegSize)) == s {

						//queries[i].Idx = uint32(idx % uint(c.SegSize))
						//queries[i].Tag = uint32(indices[0])
						tags[i] = uint32(indices[0])
						if tags[i] == uint32((*c.CfIndices)[0][0]) {
							fmt.Println("This should be the intersection!")
							fmt.Println("Idx: ", idx, "\tSegNum: ", int(idx/uint(c.SegSize)), ", \tSegIdx:", int(idx%uint(c.SegSize)))
						}
						//queries[i].QueryReq, queries[i].recon = segClient.(*pir.PuncClient).Query(int(idx % uint(c.SegSize)))
						queryReq, recon := c.ClientPIR[s].Query(int(idx % uint(c.SegSize)))
						queryReqs[i], recons[i] = *queryReq, recon
						//queryReqs[i], recons[i] = segClient.Query(int(idx % uint(c.SegSize)))
						i++
					}
				}
			}
			for ; i < c.MaxIterations; i++ {
				queryReqs[i] = *(c.ClientPIR[s].DummyQuery())
			}
			answers_c := make([][]interface{}, c.MaxIterations)
			for c_i := range [2]int{0, 1} {
				// TODO encode all queries together
				b_c_sq := psi.EncodeSegmentQueries(s, &queryReqs, c_i)

				c_sq_fbs := fbs.GetRootAsSegmentQueries(psi.FinishEncode(b_c_sq), 0)
				segNum, s_sq := psi.DecodeSegmentQueries(c_sq_fbs)

				c_sq := (queryReqs)[s][c_i]
				if !reflect.DeepEqual((c_sq), (*s_sq)[s]) {
					t.Errorf("Flatbuffer PuncQueryReq input and output are not equal")
				}
				if !reflect.DeepEqual(segNum, uint32(s)) {
					t.Errorf("Flatbuffer SegNum input and output are not equal")
				}
				answers_s := p.OnlineSegments(int(segNum), s_sq)
				fmt.Println((*answers_s)[0])
				s_sa := psi.FinishEncode(psi.EncodeSegmentAnswers(answers_s))

				c_sa := fbs.GetRootAsSegmentAnswers(s_sa, 0)
				psi.DecodeSegmentAnswer(c_sa, &answers_c, c_i)

				if !reflect.DeepEqual((*answers_s)[s].(*pir.PuncQueryResp), answers_c[s][c_i].(*pir.PuncQueryResp)) {
					t.Errorf("Flatbuffer Answer input and output are not equal")
				}
			}
			for k, ans := range answers_c {
				if tags[k] != 0 { //Only look at real answers, not dummies
					if len(ans) != 2 {
						log.Fatal("[PIR]\tDid not receive response from both servers")
					} else {
						row, err := (recons[k])(ans)
						if err != nil {
							log.Fatal("[PIR]\tCould not reconstruct answer")
						}

						if tags[k] == uint32((*c.CfIndices)[0][0]) {
							//fmt.Println("This should be the intersection!")
							fmt.Println("Row:\t", row)
						}

						for j := 0; j < util.TagsPerBucket; j++ {
							if util.SlotContains(row[j*4:(j+1)*4], tags[k]) {
								c.Results = append(c.Results, tags[k])
								break
							}
						}
					}
				}
			}
		} // remove after debugging
	}
	if len(c.Results) != 1 {
		t.Errorf("No Element in intersection :(")
	} else {
		fmt.Println("Intersection: ", c.Results)
	}
}
