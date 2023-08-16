package pir

import (
	fbs "contact-discovery/fbs"
	"crypto/aes"
	"log"

	pir "contact-discovery/pir"
	"sync"

	flatbuffers "github.com/google/flatbuffers/go"
)

func EncodeHintReq(phr *pir.PuncHintReq) *flatbuffers.Builder {
	builder := flatbuffers.NewBuilder(60)
	// fbs.HintReqStartRandSeedVector(builder, len(phr.RandSeed))
	// for i := len(phr.RandSeed) - 1; i >= 0; i-- {
	// 	builder.PlaceByte(phr.RandSeed[i])
	// }
	// randSeed := builder.EndVector(len(phr.RandSeed))

	fbs.HintReqStart(builder)
	//fbs.HintReqAddRandSeed(builder, randSeed)
	fbs.HintReqAddNumHintsMultiplier(builder, int16(phr.NumHintsMultiplier))
	hintReq := fbs.HintReqEnd(builder)

	builder.Finish(hintReq)

	return builder
}

func EncodeSegmentQueries(segNum int, queries *[][]pir.PuncQueryReq, server_i int) *flatbuffers.Builder {
	builder := flatbuffers.NewBuilder(1024 * 1024 * 3)

	var univSize int
	var setSize int
	fbs_queries := make([]flatbuffers.UOffsetT, len(*queries))

	for q := len(*queries) - 1; q >= 0; q-- {
		//queryReq := ((*queries)[q][server_i]).(*pir.PuncQueryReq)
		queryReq := (*queries)[q][server_i]

		if univSize == 0 {
			univSize = queryReq.PuncturedSet.UnivSize
			setSize = queryReq.PuncturedSet.SetSize
		}

		kLen := len(queryReq.PuncturedSet.Keys)
		fbs.QueryStartKeysVector(builder, kLen)
		for i := kLen - 1; i >= 0; i-- {
			builder.PlaceByte((queryReq.PuncturedSet.Keys[i]))
		}
		keys := builder.EndVector(len(queryReq.PuncturedSet.Keys))

		fbs.QueryStart(builder)
		fbs.QueryAddKeys(builder, keys)

		fbs.QueryAddHole(builder, int32(queryReq.PuncturedSet.Hole))
		fbs.QueryAddShift(builder, queryReq.PuncturedSet.Shift)
		fbs.QueryAddExtraElement(builder, int32(queryReq.ExtraElem))
		fbs_queries[q] = fbs.QueryEnd(builder)
	}

	fbs.SegmentQueriesStartQueriesVector(builder, len(*queries))
	for i := len(*queries) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(fbs_queries[i])
	}
	seg_queries := builder.EndVector(len(*queries))
	fbs.SegmentQueriesStart(builder)
	fbs.SegmentQueriesAddSegNum(builder, uint32(segNum))
	fbs.SegmentQueriesAddSetSize(builder, int32(setSize))
	fbs.SegmentQueriesAddUnivSize(builder, uint32(univSize))
	fbs.SegmentQueriesAddQueries(builder, seg_queries)

	seg_queries_fb := fbs.SegmentQueriesEnd(builder)

	builder.Finish(seg_queries_fb)
	return builder

}

func EncodeQuery(segNum uint32, q *pir.PuncQueryReq) *flatbuffers.Builder {
	builder := flatbuffers.NewBuilder(1024 * 1024 * 2)

	kLen := len(q.PuncturedSet.Keys)
	fbs.QueryStartKeysVector(builder, kLen)
	for i := kLen - 1; i >= 0; i-- {
		builder.PlaceByte((q.PuncturedSet.Keys[i]))
	}
	keys := builder.EndVector(len(q.PuncturedSet.Keys))

	fbs.QueryStart(builder)
	fbs.QueryAddKeys(builder, keys)

	fbs.QueryAddHole(builder, int32(q.PuncturedSet.Hole))
	fbs.QueryAddShift(builder, q.PuncturedSet.Shift)
	fbs.QueryAddExtraElement(builder, int32(q.ExtraElem))
	query := fbs.QueryEnd(builder)

	builder.Finish(query)
	return builder
}

