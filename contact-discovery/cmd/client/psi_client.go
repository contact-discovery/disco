package main

import (
	psi "contact-discovery/psi"
	"contact-discovery/util"
	"flag"
	"log"
	"math"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"

	"fmt"
	"time"
)

var (
	oprfaddr        = flag.String("oprfaddr", "127.0.0.1:50051", "<HOSTNAME>:<PORT> of OPRF server")
	offaddr         = flag.String("offaddr", "127.0.0.1:50052", "<HOSTNAME>:<PORT> of PIR Offline server")
	onaddr          = flag.String("onaddr", "127.0.0.1:50053", "<HOSTNAME>:<PORT> of PIR Online server")
	clientExp       = flag.Int("exp", 10, "Num items in client set (2^x)")
	segExp          = flag.Int("segexp", 10, "Size of CF paritions, needs to fit into dbSize, (2^x)")
	dbExp           = flag.Int("dbexp", 14, "Num items in server set (2^x)")
	prfType         = flag.String("prf", "ECNR", "PRF types: ECNR|GCAES|GCLOWMC")
	threshold       = flag.Float64("tmap", 0.99, "Max. items in mapping table")
	cpuprofile      = flag.String("cpuprofile", "", "Write cpu profile to file")
	numWorker       = flag.Int("worker", 4, "Num concurrent worker threads")
	testDistributed = flag.Bool("test", false, "Test: Dedicated Server/Segment")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	ncf := int(1) << int(math.Ceil(math.Log2(float64((int(1)<<*dbExp)/util.TagsPerBucket))))

	var clientOfflineTime time.Duration
	start := time.Now()

	c := psi.RunOfflineAll(*offaddr,
		int(1)<<*segExp, // Parition Size
		int(math.Ceil(float64(ncf)/float64(int(1)<<*segExp))), // Number of Partitions
		ncf,
		*threshold,
		*numWorker)
	clientOfflineTime += time.Since(start)

	var clientOnlineOPRFTime time.Duration
	start = time.Now()
	c.RunOPRF(uint16(*clientExp), true, *oprfaddr, *prfType)
	clientOnlineOPRFTime += time.Since(start)

	var clientOnlinePIRTime time.Duration
	start = time.Now()
	c.RunOnline_PIR(uint16(*clientExp), *oprfaddr, *offaddr, *onaddr, *prfType, *numWorker, *testDistributed)
	clientOnlinePIRTime += time.Since(start)

	var out string

	if len(c.Results) == 0 {
		out += fmt.Sprintf("PSI\tNo items in intersection\n")
	} else if len(c.Results) == 1 {
		out += fmt.Sprintf("PSI\tIntersection includes 1 item, with tag: %d\n", c.Results[0])
	} else if len(c.Results) > 1 {
		out += fmt.Sprintf("PSI\tIntersection includes %d items, with tags: \n", len(c.Results))
		for _, res := range c.Results {
			out += fmt.Sprintf("%d\n", res)
		}
	}
	log.Println(out)

	fmt.Println("Comm_Off_CS,Comm_Off_SC,Comm_On_CS,Comm_On_SC,Time_Off,Time_On_OPRF,Time_On_PIR")
	fmt.Println(c.Communication[0], ",",
		c.Communication[1], ",",
		c.Communication[2], ",",
		c.Communication[3], ",",
		clientOfflineTime, ",",
		clientOnlineOPRFTime, ",",
		clientOnlinePIRTime)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile + "_heap")
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
	}
}
