package main

import (
	fbs "contact-discovery/fbs"
	pir "contact-discovery/pir"
	psi "contact-discovery/psi"
	"contact-discovery/util"
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"google.golang.org/grpc"
)

var (
	addr       = flag.String("addr", "localhost:50052", "<HOSTNAME>:<PORT> of this PIR Offline server")
	pprofAddr  = flag.String("pprofaddr", "", "<HOSTNAME>:<PORT> of the pprof http server")
	segSizeExp = flag.Int("segexp", 10, "Size of CF paritions, needs to fit into dbSize, (2^x)")
	cfFile     = flag.String("cf", "", "path to cf file")
	prfType    = flag.String("prf", "ECNR", "PRF types: ECNR|GCAES|GCLOWMC")
	numWorker  = flag.Int("worker", 4, "Num concurrent worker threads")
	test       = flag.Bool("test", false, "Test: Dedicated Server/Segment")
)

var maxMsgSize = 2147483648

type gRPCProvider struct {
	fbs.UnimplementedPIRServer
	PIRProvider *psi.Provider
	SegCounter  uint32
	onlineTime  time.Duration
}

func newGRPCProvider(pirProv *psi.Provider) *gRPCProvider {
	grpcProv := &gRPCProvider{PIRProvider: pirProv}
	return grpcProv
}

func (grpcProv *gRPCProvider) OfflineAll(ctx context.Context, hr *fbs.HintReq) (*flatbuffers.Builder, error) {
	var serverOfflineTime time.Duration
	start := time.Now()
	builder := psi.EncodeAllHintResp(grpcProv.PIRProvider.ClientHints)
	log.Println("PSI\t Offline S->C: ", len(builder.FinishedBytes()))
	log.Println("PSI\tOffline Server Time: ", serverOfflineTime+time.Since(start))
	return builder, nil
}

func (grpcProv *gRPCProvider) Online(ctx context.Context, q_fb *fbs.Query) (*flatbuffers.Builder, error) {
	start := time.Now()
	segNum, q := psi.DecodeQuery(q_fb)
	answer := grpcProv.PIRProvider.OnlinePIR(q, uint32(segNum))
	builder := psi.EncodeAnswer(answer.(*pir.PuncQueryResp))
	log.Println("PSI\tOnline Server Time: ", time.Since(start))
	return builder, nil
}

func (grpcProv *gRPCProvider) OnlineSegment(ctx context.Context, q_fbs *fbs.SegmentQueries) (*flatbuffers.Builder, error) {
	start := time.Now()
	segNum, queries := psi.DecodeSegmentQueries(q_fbs)
	answers := grpcProv.PIRProvider.OnlineSegments(int(segNum), queries)
	builder := psi.EncodeSegmentAnswers(answers)

	grpcProv.SegCounter++
	grpcProv.onlineTime += time.Since(start)
	if *test {
		log.Println("PSI\tOnline Server Time: ", grpcProv.onlineTime)
		grpcProv.onlineTime = 0
		grpcProv.SegCounter = 0
	} else if grpcProv.SegCounter == grpcProv.PIRProvider.NumSegs {
		log.Println("PSI\tOnline Server Time: ", grpcProv.onlineTime)
		grpcProv.onlineTime = 0
		grpcProv.SegCounter = 0
	}
	return builder, nil
}

func runServer(segSize uint32, addr string) {
	// Read CF from file
	cf := util.CFfromFile(*cfFile)
	log.Println(cf.Info())

	start := time.Now()
	p_pir := psi.NewProvider(cf, segSize)
	p_pir.SetupHints(*numWorker, *test)
	lis, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("PSI\tFailed to listen: %v", err)
	} else {
		log.Println("PSI\tServer ready")
	}

	grpcProvider := grpc.NewServer(grpc.ForceServerCodec(flatbuffers.FlatbuffersCodec{}),
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize))
	fbs.RegisterPIRServer(grpcProvider, newGRPCProvider(&p_pir))
	log.Println("PSI\tSetup Server Time: ", time.Since(start))
	if err := grpcProvider.Serve(lis); err != nil {
		log.Fatalf("PSI\tFailed to serve: %v", err)
	}
}

func main() {
	flag.Parse()
	if *cfFile == "" {
		log.Fatalln("PSI\tNo cuckoo filter file provided")
	}

	if *pprofAddr != "" {
		runtime.SetBlockProfileRate(1) // used for pprof http server
		go func() {
			http.ListenAndServe(*pprofAddr, nil)
		}()
	}
	runServer(uint32(int(1)<<*segSizeExp), *addr)
}
