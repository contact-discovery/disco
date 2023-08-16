package main

import (
	"contact-discovery/oprf_c"
	"contact-discovery/util"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

var (
	percent   = flag.Float64("percent", 1.0, "Percent of reinsertions")
	cffile    = flag.String("cf", "", "path to cf file")
	rounds    = flag.Int("rounds", 500, "number of reinsertion round")
	prf       = flag.String("prf", "ECNR", "ECNR|GCAES|GCLOWMC")
	filename  = flag.String("out", "cf_sim.csv", "path to result csv")
	dbSizeExp = flag.Int("dbsize", 0, "Number of items in CF (if different to generated CF size)")
)

func main() {
	flag.Parse()

	if *cffile == "" {
		log.Fatalln("No cf file provided")
	}
	// Read CF
	cf := util.CFfromFile(*cffile)
	fmt.Println(cf.Info())

	numElements := int(float64(cf.Size()) / 100.0 * (*percent))
	if *dbSizeExp != 0 {
		numElements = int(float64(int(1)<<*dbSizeExp) / 100.0 * (*percent))
	}

	fmt.Println("cf.Size(): ", cf.Size(), "\tnumElements: ", numElements)

	var lfSlice []float64

	f, err := os.Create(*filename)
	defer f.Close()
	if err != nil {

		log.Fatalln("failed to open file", err)
	}
	w := csv.NewWriter(f)
	done := false
	for round := 0; round < *rounds; round++ {
		if !done {
			var round_results []int
			var elements *[][]byte
			elements = oprf_c.PRF(numElements, *prf, false, 10)
			lfSlice = append(lfSlice, float64(cf.LoadFactor()))
			for _, element := range *elements {
				res, numVictims := cf.Add_sim(element)

				if res {
					round_results = append(round_results, int(numVictims))
				} else {
					done = true
					break
				}
			}
			var victimCount []int
			for _, reinsertions := range round_results {
				for {
					if int(reinsertions) >= len(victimCount) {
						victimCount = append(victimCount, 0)
					} else {
						break
					}
				}
				victimCount[reinsertions]++
			}
			striiings := make([]string, len(victimCount))
			for i, val := range victimCount {
				striiings[i] = strconv.Itoa(val)
			}

			if err := w.Write(striiings); err != nil {
				log.Fatalln("error writing record to file", err)
			}
			w.Flush()
		} else {
			break
		}
	}

}
