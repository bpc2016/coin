package main

import (
	"coin"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	cpb "coin/service"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	debug       = flag.Bool("d", false, "debug mode")
	tosses      = flag.Int("t", 2, "number of tosses")
	user        = flag.String("u", "sole", "the client name")
	server      = flag.Int("s", 0, "server offset from 50051 - will include full URL later")
	maxSleep    = flag.Int("quit", 4, "number of multiples of 5 seconds before server declared dead")
	myID        uint32
	serverAlive bool
)

// annouceWin is what causes the server to issue a cancellation
func annouceWin(c cpb.CoinClient, nonce uint32, block []byte, winner string) bool {
	win := &cpb.Win{Block: block, Nonce: nonce, Identity: winner}
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: win})
	if skipF("could not announce win", err) {
		return false
	}
	return r.Ok
}

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string, stopLooking chan struct{}, endLoop chan struct{}) {
	_, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name})
	skipF("could not request cancellation", err) // drop through
	stopLooking <- struct{}{}                    // stop search
	endLoop <- struct{}{}                        // quit loop
}

// dice
func toss() int {
	return rand.Intn(6)
}

// rools returns true if n tosses are all 5's
func rolls(n int) bool {
	ok := true
	for i := 0; i < n; i++ {
		ok = ok && toss() == 5
		if !ok {
			break
		}
	}
	return ok
}

// search tosses two dice waiting for a double 5. exit on cancel or win
func search(work *cpb.Work, stopLooking chan struct{}) (uint32, bool) {
	// we must combine the coinbase + rest of block here  ...
	prepare(work)
	// toy version
	var theNonce uint32
	var ok bool
	tick := time.Tick(1 * time.Second)
	for cn := 0; ; cn++ {
		if rolls(*tosses) { // a win?
			theNonce = uint32(cn)
			debugF("winning! %d nonce: %d\n", myID, cn)
			ok = true
			break
		}
		// check for a stop order
		select {
		case <-stopLooking: // if so ... break out of this cycle, ok=false
			return theNonce, false
		default: // continue
		}
		// wait for a second here ...
		<-tick
		debugF("| %d %d\n", myID, cn)
	}
	return theNonce, ok
}

var (
	coinbase      coin.Transaction
	block         coin.Block
	target, share []byte
)

func prepare(work *cpb.Work) { //{Coinbase: coinbaseBytes, Block: partblock, Skel: merkSkel}
	//work.Coinbase
	coinbase = coin.Transaction(work.Coinbase)
	block = coin.Block(work.Block)
	cbhash := coin.Sha256(coinbase)                        // TODO - confirm this is not a double hash!!
	merkleroot, err := coin.Skel2Merkle(cbhash, work.Skel) // TODO - remove error here ...
	if err != nil {
		log.Fatal("failed to create merkelroot")
	}
	block.AddMerkle(merkleroot)
	target = coin.Bits2Target(work.Bits)
	/*
		this routine should place the coinbase in the blockheader
		compute the new merkle root hash and put in place
		change these when called to - each cycle? every overflow?

	*/
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()

	address := fmt.Sprintf("localhost:%d", 50051+*server) //"localhost:" + *server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := cpb.NewCoinClient(conn)
	name := *user
	serverAlive = true
	countdown := 0
	// outer OMIT
	for {
		r, err := c.Login(context.Background(), &cpb.LoginRequest{Name: name}) // HL
		if skipF("could not login", err) {                                     // HL
			time.Sleep(5 * time.Second)
			countdown++
			if countdown > *maxSleep {
				log.Fatalf("%s\n", "Server dead ... quitting") // HL
			}
			continue
		} else if !serverAlive { // HL
			countdown = 0
			serverAlive = true // we are back
		}
		log.Printf("Login successful. Assigned id: %d\n", r.Id)
		myID = r.Id
		// main cycle OMIT
		for {
			var ( // OMIT
				work                 *cpb.Work     // OMIT
				stopLooking, endLoop chan struct{} // OMIT
				theNonce             uint32        // OMIT
				ok                   bool          // OMIT
			) // OMIT
			fmt.Printf("Fetching work %s ..\n", name)
			r, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: name})
			if skipF("could not get work", err) {
				break
			}
			work = r.Work                        // HL
			stopLooking = make(chan struct{}, 1) // HL
			endLoop = make(chan struct{}, 1)     // HL
			// look out for  cancellation
			go getCancel(c, name, stopLooking, endLoop) // HL
			// search blocks
			theNonce, ok = search(work, stopLooking) // HL
			if ok {                                  // we completed search
				fmt.Printf("%d ... sending solution (%d) \n", myID, theNonce)
				win := annouceWin(c, theNonce, work.Block, name) // HL
				if win {                                         // late?
					fmt.Printf("== %d == FOUND -> %d\n", myID, theNonce)
				}
			}
			<-endLoop // wait here for cancel from server
			fmt.Printf("-----------------------\n")
		}
		// main end OMIT
	}
} // outerend OMIT

// utilities
func skipF(message string, err error) bool {
	if err != nil {
		if *debug {
			log.Printf(message+": %v", err)
		} else {
			log.Println(message)
		}
		if serverAlive {
			serverAlive = false
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
