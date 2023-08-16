package client_android

import (
	"fmt"
	"log"
	"math"
	psi "mobile-contact-discovery/psi"
	"mobile-contact-discovery/util"
	"time"
)

var (
	maxMsgSize = 268435456
)

var jc JavaCallback

type JavaCallback interface {
	CallFromGo(string)
}

func RegisterJavaCallback(c JavaCallback) {
	jc = c
	log.Println("PSI\tJava callback registered", jc)
}

func PIRCallback(pir1 string, pir2 string, clientExp int, segExp int, dbExp int, prfType string, threshold float64, numWorker int, test bool) {
	go func() {
		jc.CallFromGo(RunPIR(pir1, pir2, clientExp, segExp, dbExp, prfType, threshold, numWorker, test))
	}()
}

func bitToMiB(num int) float32 {
	return float32(num) / 1024 / 1024
}

func RunPIR(pir1 string, pir2 string, clientExp int, segExp int, dbExp int, prfType string, threshold float64, numWorker int, test bool) string {

	ncf := int(1) << int(math.Ceil(math.Log2(float64((int(1)<<dbExp)/util.TagsPerBucket))))

	var clientOfflineTime time.Duration
	start := time.Now()
	c := psi.RunOfflineAll(pir1,
		int(1)<<segExp, //parition size
		int(math.Ceil(float64(ncf)/float64(int(1)<<segExp))), //number of partitions
		ncf,
		threshold,
		numWorker)
	clientOfflineTime += time.Since(start)

	c.RunOPRF_Android(uint16(clientExp), true, "", prfType)

	var clientOnlineTime time.Duration
	start = time.Now()
	c.RunOnline_PIR(
		uint16(clientExp),
		"", pir1, pir2,
		prfType,
		numWorker,
		test)
	clientOnlineTime += time.Since(start)

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
	out += fmt.Sprintf("Offline Comm [MiB]: C->S: %.4f, S->C: %.4f\n",
		bitToMiB(c.Communication[0]),
		bitToMiB(c.Communication[1]))
	out += fmt.Sprintf("Online Comm [MiB]: C->S: %.4f, S->C: %.4f\n",
		bitToMiB(c.Communication[2]),
		bitToMiB(c.Communication[3]))
	out += fmt.Sprintf("Offline Time:\t%s \n", clientOfflineTime.String())
	out += fmt.Sprintf("Online Time:\t%s \n", clientOnlineTime.String())

	return fmt.Sprintf("%s", out)
}