func decodeHintRespWorker(id int, jobs <-chan int, wg *sync.WaitGroup, all_hr *fbs.AllHintResp, hintResps *[]pir.PuncHintResp) {
	for j := range jobs {
		//fmt.Printf("decodeHintRespWorker %d started job %d\n", id, j)
		defer wg.Done()

		var phr pir.PuncHintResp
		phr.NRows = int(all_hr.NRows())
		phr.RowLen = int(all_hr.RowLen())
		phr.SetSize = int(all_hr.SetSize())

		phr_b := new(fbs.HintResp)
		if all_hr.HintResps(phr_b, j) {
			hints := make([]pir.Row, phr_b.HintsLength()/phr.RowLen)
			// iterate over number of hints (not bytes of all hints)
			for i := 0; i < phr_b.HintsLength()/phr.RowLen; i++ {
				prow := make([]byte, phr.RowLen)
				for j := 0; j < phr.RowLen; j++ {
					prow[j] = byte(phr_b.Hints(i*phr.RowLen + j))
				}
				hints[i] = prow
			}
			phr.Hints = hints

			var prgKey pir.PRGKey
			for i := 0; i < aes.BlockSize; i++ {
				prgKey[i] = byte(phr_b.SetGenKey(i))
			}
			phr.SetGenKey = prgKey
			phr.RandInit = int(phr_b.RandInit())
			(*hintResps)[j] = phr
		}
	}
	//fmt.Printf("decodeHintRespWorker %d finished job %d\n", id, j)
}

func DecodeAllHintResp(all_hr *fbs.AllHintResp, numWorker int) *[]pir.PuncHintResp {
	sh := make([]pir.PuncHintResp, all_hr.HintRespsLength())

	jobs := make(chan int, all_hr.HintRespsLength())
	var wg sync.WaitGroup
	wg.Add(int(all_hr.HintRespsLength()))

	for w := 0; w < numWorker; w++ {
		go decodeHintRespWorker(w, jobs, &wg, all_hr, &sh)
	}
	for j := 0; j < int(all_hr.HintRespsLength()); j++ {
		jobs <- j
	}
	close(jobs)
	wg.Wait()

	return &sh
}

func DecodeSegmentAnswer(pqresps *fbs.SegmentAnswers, answers *[][]interface{}, server_i int) {
	for s := 0; s < pqresps.AnswersLength(); s++ {
		pqr_b := new(fbs.Answer)
		if pqresps.Answers(pqr_b, s) {
			if pqr_b.AnswerLength() != pqr_b.ExtraElementLength() {
				log.Fatalln("Length of answer and ExtraElement buffer don't match")
			}
			var r1, r2 pir.Row
			r1 = make([]byte, pqr_b.AnswerLength())
			r2 = make([]byte, pqr_b.AnswerLength())
			for i := 0; i < pqr_b.AnswerLength(); i++ {
				r1[i] = byte(pqr_b.Answer(i))
				r2[i] = byte(pqr_b.ExtraElement(i))
			}

			(*answers)[s] = append((*answers)[s], &pir.PuncQueryResp{Answer: r1, ExtraElem: r2})
		}
	}
}

func DecodeAnswer(pqr_b *fbs.Answer) *pir.PuncQueryResp {
	var pqr pir.PuncQueryResp
	if pqr_b.AnswerLength() != pqr_b.ExtraElementLength() {
		log.Fatalln("Length of answer and ExtraElement buffer don't match")
	}
	var r1, r2 pir.Row
	r1 = make([]byte, pqr_b.AnswerLength())
	r2 = make([]byte, pqr_b.AnswerLength())
	for i := 0; i < pqr_b.AnswerLength(); i++ {
		r1[i] = byte(pqr_b.Answer(i))
		r2[i] = byte(pqr_b.ExtraElement(i))
	}
	pqr.Answer = r1
	pqr.ExtraElem = r2
	return &pqr
}
