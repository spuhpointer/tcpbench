package main

import (
	"fmt"
	"os"
	"time"
	u "ubftab"

	atmi "github.com/endurox-dev/endurox-go"
)

const (
	SUCCEED = atmi.SUCCEED
	FAIL    = atmi.FAIL
)

var M_maxblock [20 * 1024 * 1024]byte

//NETRSP service
//@param ac ATMI Context
//@param svc Service call information
func NETRSP(ac *atmi.ATMICtx, svc *atmi.TPSVCINFO) {

	ret := SUCCEED
	var sz int64 = 0

	//Return to the caller
	defer func() {

		ac.TpLogCloseReqFile()
		if SUCCEED == ret {
			ac.TpReturn(atmi.TPSUCCESS, 0, &svc.Data, 0)
		} else {
			ac.TpReturn(atmi.TPFAIL, 0, &svc.Data, 0)
		}
	}()

	//Get UBF Handler
	ub, _ := ac.CastToUBF(&svc.Data)

	//Print the buffer to stdout
	//fmt.Println("Incoming request:")
	//ub.TpLogPrintUBF(atmi.LOG_DEBUG, "Incoming request:")

	//Resize buffer, to have some more space - say 20 MB (max buffer..)
	if err := ub.TpRealloc(20 * 1024 * 1024); err != nil {
		ac.TpLogError("TpRealloc() Got error: %d:[%s]", err.Code(), err.Message())
		ret = FAIL
		return
	}

	//Get the message size requested..
	sz, err := ub.BGetInt64(u.EX_NETDATA, 0)

	if err != nil {
		ac.TpLogError("BGetInt64() Got error: %d:[%s]", err.Code(), err.Message())
		ret = FAIL
		return
	}
	//malloc some block..

	sleeptime := sz % 100

	// get the size left over
	sz /= 100

	block := M_maxblock[0:sz]

	ac.TpLogInfo("requested buffer size: %d blocksz: %d", sz, len(block))

	err = ub.BChg(u.EX_NETDATA, 0, block)

	if err != nil {
		ac.TpLogError("BChg() Got error: %d:[%s]", err.Code(), err.Message())
		ret = FAIL
		return
	}

	if sleeptime > 0 {
		//time.Sleep(sleeptime * time.Second)
		time.Sleep(time.Duration(sleeptime*1000) * time.Millisecond)
	}

	//Lets send response to connection, no reply
	_, err = ac.TpACall("GATEWAYSERVER", ub, atmi.TPNOREPLY)

	if err != nil {
		ac.TpLogError("TpACall() Got error: %d:[%s]", err.Code(), err.Message())
		ret = FAIL
		return
	}

	return
}

//Server init, called when process is booted
//@param ac ATMI Context
func Init(ac *atmi.ATMICtx) int {

	ac.TpLogWarn("Doing server init...")

	//Advertize service
	if err := ac.TpAdvertise("NETRSP", "NETRSP", NETRSP); err != nil {
		ac.TpLogError("Failed to Advertise: ATMI Error %d:[%s]\n", err.Code(), err.Message())
		return atmi.FAIL
	}

	return SUCCEED
}

//Server shutdown
//@param ac ATMI Context
func Uninit(ac *atmi.ATMICtx) {
	ac.TpLogWarn("Server is shutting down...")
}

//Executable main entry point
func main() {
	//Have some context
	ac, err := atmi.NewATMICtx()

	if nil != err {
		fmt.Fprintf(os.Stderr, "Failed to allocate new context: %s", err)
		os.Exit(atmi.FAIL)
	} else {
		//Run as server
		if err = ac.TpRun(Init, Uninit); nil != err {
			ac.TpLogError("Exit with failure")
			os.Exit(atmi.FAIL)
		} else {
			ac.TpLogInfo("Exit with success")
			os.Exit(atmi.SUCCEED)
		}
	}
}
