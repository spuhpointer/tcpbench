package main

import (
	"exutil"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	u "ubftab"

	atmi "github.com/endurox-dev/endurox-go"
)

/*
#include <signal.h>
*/
import "C"

const (
	ProgSection = "netcl"
)

var MRunConf string = "" //Run config

var Mmutex sync.Mutex

//Run the listener
func apprun(ac *atmi.ATMICtx, connr int, size int, sleep int, loops int) {

	var donetps int = 0
	var donetot int = 0
	var i int
	var t exutil.StopWatch
	//Do some work here
	for i = 0; i < loops; i++ {

		buf, err := ac.NewUBF(1024)

		if err != nil {
			log.Fatal(err)
			os.Exit(atmi.FAIL)
		}

		//Set sleep time
		//Alter the size with returns:
		size_out := size*100 + sleep

		if err := buf.BChg(u.EX_NETDATA, 0, size_out); nil != err {
			//return errors.New(err.Error())
			//fmt.
			log.Fatal(err)
			os.Exit(atmi.FAIL)
		}

		//Set target connection id

		if err := buf.BChg(u.EX_NETCONNID, 0, connr); nil != err {
			//return errors.New(err.Error())
			log.Fatal(err)
			os.Exit(atmi.FAIL)
		}

		//Call the server
		if _, err := ac.TpCall("TCPGATE", buf, 0); nil != err {

			buf.TpLogPrintUBF(atmi.LOG_ERROR, "Got error response")

			log.Fatal(err)
			os.Exit(atmi.FAIL)
		}

		//Set reply size
		retData, err := buf.BGetByteArr(u.EX_NETDATA, 0)

		if nil != err {
			log.Fatal(err)
			os.Exit(atmi.FAIL)
		}

		if len(retData) != size {
			ac.TpLogError("Invalid size received got %d expected %d",
				size, len(retData))
			log.Fatal(err)
			os.Exit(atmi.FAIL)
		}

		donetps++
		donetot++
		if t.GetDetlaSec() > 1 {
			var tps float64 = float64(donetps) / float64(t.GetDeltaMillis()) * 1000.0

			var szSent float64 = float64(size) * float64(donetps)
			Mmutex.Lock()
			fmt.Printf("conn: %d TPS: %.2f size: %.2f (%.2f MB) loops: %d done: %d\n",
				connr, tps, szSent, szSent/1024.0/1024.0, loops, donetot)
			Mmutex.Unlock()
			donetps = 0
			t.Reset()
		}

		//Print response
		//buf.TpLogPrintUBF(atmi.LOG_DEBUG, "Got response")
	}

	return
}

//Init function
//@param ac	ATMI context
//@return error (if erro) or nil
func appinit(ac *atmi.ATMICtx) error {

	strPtr := flag.String("run", "70:2000:0:1000000,30:5000000:0:1000000", "Message size")
	flag.Parse()

	if nil == strPtr || "" == *strPtr {
		fmt.Fprintf(os.Stderr, "usage: %s -run=<thcount1>:<size1>:<sleep1>:<loops1>,...,<thcountN>:<sizeN>:<sleepN>:<loopsN>,\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	MRunConf = *strPtr

	return nil
}

//Un-init & Terminate the application
//@param ac	ATMI Context
//@param restCode	Return code. atmi.FAIL (-1) or atmi.SUCCEED(0)
func unInit(ac *atmi.ATMICtx, retCode int) {

	ac.TpTerm()
	ac.FreeATMICtx()
	os.Exit(retCode)
}

//Cliet process main entry
func main() {

	ac, errA := atmi.NewATMICtx()

	if nil != errA {
		fmt.Fprintf(os.Stderr, "Failed to allocate cotnext %d:%s!\n",
			errA.Code(), errA.Message())
		os.Exit(atmi.FAIL)
	}

	if err := appinit(ac); nil != err {
		ac.TpLogError("Failed to init: %s", err)
		os.Exit(atmi.FAIL)
	}

	ac.TpLogWarn("Init complete, starting up...")

	//run_blocks := strings.Fields(",")

	run_blocks := strings.Split(MRunConf, `,`)
	con := 1

	for _, block := range run_blocks {

		fields := strings.Split(block, `:`)

		thcount, _ := strconv.Atoi(fields[0])
		size, _ := strconv.Atoi(fields[1])
		sleep, _ := strconv.Atoi(fields[2])
		loops, _ := strconv.Atoi(fields[3])

		for j := 0; j < thcount; j++ {

			ac2, errA := atmi.NewATMICtx()

			if nil != errA {
				fmt.Fprintf(os.Stderr, "Failed to allocate cotnext %d:%s!\n",
					errA.Code(), errA.Message())
				os.Exit(atmi.FAIL)
			}

			go apprun(ac2, con, size, sleep, loops)
			con++
		}
	}

	//wait

	time.Sleep(9999 * time.Second)

	unInit(ac, atmi.SUCCEED)
}
