package main

import (
	"coin"
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
	if skipF(c, "could not login", err) {
		return 0
	}

	log.Printf("Login successful. Assigned id: %d\n", r.Id)
	return r.Id
}

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string, stopLooking chan struct{}, endLoop chan struct{}) {
	_, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name})
	if skipF(c, "could not request cancellation", err) {
		return
	}

	stopLooking <- struct{}{} // stop search
	endLoop <- struct{}{}     // quit loop
}

var servers []cpb.CoinClient

// getResult makes a blocking request to the server
func getResult(c cpb.CoinClient, name string, theWinner chan string, lateEntry chan struct{}) {
	res, err := c.GetResult(context.Background(), &cpb.GetResultRequest{Name: name})
	if skipF(c, "could not request result", err) {
		return
	}

	if res.Winner.Identity != "EXTERNAL" { // avoid echoes
		declareWin(theWinner, lateEntry, res.Index, res.Winner.Identity, res.Winner.Nonce) // HL
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
			if uint32(i) == index || !alive[c] {
				continue
			}
			annouceWin(c, 99, []byte{}, "EXTERNAL") // bogus  announcement
		}
	}
}

// annouceWin is what causes the server to issue a  cancellation
func annouceWin(c cpb.CoinClient, nonce uint32, block []byte, winner string) bool {
	win := &cpb.Win{Block: block, Nonce: nonce, Identity: winner}
	// win := &cpb.Win{Coinbase: coinbase, Nonce: nonce}
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: win})
	if skipF(c, "could not announce win", err) {
		return false
	}

	return r.Ok
}

var alive map[cpb.CoinClient]bool

// utilities
func skipF(c cpb.CoinClient, message string, err error) bool {
	if err != nil {
		log.Printf(message+": %v", err)
		if alive[c] {
			alive[c] = false
		}
		return true // we have skipped
	}
	return false
}

func debugF(format string, args ...interface{}) {
	if *debug {
		log.Printf(format, args...)
	}
}

func newBlock() (upper, lower []byte, blockheight uint32) {
	blockHeight := 433789
	blockFees := 8756123 // satoshi
	pubkey := "0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086"
	// conductor generates this ..
	upper, lower, err := coin.CoinbaseTemplates(blockHeight, blockFees, pubkey)
	if err != nil {
		log.Fatalf("failed to generate coinbase: %v", err)
	}
	// sends upper, lower , blockHeight --> server
	return upper, lower, uint32(blockHeight)
}

func main() {
	flag.Parse()
	alive = make(map[cpb.CoinClient]bool)
	for index := 0; index < *numServers; index++ {
		addr := fmt.Sprintf("localhost:%d", 50051+index)
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		defer conn.Close()
		c := cpb.NewCoinClient(conn) // note that we do not login!
		servers = append(servers, c)
		alive[c] = true
	}
	// OMIT
	for {
		stopLooking := make(chan struct{}, *numServers)   // for search OMIT
		endLoop := make(chan struct{}, *numServers)       // for this loop OMIT
		serverUpChan := make(chan *cpb.Work, *numServers) // for gathering signins OMIT
		lateEntry := make(chan struct{})                  // no more results please OMIT
		theWinner := make(chan string, *numServers)       //  OMIT
		//newBlock := fmt.Sprintf("BLOCK: %v", time.Now())  // next block
		u, l, b := newBlock()
		// OMIT
		for _, c := range servers {
			go func(c cpb.CoinClient, // HL
				stopLooking chan struct{}, endLoop chan struct{},
				theWinner chan string, lateEntry chan struct{}) {
				_, err := c.IssueBlock(context.Background(), &cpb.IssueBlockRequest{Upper: u, Lower: l, Blockheight: b})
				if skipF(c, "could not issue block", err) {
					return
				}
				// conductor handles results
				go getResult(c, "EXTERNAL", theWinner, lateEntry) // HL
				// get ready, get set ... this needs to block  OMIT
				r, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: "EXTERNAL"}) // HL
				if skipF(c, "could not reconnect", err) {                                        // HL
					return
				} else if !alive[c] { // HL
					alive[c] = true
				}
				//  OMIT
				serverUpChan <- r.Work // HL
				// in parallel - seek cancellation
				go getCancel(c, "EXTERNAL", stopLooking, endLoop)
			}(c, stopLooking, endLoop, theWinner, lateEntry)
		}
		//  collect the work request acks from servers b OMIT
		for c := range alive {
			if !alive[c] {
				continue
			}
			debugF("%+v\n", <-serverUpChan)
		}
		// OMIT
		debugF("%s\n", "...") //  OMIT
		// 'search' - as the common 'External' miner
		theNonce, ok := search(stopLooking)
		if ok {
			declareWin(theWinner, lateEntry, uint32(*numServers), // HL
				"external", theNonce)
		}
		//  wait for server cancellation responses
		for c := range alive {
			if !alive[c] {
				continue
			}
			<-endLoop // wait for cancellation from each server
		}
		//  OMIT
		fmt.Println(<-theWinner, "\n---------------------------") // a OMIT
	}
} // c OMIT
