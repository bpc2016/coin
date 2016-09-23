package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	cpb "coin/service"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	debug      = flag.Bool("d", false, "debug mode")
	numServers = flag.Int("s", 1, "number of servers  each at 50051+i, i =0 ,,,")
	timeOut    = flag.Int("o", 14, "timeout for EXTERNAL")
)

// 'search' here models external net: timeout after timeOut seconds
func search(stopLooking chan struct{}) (uint32, bool) {
	var theNonce uint32
	var ok bool
	tick := time.Tick(1 * time.Second)
	for cn := 0; ; cn++ {
		if cn >= *timeOut {
			theNonce = uint32(cn)
			ok = true
			break
		}
		// check for a stop order
		select {
		case <-stopLooking:
			goto done
		default: // continue
		}
		// wait for a second here ...
		<-tick
		debugF(" | EXT %d\n", cn)
	}

done:
	return theNonce, ok
}

// login to server c, returns a id
func login(c cpb.CoinClient, name string) uint32 {
	r, err := c.Login(context.Background(), &cpb.LoginRequest{Name: name})
	fatalF("could not login", err)

	log.Printf("Login successful. Assigned id: %d\n", r.Id)
	return r.Id
}

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string, stopLooking chan struct{}, endLoop chan struct{}) {
	_, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name})
	fatalF("could not request cancellation", err)

	stopLooking <- struct{}{} // stop search
	endLoop <- struct{}{}     // quit loop
}

var servers []cpb.CoinClient

// getResult makes a blocking request to the server
func getResult(c cpb.CoinClient, name string, theWinner chan string, lateEntry chan struct{}) {
	res, err := c.GetResult(context.Background(), &cpb.GetResultRequest{Name: name})
	fatalF("could not request result", err)

	if res.Winner.Coinbase != "EXTERNAL" { // avoid echoes
		declareWin(theWinner, lateEntry, res.Index, res.Winner.Coinbase, res.Winner.Nonce) // HL
	}
}

func declareWin(theWinner chan string, lateEntry chan struct{},
	index uint32, coinbase string, nonce uint32) {
	select {
	case <-lateEntry: // we already have declared a winner, do nothing
	default:
		close(lateEntry) // HL
		str := fmt.Sprintf("%s - ", time.Now().Format("15:04:05"))
		if index == uint32(*numServers) {
			str += "external" // HL
		} else {
			str += fmt.Sprintf("miner %d:%s, nonce %d", index, coinbase, nonce)
		}
		theWinner <- str // HL
		for i, c := range servers {
			if uint32(i) == index {
				continue
			}
			annouceWin(c, 99, "EXTERNAL") // bogus announcement
		}
	}
}

// annouceWin is what causes the server to issue a cancellation
func annouceWin(c cpb.CoinClient, nonce uint32, coinbase string) bool {
	win := &cpb.Win{Coinbase: coinbase, Nonce: nonce}
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: win})
	fatalF("could not announce win", err)

	return r.Ok
}

// utilities
func fatalF(message string, err error) {
	if err != nil {
		log.Fatalf(message+": %v", err)
	}
}

func debugF(format string, args ...interface{}) {
	if *debug {
		log.Printf(format, args...)
	}
}

func main() {
	flag.Parse()

	for index := 0; index < *numServers; index++ {
		addr := fmt.Sprintf("localhost:%d", 50051+index)
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		fatalF("fail to dial", err)

		defer conn.Close()
		c := cpb.NewCoinClient(conn)
		login(c, "EXTERNAL")
		servers = append(servers, c)
	}

	for {
		stopLooking := make(chan struct{}, *numServers) // for search
		endLoop := make(chan struct{}, *numServers)     // for this loop
		workChan := make(chan *cpb.Work, *numServers)   // for gathering signins
		lateEntry := make(chan struct{})                // no more results please
		theWinner := make(chan string, *numServers)
		newBlock := fmt.Sprintf("BLOCK: %v", time.Now()) // next block

		for _, c := range servers { // will need to use the index!!
			go func(c cpb.CoinClient, newBlock string,
				stopLooking chan struct{}, endLoop chan struct{},
				theWinner chan string, lateEntry chan struct{}) {
				_, err := c.IssueBlock(context.Background(), &cpb.IssueBlockRequest{Block: newBlock})
				fatalF("could not issue block", err)

				// conductor handles results
				go getResult(c, "EXTERNAL", theWinner, lateEntry)
				// get ready, get set ... this needs to block
				r, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: "EXTERNAL"})
				fatalF("could not get work", err)
				
				workChan <- r.Work // HL
				// in parallel - seek cancellation
				go getCancel(c, "EXTERNAL", stopLooking, endLoop)
			}(c, newBlock, stopLooking, endLoop, theWinner, lateEntry)
		}

		for i := 0; i < *numServers; i++ {
			debugF("%+v\n", <-workChan)
		}

		debugF("%s\n", "...")

		// 'search' blocks - the *sole* External one
		theNonce, ok := search(stopLooking)
		if ok {
			declareWin(theWinner, lateEntry, uint32(*numServers),
				"external", theNonce)
		}

		for i := 0; i < *numServers; i++ {
			<-endLoop // wait for cancellation from each server
		}

		fmt.Println(<-theWinner, "\n---------------------------")
	}
}
