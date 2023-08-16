package pir

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"

	"math/rand"

	"contact-discovery/psetggm"
)

const (
	// SecParam is the security parameter in bits.
	SecParam = 128
)

const Left int = 0
const Right int = 1

// One database row.
type Row []byte

type ReconstructFunc func(resp []interface{}) (Row, error)

type PuncClient struct {
	nRows   int
	RowLen  int
	setSize int

	origSetGen, setGen SetGenerator

	randSource       *rand.Rand
	sets             []SetKey
	hints            []Row
	IdxToSetIdx      []int32 //original mapping
	occupancyIndices []bool
}

type PuncHintReq struct {
	//RandSeed           PRGKey
	NumHintsMultiplier int
}

type PuncHintResp struct {
	NRows     int
	RowLen    int
	SetSize   int
	RandInit  int
	SetGenKey PRGKey
	Hints     []Row
}

func xorRowsFlatSlice(db *StaticDB, out []byte, indices Set) {
	for i := range indices {
		indices[i] *= db.RowLen
	}
	psetggm.XorBlocks(db.FlatDb, indices, out)
}

func NewPuncHintReq() *PuncHintReq {
	req := &PuncHintReq{
		NumHintsMultiplier: int(float64(SecParam) * math.Log(2)),
	}
	return req
}

func (req *PuncHintReq) Process(db StaticDB) (*PuncHintResp, error) {
	setSize := int(math.Round(math.Pow(float64(db.NumRows), 0.5)))
	nHints := req.NumHintsMultiplier * db.NumRows / setSize

	hints := make([]Row, nHints)
	hintBuf := make([]byte, db.RowLen*nHints)
	randSource := RandSource()
	var randSeed PRGKey
	_, err := io.ReadFull(randSource, randSeed[:])
	if err != nil {
		log.Fatalf("Failed to initialize random seed: %s", err)
	}
	setGen := NewSetGenerator(randSeed, 0, db.NumRows, setSize)
	var pset PuncturableSet
	for i := 0; i < nHints; i++ {
		setGen.Gen(&pset)
		hints[i] = Row(hintBuf[db.RowLen*i : db.RowLen*(i+1)])
		xorRowsFlatSlice(&db, hints[i], pset.elems)
	}

	return &PuncHintResp{
		Hints:     hints,
		NRows:     db.NumRows,
		RowLen:    db.RowLen,
		SetSize:   setSize,
		SetGenKey: randSeed,
		RandInit:  0,
	}, nil
}

func dbElem(db StaticDB, i int) Row {
	if i < db.NumRows {
		return db.Row(i)
	} else {
		return make(Row, db.RowLen)
	}
}

func (resp *PuncHintResp) NumRows() int {
	return resp.NRows
}

// improved performance
func (resp *PuncHintResp) InitClient(source *rand.Rand, percent float64) *PuncClient {
	c := PuncClient{
		randSource: source,
		nRows:      resp.NRows,
		RowLen:     resp.RowLen,
		setSize:    resp.SetSize,
		hints:      resp.Hints,
		origSetGen: NewSetGenerator(resp.SetGenKey, 0, resp.NRows, resp.SetSize),
	}
	c.initSets(percent)
	return &c
}

func (c *PuncClient) initSets(threshold float64) {
	c.sets = make([]SetKey, len(c.hints))
	c.IdxToSetIdx = make([]int32, c.nRows)
	c.occupancyIndices = make([]bool, c.nRows)
	var pset PuncturableSet
	mappedIndex := 0
	//counter := 0
	for i := 0; i < len(c.hints); i++ {
		c.origSetGen.Gen_NoShift(&pset)
		c.sets[i] = pset.SetKey

		if float64(mappedIndex)/float64(len(c.IdxToSetIdx)) < threshold {
			for _, j := range pset.elems {
				shiftedV := int((uint32(j) + pset.shift) % uint32(pset.univSize))
				if !c.occupancyIndices[shiftedV] {
					c.IdxToSetIdx[shiftedV] = int32(i)
					mappedIndex++
					c.occupancyIndices[shiftedV] = true
				}
			}
			//counter++
		}
	}
	// fmt.Printf("%d/%d Slots in Mapping are filled with threshold of %f \n", mappedIndex, len(c.IdxToSetIdx), threshold)
	//log.Printf("%d sets were considered to fill %f of mapping table", counter, threshold)
	// Use a separate set generator with a new key for all future sets
	// since they must look random to the left server.
	var newSetGenKey PRGKey
	io.ReadFull(c.randSource, newSetGenKey[:])
	c.setGen = NewSetGenerator(newSetGenKey, c.origSetGen.num, c.nRows, c.setSize)
}

