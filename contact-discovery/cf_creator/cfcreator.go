package main

import (
	"contact-discovery/oprf_c"
	"contact-discovery/util"
	"flag"
	"log"
	"time"

	"github.com/linvon/cuckoo-filter"
)

var (
	dbSizeExp   = flag.Int("dbsize", 14, "Num DB Rows (2^x)")
	numItemsExp = flag.Int("numItems", 0, "Num Items to insert to CF")
	cfFile      = flag.String("cf", "cfFile.data", "path to cf file")
	prfType     = flag.String("prf", "GCAES", "PRF types: ECNR|GCAES|GCLOWMC")
	threads     = flag.Int("threads", 1, "Number of threads")
)

func main() {
	flag.Parse()
	dbSize := int(1) << *dbSizeExp
	numItems := dbSize
	var cftime time.Duration
	start := time.Now()
	cf := cuckoo.NewFilter(util.TagsPerBucket, util.BitsPerItem,
		uint(dbSize), cuckoo.TableTypeSingle)
	cftime += time.Since(start)
	if *numItemsExp != 0 {
		numItems = int(1) << *numItemsExp
		oprf_c.CF_PRF(cf, *prfType, numItems, true, *threads)
	} else {
		oprf_c.CF_PRF(cf, *prfType, dbSize, true, *threads)
	}

	log.Println("CF\tEmpty CF Creation: \tTime: ", cftime)
	log.Println("CF\tCF with size ", dbSize, ", and ", numItems, " elements.")

	util.CFToFile(cf, *cfFile)
	log.Println("CF\tCF Written to ", *cfFile)
}
