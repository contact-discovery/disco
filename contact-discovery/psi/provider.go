package pir

import (
	"contact-discovery/pir"
	"contact-discovery/util"
	"log"
	"sync"

	"github.com/linvon/cuckoo-filter"
)

// Requires GCAES-OPRF Server on port 50051
type Provider struct {
	SegSize     uint32
	NumBuckets  uint32
	NumSegs     uint32
	cf          *cuckoo.Filter
	segments    []*pir.StaticDB
	ClientHints *[]pir.PuncHintResp
}

func NewProvider(cf *cuckoo.Filter, segSize uint32) (p Provider) {
	p.SetupDBFromCF(cf, segSize)
	return p
}

func (p *Provider) SetupDBFromCF(cf *cuckoo.Filter, segSize uint32) {
	p.cf = cf
	p.SegSize = segSize
	p.NumBuckets = uint32(p.cf.NumBuckets())
	p.NumSegs = p.NumBuckets / segSize

	if p.NumBuckets%segSize != 0 {
		log.Fatalln("[PSI]\tChoose the segSize s.t. numBuckets mod segSize = 0")
	}
	if p.NumBuckets/segSize == 0 {
		log.Fatalln("[PSI]\tChoose the segSize s.t. numBuckets / segSize > 0")
	}
	p.segments = make([]*pir.StaticDB, p.NumSegs)
	for i := 0; uint32(i) < p.NumSegs; i++ {
		// TODO optimize this if too slow
		p.segments[i] = util.CFSegment(p.cf, uint32(i), segSize)
	}
}

func (p *Provider) SetupHints(numWorker int, test bool) {
	// Server generates hints before receiving request
	// to avoid rewriting old code a dummy request is created
	fake_hr := pir.NewPuncHintReq()
	p.ClientHints = p.AllHintReponse(fake_hr, numWorker, test)
}

func offlineWorkerP(p *Provider, id int, jobs <-chan int, wg *sync.WaitGroup, hintreq pir.PuncHintReq, hints *[]pir.PuncHintResp) {
	for i := range jobs {
		hr, err := hintreq.Process(*p.segments[i])
		(*hints)[i] = *hr
		if err != nil {
			log.Fatal("[PIR]\tOffline hint generation failed")
		}
		wg.Done()
	}
}

func (p *Provider) AllHintReponse(hintreq *pir.PuncHintReq, numWorker int, test bool) *[]pir.PuncHintResp {
	hints := make([]pir.PuncHintResp, p.NumSegs)
	var wg sync.WaitGroup
	var jobs chan int
	var numJobs int
	// for test when each segment on individual server
	if test {
		numJobs = 1
	} else {
		numJobs = int(p.NumSegs)
	}
	wg.Add(int(numJobs))
	jobs = make(chan int, numJobs)

	for w := 0; w < numWorker; w++ {
		go offlineWorkerP(p, w, jobs, &wg, *hintreq, &hints)
	}
	for j := 0; j < int(numJobs); j++ {
		jobs <- j
	}
	close(jobs)
	wg.Wait()
	if test {
		for i := range hints {
			hints[i] = hints[0]
		}
	}
	return &hints
}

func (p *Provider) OnlineSegments(segNum int, queries *[]pir.PuncQueryReq) *[]interface{} {
	if uint32(segNum) >= p.NumSegs {
		log.Fatal("[PIR]\tError selected segment number not valid")
	}
	var err error
	answers := make([]interface{}, len(*queries))

	for i, query := range *queries {
		answers[i], err = query.Process(*p.segments[segNum])
		if err != nil {
			log.Fatal("[PIR]\tError answering query")
		}
	}
	return &answers
}

func (p *Provider) OnlinePIR(query *pir.PuncQueryReq, segNum uint32) interface{} {
	if segNum >= p.NumSegs {
		log.Fatal("[PIR]\tError selected segment number not valid")
	}
	answer, err := query.Process(*p.segments[segNum])
	if err != nil {
		log.Fatal("[PIR]\tError answering query")
	}
	return answer
}

func (p *Provider) CF() (cf *cuckoo.Filter) {
	return p.cf
}

func (p *Provider) Segments() []*pir.StaticDB {
	return p.segments
}
