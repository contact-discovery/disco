package pir

import (
	fbs "contact-discovery/fbs"
	"contact-discovery/oprf_c"
	"contact-discovery/pir"
	"contact-discovery/util"
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var maxMsgSize = 2147483648

// max iteration for query scheduling based on simple hashing analysis
var MaxQueryIter = [11][3]int{
	{2, 2048, 32768}, //for 2^0 = 0 segments for {1, 2^10, 2^14}
	{2, 1186, 17031}, //2^1
	{2, 659, 8765},   //2^2
	{2, 373, 4542},   //2^3
	{2, 218, 2382},   //2^4
	{2, 132, 1270},   //2^5
	{2, 84, 693},     //2^6
	{2, 56, 389},     //2^7
	{2, 39, 226},     //2^8
	{2, 29, 137},     //2^9
	{2, 22, 87},      //2^10
}

type Client struct {
	NumSegs       uint32 // = len(ClientPIR)
	SegSize       uint32
	N_cf          uint32 // numBuckets, required for CF hashing
	MaxIterations int    // Number of query rounds that need to be send
	Communication []int  // size 2: send and receive communciation
	ServerAddr    [3]string
	Results       []uint32
	ClientPIR     []pir.PuncClient
	CfIndices     *[][3]uint
}

func NewClient(segsize int, numsegs int, npir int) (c Client, hintreq *pir.PuncHintReq) {
	c.SegSize = uint32(segsize)
	c.NumSegs = uint32(numsegs)
	c.N_cf = uint32(npir)
	c.Communication = make([]int, 4)
	return c, pir.NewPuncHintReq()
}

/////////////////////// Offline

func (c *Client) RequestHintPIR() *pir.PuncHintReq {
	return pir.NewPuncHintReq()
}

func offlineWorkerC(c *Client, id int, jobs <-chan int, wg *sync.WaitGroup, hintResps *[]pir.PuncHintResp, threshold float64) {
	for j := range jobs {
		defer wg.Done()
		// Init RandSource with value from hintresp
		c.ClientPIR[j] = *(*hintResps)[j].InitClient(pir.RandSource(), threshold)
	}
}

func (c *Client) ProcessHintResp(hintResps *[]pir.PuncHintResp, threshold float64, numWorker int) {
	jobs := make(chan int, c.NumSegs)
	c.ClientPIR = make([]pir.PuncClient, c.NumSegs)
	var wg sync.WaitGroup
	wg.Add(int(c.NumSegs))
	for w := 0; w < numWorker; w++ {
		go offlineWorkerC(c, w, jobs, &wg, hintResps, threshold)
	}
	for j := 0; j < int(c.NumSegs); j++ {
		jobs <- j
	}
	close(jobs)
	wg.Wait()
}

func RunOfflineAll(offServer string, segsize int, numsegs int, npir int, threshold float64, numWorker int) *Client {

	// conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", 3000),
	conn, err := grpc.Dial(offServer,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(flatbuffers.FlatbuffersCodec{}),
			grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)))
	if err != nil {
		log.Fatalf("PSI\tfail to dial: %v", err)
	}
	defer conn.Close()
	c_pir := fbs.NewPIRClient(conn)

	// Create PIR Hint Request
	c, hintreq := NewClient(segsize, numsegs, npir)
	b := EncodeHintReq(hintreq)
	c.Communication[0] = len(b.FinishedBytes())
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Minute)
	defer cancel()
	log.Println("PSI\tRequest hints")
	hr, err := (c_pir).OfflineAll(ctx, b, grpc.CallContentSubtype("flatbuffers"))
	if err != nil {
		log.Fatalf("%v.OfflineAll(_) = _, %v: ", c_pir, err)
	}
	b = nil
	log.Println("PSI\tReceived hints")
	c.Communication[1] = len(hr.Table().Bytes)
	fmt.Println(len(hr.Table().Bytes))
	c.ProcessHintResp(DecodeAllHintResp(hr, numWorker), threshold, numWorker)
	log.Printf("PSI\tProcessed Hints for %d partitions", (len(c.ClientPIR)))
	return &c
}

/////////////////////// Online

