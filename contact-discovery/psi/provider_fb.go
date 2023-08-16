package pir

import (
	fbs "contact-discovery/fbs"
	"contact-discovery/pir"
	"crypto/aes"

	flatbuffers "github.com/google/flatbuffers/go"
)

func DecodeHintReq(hr_b *fbs.HintReq) *pir.PuncHintReq {
	var hintReq pir.PuncHintReq
	// if hr_b.RandSeedLength() != aes.BlockSize {
	// 	log.Fatalln("Rand Seed could not be read correcly")
	// }
	// var prgKey pir.PRGKey
	// for i := 0; i < aes.BlockSize; i++ {
	// 	prgKey[i] = byte(hr_b.RandSeed(i))
	// }
	// hintReq.RandSeed = prgKey
	hintReq.NumHintsMultiplier = int(hr_b.NumHintsMultiplier())
	return &hintReq
}

func DecodeSegmentQueries(queries *fbs.SegmentQueries) (uint32, *[]pir.PuncQueryReq) {
	queryReqs := make([]pir.PuncQueryReq, queries.QueriesLength())
	univSize := int(queries.UnivSize())
	setSize := queries.SetSize()

	for q := 0; q < queries.QueriesLength(); q++ {
		query := new(fbs.Query)
		if queries.Queries(query, q) {
			queryReqs[q].ExtraElem = int(query.ExtraElement())
			queryReqs[q].PuncturedSet.UnivSize = univSize
			queryReqs[q].PuncturedSet.SetSize = int(setSize)

			keys := make([]byte, query.KeysLength())
			for i := 0; i < query.KeysLength(); i++ {
				keys[i] = byte(query.Keys(i))
			}
			queryReqs[q].PuncturedSet.Keys = keys
			queryReqs[q].PuncturedSet.Hole = int(query.Hole())
			queryReqs[q].PuncturedSet.Shift = query.Shift()
		}
	}
	return queries.SegNum(), &queryReqs
}

func DecodeQuery(query *fbs.Query) (uint32, *pir.PuncQueryReq) {
	var queryReq pir.PuncQueryReq
	queryReq.ExtraElem = int(query.ExtraElement())
	queryReq.PuncturedSet.UnivSize = int(query.UnivSize())
	queryReq.PuncturedSet.SetSize = int(query.SetSize())

	keys := make([]byte, query.KeysLength())
	for i := 0; i < query.KeysLength(); i++ {
		keys[i] = byte(query.Keys(i))
	}
	queryReq.PuncturedSet.Keys = keys
	queryReq.PuncturedSet.Hole = int(query.Hole())
	queryReq.PuncturedSet.Shift = query.Shift()
	return query.SegNum(), &queryReq
}

func EncodeSegmentAnswers(pqresps *[]interface{}) *flatbuffers.Builder {
	builder := flatbuffers.NewBuilder(1024 * 1024 * 2)

	fbs_answers := make([]flatbuffers.UOffsetT, len(*pqresps))
	for a := len(*pqresps) - 1; a >= 0; a-- {
		pqr := (*pqresps)[a].(*pir.PuncQueryResp)
		rowLen := len(pqr.Answer)
		fbs.AnswerStartAnswerVector(builder, rowLen)
		for i := rowLen - 1; i >= 0; i-- {
			builder.PrependByte((pqr.Answer)[i])
		}
		row := builder.EndVector(rowLen)

		fbs.AnswerStartExtraElementVector(builder, rowLen)
		for i := rowLen - 1; i >= 0; i-- {
			builder.PrependByte((pqr.ExtraElem)[i])
		}
		ex := builder.EndVector(rowLen)

		fbs.AnswerStart(builder)
		fbs.AnswerAddAnswer(builder, row)
		fbs.AnswerAddExtraElement(builder, ex)

		fbs_answers[a] = fbs.AnswerEnd(builder)
	}

	fbs.SegmentAnswersStartAnswersVector(builder, len(*pqresps))
	for i := len(*pqresps) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(fbs_answers[i])
	}
	seq_answers := builder.EndVector(len(*pqresps))
	fbs.SegmentAnswersStart(builder)
	fbs.SegmentAnswersAddAnswers(builder, seq_answers)
	answers := fbs.SegmentAnswersEnd(builder)
	builder.Finish(answers)
	return builder
}

func EncodeAnswer(pqreq *pir.PuncQueryResp) *flatbuffers.Builder {
	builder := flatbuffers.NewBuilder(1024 * 2)
	rowLen := len(pqreq.Answer)
	fbs.AnswerStartAnswerVector(builder, rowLen)
	for i := rowLen - 1; i >= 0; i-- {
		builder.PrependByte((pqreq.Answer)[i])
	}
	row := builder.EndVector(rowLen)

	fbs.AnswerStartExtraElementVector(builder, rowLen)
	for i := rowLen - 1; i >= 0; i-- {
		builder.PrependByte((pqreq.ExtraElem)[i])
	}
	ex := builder.EndVector(rowLen)

	fbs.AnswerStart(builder)
	fbs.AnswerAddAnswer(builder, row)
	fbs.AnswerAddExtraElement(builder, ex)

	ans := fbs.AnswerEnd(builder)
	builder.Finish(ans)
	return builder
}

func EncodeAllHintResp(all_phr *[]pir.PuncHintResp) *flatbuffers.Builder {
	builder := flatbuffers.NewBuilder(1024 * 1024 * 650)
	all_enc_hr := make([]flatbuffers.UOffsetT, len(*all_phr))

	for s := len(*all_phr) - 1; s >= 0; s-- {
		phr := (*all_phr)[s]

		fbs.HintRespStartHintsVector(builder, len(phr.Hints)*phr.RowLen)
		for i := len(phr.Hints) - 1; i >= 0; i-- {
			for j := phr.RowLen - 1; j >= 0; j-- {
				builder.PlaceByte(phr.Hints[i][j])
			}
		}
		hints := builder.EndVector(len(phr.Hints) * phr.RowLen)
		fbs.HintRespStartSetGenKeyVector(builder, aes.BlockSize)
		for i := aes.BlockSize - 1; i >= 0; i-- {
			builder.PlaceByte(phr.SetGenKey[i])
		}
		sgk := builder.EndVector(aes.BlockSize)

		fbs.HintRespStart(builder)
		fbs.HintRespAddSetGenKey(builder, sgk)
		fbs.HintRespAddHints(builder, hints)
		fbs.HintRespAddRandInit(builder, int32(phr.RandInit))
		hintResp := fbs.HintRespEnd(builder)
		all_enc_hr[s] = hintResp
	}

	fbs.AllHintRespStartHintRespsVector(builder, len(*all_phr))
	for i := len(*all_phr) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(all_enc_hr[i])
	}
	all_enc_hr_vec := builder.EndVector(len(*all_phr))
	fbs.AllHintRespStart(builder)
	fbs.AllHintRespAddNRows(builder, uint32((*all_phr)[0].NRows))
	fbs.AllHintRespAddRowLen(builder, int32((*all_phr)[0].RowLen))
	fbs.AllHintRespAddSetSize(builder, int32((*all_phr)[0].SetSize))
	fbs.AllHintRespAddHintResps(builder, all_enc_hr_vec)
	all_hintsResp := fbs.AllHintRespEnd(builder)
	builder.Finish(all_hintsResp)
	return builder
}

func FinishEncode(builder *flatbuffers.Builder) (buf []byte) {
	return builder.FinishedBytes()
}