// Sample a biased coin that comes up heads (true) with
// probability (nHeads/total).
func (c *PuncClient) bernoulli(nHeads int, total int) bool {
	coin := c.randSource.Intn(total)
	return coin < nHeads
}

func (c *PuncClient) sample(odd1 int, odd2 int, total int) int {
	coin := c.randSource.Intn(total)
	if coin < odd1 {
		return 1
	} else if coin < odd1+odd2 {
		return 2
	} else {
		return 0
	}
}

// Improved performance
func (c *PuncClient) findIndex(idx int) (setIdx int) {
	if idx >= c.nRows {
		return -1
	}
	idx2 := MathMod(idx, c.nRows)
	if c.occupancyIndices[idx2] {
		return int(c.IdxToSetIdx[idx2])
	}
	var pset PuncturableSet
	//go through sets in reverse to avoid invalidating entries in IdxToSetIdx, which gets populated starting from 0
	for j := len(c.sets) - 1; j >= 0; j-- {
		setGen := c.setGenForSet(j)
		setKeyNoShift := c.sets[j]
		shift := setKeyNoShift.shift
		setKeyNoShift.shift = 0
		setGen.EvalInPlace(setKeyNoShift, &pset)

		for _, v := range pset.elems {
			shiftedV := int((uint32(v) + shift) % uint32(setGen.univSize))
			if shiftedV == idx {
				return j
			}
		}
	}
	return -1
}

type PuncQueryReq struct {
	PuncturedSet PuncturedSet
	ExtraElem    int
}

type PuncQueryResp struct {
	Answer    Row
	ExtraElem Row
}

type puncQueryCtx struct {
	randCase int
	setIdx   int
}

func (c *PuncClient) Query(i int) (*[]PuncQueryReq, ReconstructFunc) {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}

	var ctx puncQueryCtx

	if ctx.setIdx = c.findIndex(i); ctx.setIdx < 0 {
		return nil, nil
	}
	i = MathMod(i, c.nRows)

	pset := c.eval(ctx.setIdx)

	var puncSetL, puncSetR PuncturedSet
	var extraL, extraR int
	ctx.randCase = c.sample(c.setSize-1, c.setSize-1, c.nRows)

	switch ctx.randCase {
	case 0:
		newSet := c.setGen.GenWith(i)
		extraL = c.randomMemberExcept(newSet, i)
		extraR = c.randomMemberExcept(pset, i)
		puncSetL = c.setGen.Punc(newSet, i)
		puncSetR = c.setGen.Punc(pset, i)
		if ctx.setIdx >= 0 {
			c.replaceSet(ctx.setIdx, newSet)
		}
	case 1:
		newSet := c.setGen.GenWith(i)
		extraR = c.randomMemberExcept(newSet, i)
		extraL = c.randomMemberExcept(newSet, extraR)
		puncSetL = c.setGen.Punc(newSet, extraR)
		puncSetR = c.setGen.Punc(newSet, i)
	case 2:
		newSet := c.setGen.GenWith(i)
		extraL = c.randomMemberExcept(newSet, i)
		extraR = c.randomMemberExcept(newSet, extraL)
		puncSetL = c.setGen.Punc(newSet, i)
		puncSetR = c.setGen.Punc(newSet, extraL)
	}

	return &[]PuncQueryReq{
			{PuncturedSet: puncSetL, ExtraElem: extraL},
			{PuncturedSet: puncSetR, ExtraElem: extraR},
		},
		func(resps []interface{}) (Row, error) {
			queryResps := make([]*PuncQueryResp, len(resps))
			var ok bool
			for i, r := range resps {
				if queryResps[i], ok = r.(*PuncQueryResp); !ok {
					return nil, fmt.Errorf("Invalid response type: %T, expected: *PuncQueryResp", r)
				}
			}

			return c.reconstruct(ctx, queryResps)
		}
}

