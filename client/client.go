package main

import (
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
	debug  = flag.Bool("d", false, "debug mode")
	tosses = flag.Int("t", 2, "number of tosses")
	user   = flag.String("u", "sole", "the client name")
	server = flag.Int("s", 0, "server offset from 50051 - will include full URL later")

	myID uint32
)

// login to server c, returns an id
func login(c cpb.CoinClient, name string) uint32 {
	r, err := c.Login(context.Background(), &cpb.LoginRequest{Name: name}) // HL
	fatalF("could not login", err)

	log.Printf("Login successful. Assigned id: %d\n", r.Id)
	return r.Id
}

// sign up with server c
func signUp(c cpb.CoinClient, name string) *cpb.Work {
	r, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: name})
	fatalF("could not get work", err)

	debugF("Got work:\n%s\n", r.Work)
	return r.Work
}

// annouceWin is what causes the server to issue a cancellation
func annouceWin(c cpb.CoinClient, nonce uint32, coinbase string) bool {
	win := &cpb.Win{Coinbase: coinbase, Nonce: nonce}
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: win})
	fatalF("could not announce win", err)

	return r.Ok
}

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string, stopLooking chan struct{}, endLoop chan struct{}) {
	_, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name})
	fatalF("could not request cancellation", err)

	stopLooking <- struct{}{} // stop search
	endLoop <- struct{}{}     // quit loop
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
		case <-stopLooking:
			goto done // if so ... break out of this cycle, return (with ok=false!)
		default: // continue
		}
		// wait for a second here ...
		<-tick
		debugF("| %d %d\n", myID, cn)
	}

done:
	return theNonce, ok
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()

	address := fmt.Sprintf("localhost:%d", 50051+*server) //"localhost:" + *server
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	fatalF("did not connect", err)

	defer conn.Close()

	c := cpb.NewCoinClient(conn)
	name := *user
	// Contact the server and print out its response.
	myID = login(c, name)

	// main cycle OMIT
	for {
		fmt.Printf("Fetching work %s ..\n", name)
		work := signUp(c, name) // HL

		stopLooking := make(chan struct{}, 1) // HL
		endLoop := make(chan struct{}, 1)     // HL
		// look out for cancellation
		go getCancel(c, name, stopLooking, endLoop) // HL
		// search blocks
		theNonce, ok := search(work, stopLooking) // HL

		if ok { // we completed search
			fmt.Printf("%d ... sending solution (%d) \n", myID, theNonce)
			win := annouceWin(c, theNonce, work.Coinbase) // HL

			if win { // it's possible that my winning nonce was late!
				fmt.Printf("== %d == FOUND -> %d\n", myID, theNonce)
			}
		}

		<-endLoop // wait here for cancel from server
		fmt.Printf("-----------------------\n")
	}
	// main end OMIT
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