func (c *Client) RunOPRF(exp uint16, first bool, oprf string, prfType string) {

	if exp == 0 {
		c.MaxIterations = MaxQueryIter[int(math.Log2(float64(c.NumSegs)))][0]
	} else if exp == 10 {
		c.MaxIterations = MaxQueryIter[int(math.Log2(float64(c.NumSegs)))][1]
	} else if exp == 14 {
		c.MaxIterations = MaxQueryIter[int(math.Log2(float64(c.NumSegs)))][2]
	} else {
		log.Fatalln("PSI\tClient set has to have size 2^0, 2^10 or 2^14")
	}

	oprf_addr := strings.Split(oprf, ":")
	oprf_port, err := strconv.Atoi(oprf_addr[1])
	if err != nil {
		log.Fatalln("PSI\tError getting oprf server port")
	}
	c.CfIndices = oprf_c.OPRF_CFValues(int(1)<<int(exp), prfType, first, oprf_addr[0], oprf_port, c.N_cf)
}

func onlineWorkerC(c *Client, id int, jobs <-chan int, wg_segs *sync.WaitGroup, test bool) {
	for s := range jobs {
		defer wg_segs.Done()
		/***** QueryScheduling *****/
		queryReqs := make([][]pir.PuncQueryReq, c.MaxIterations)
		tags := make([]uint32, c.MaxIterations)
		recons := make([]pir.ReconstructFunc, c.MaxIterations)

		i := 0
		for _, indices := range *c.CfIndices {
			for _, idx := range indices[1:] { // indices[0] contains tag
				if int(idx/uint(c.SegSize)) == s {
					tags[i] = uint32(indices[0])
					queryReq, recon := c.ClientPIR[s].Query(int(idx % uint(c.SegSize)))
					queryReqs[i], recons[i] = *queryReq, recon
					if queryReqs[i] == nil {
						log.Fatalln("PSI\tCould not find index!")
					}
					i++
				}
			}
		}
		for ; i < c.MaxIterations; i++ {
			queryReqs[i] = *(c.ClientPIR[s].DummyQuery())
		}

		/***** Query Segment *****/
		conn1, err := grpc.Dial(c.ServerAddr[0],
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.ForceCodec(flatbuffers.FlatbuffersCodec{}),
				grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)))
		if err != nil {
			log.Fatalf("PSI\tfail to dial: %v", err)
		}
		defer conn1.Close()

		conn2, err := grpc.Dial(c.ServerAddr[1],

			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.ForceCodec(flatbuffers.FlatbuffersCodec{}),
				grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)))
		if err != nil {
			log.Fatalf("PSI\tfail to dial: %v", err)
		}
		defer conn2.Close()

		c_pirs := [2]fbs.PIRClient{fbs.NewPIRClient(conn1), fbs.NewPIRClient(conn2)}
		answers := make([][]interface{}, c.MaxIterations)

		for c_i, grpc_client := range c_pirs {
			var b *flatbuffers.Builder
			if test {
				b = EncodeSegmentQueries(0, &queryReqs, c_i)
			} else {
				b = EncodeSegmentQueries(s, &queryReqs, c_i)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Minute)
			defer cancel()
			ans_b, err := grpc_client.OnlineSegment(ctx, b, grpc.CallContentSubtype("flatbuffers"))
			if err != nil {
				log.Fatalf("PSI\t%v.OnlineSegment(_) = _, %v: ", grpc_client, err)
			}
			DecodeSegmentAnswer(ans_b, &answers, c_i)

			c.Communication[2] += len((b.FinishedBytes()))
			c.Communication[3] += len(ans_b.Table().Bytes)
		}
		queryReqs = nil

		/***** Process Answers *****/
		for k, ans := range answers {
			if tags[k] != 0 { //Only look at real answers, not dummies
				if len(ans) != 2 {
					log.Fatal("PSI\tDid not receive response from both servers")
				} else {
					row, err := (recons[k])(ans)
					if err != nil {
						log.Fatal("PSI\tCould not reconstruct answer")
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
	}
}

func (c *Client) RunOnline_PIR(exp uint16, oprf string, pir1 string, pir2 string, prfType string, numWorker int, test bool) {

	c.ServerAddr = [3]string{pir1, pir2, oprf}

	jobs := make(chan int, c.NumSegs)
	var wg_segs sync.WaitGroup
	wg_segs.Add(int(c.NumSegs))

	for w := 0; w < numWorker; w++ {
		go onlineWorkerC(c, w, jobs, &wg_segs, test)
	}
	for j := 0; j < int(c.NumSegs); j++ {
		jobs <- j
	}
	close(jobs)
	wg_segs.Wait()
}