func (c *PuncClient) eval(setIdx int) PuncturableSet {
	if c.sets[setIdx].id < c.origSetGen.num {
		return c.origSetGen.Eval(c.sets[setIdx])
	} else {
		return c.setGen.Eval(c.sets[setIdx])
	}
}

func (c *PuncClient) setGenForSet(setIdx int) *SetGenerator {
	if c.sets[setIdx].id < c.origSetGen.num {
		return &c.origSetGen
	} else {
		return &c.setGen
	}
}

func (c *PuncClient) replaceSet(setIdx int, newSet PuncturableSet) {
	pset := c.eval(setIdx)
	for _, idx := range pset.elems {
		if idx < c.nRows && c.occupancyIndices[idx] && c.IdxToSetIdx[idx] == int32(setIdx) {
			c.occupancyIndices[idx] = false
		}
	}
	c.sets[setIdx] = newSet.SetKey
	for _, v := range newSet.elems {
		if !c.occupancyIndices[v] {
			c.IdxToSetIdx[v] = int32(setIdx)
			c.occupancyIndices[v] = true
		}
	}
}

func (c *PuncClient) DummyQuery() *[]PuncQueryReq {
	newSet := c.setGen.GenWith(0)
	extra := c.randomMemberExcept(newSet, 0)
	puncSet := c.setGen.Punc(newSet, 0)
	q := PuncQueryReq{PuncturedSet: puncSet, ExtraElem: extra}
	return &[]PuncQueryReq{q, q}
}

func (q *PuncQueryReq) Process(db StaticDB) (interface{}, error) {
	resp := PuncQueryResp{Answer: make(Row, db.RowLen)}
	psetggm.FastAnswer(q.PuncturedSet.Keys, q.PuncturedSet.Hole, q.PuncturedSet.UnivSize, q.PuncturedSet.SetSize, int(q.PuncturedSet.Shift),
		db.FlatDb, db.RowLen, resp.Answer)
	resp.ExtraElem = db.FlatDb[db.RowLen*q.ExtraElem : db.RowLen*q.ExtraElem+db.RowLen]

	return &resp, nil
}

func (c *PuncClient) reconstruct(ctx puncQueryCtx, resp []*PuncQueryResp) (Row, error) {
	if len(resp) != 2 {
		return nil, fmt.Errorf("Unexpected number of answers: have: %d, want: 2", len(resp))
	}

	out := make(Row, len(c.hints[0]))
	if ctx.setIdx < 0 {
		return nil, errors.New("couldn't find element in collection")
	}

	switch ctx.randCase {
	case 0:
		hint := c.hints[ctx.setIdx]
		xorInto(out, hint)
		xorInto(out, resp[Right].Answer)
		// Update hint with refresh info
		xorInto(hint, hint)
		xorInto(hint, resp[Left].Answer)
		xorInto(hint, out)
	case 1:
		xorInto(out, out)
		xorInto(out, resp[Left].Answer)
		xorInto(out, resp[Right].Answer)
		xorInto(out, resp[Right].ExtraElem)
	case 2:
		xorInto(out, out)
		xorInto(out, resp[Left].Answer)
		xorInto(out, resp[Right].Answer)
		xorInto(out, resp[Left].ExtraElem)
	}
	return out, nil
}

func (c *PuncClient) NumCovered() int {
	covered := make(map[int]bool)
	for j := range c.sets {
		for _, elem := range c.eval(j).elems {
			covered[elem] = true
		}
	}
	return len(covered)
}

// Sample a random element of the set that is not equal to `idx`.
func (c *PuncClient) randomMemberExcept(set PuncturableSet, idx int) int {
	for {
		// TODO: If this is slow, use a more clever way to
		// pick the random element.
		//
		// Use rejection sampling.
		val := set.elems[c.randSource.Intn(c.setSize)]
		if val != idx {
			return val
		}
	}
}

func (c *PuncClient) StateSize() (bitsPerKey, fixedBytes int) {
	return int(math.Log2(float64(len(c.hints)))), len(c.hints) * c.RowLen
}
