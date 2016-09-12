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

const (
	address = "localhost:50051"
)

var (
	debug  = flag.Bool("d", false, "debug mode")
	tosses = flag.Int("t", 2, "number of tosses")
	user   = flag.String("u", "busiso", "the client name")
)
var waitForCancel chan struct{}

func getWork(c cpb.CoinClient, name string) {
	// get ready, get set ... this needs to block
	r, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: name})
	if err != nil {
		log.Fatalf("could not get work: %v", err)
	}
	if *debug {
		log.Printf("Got work %+v\n", r.Work)
	}
	// in parallel - seek cancellation
	go getCancel(c, name)
	// search blocks
	theNonce, ok := search(r.Work)
	if ok {
		fmt.Printf("%d ... sending solution (%d) \n", myID, theNonce)
		annouceWin(c, theNonce, r.Work.Coinbase)
	}
	<-waitForCancel
}

// annouceWin is what causes the server to issue a cancellation
func annouceWin(c cpb.CoinClient, nonce uint32, coinbase string) {
	win := &cpb.Win{Coinbase: coinbase, Nonce: nonce}
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: win})
	if err != nil {
		log.Fatalf("could not announce win: %v", err)
	}
	if r.Ok { // it's possible that my winning nonce was late!
		fmt.Printf("== %d == FOUND -> %d\n", myID, nonce)
	}
}

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string) {
	if _, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name}); err != nil {
		log.Fatalf("could not request cancellation: %v", err)
	}
	close(waitForCancel) // assume that we got an ok=true
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
func search(work *cpb.Work) (uint32, bool) {
	var theNonce uint32
	var ok bool
	tick := time.Tick(1 * time.Second) // spin wheels
	for cn := 0; ; cn++ {
		if rolls(*tosses) { // a win?
			theNonce = uint32(cn)
			ok = true
			break
		}
		// check for a stop order
		if gotcancel() {
			break
		}
		<-tick
		if *debug {
			fmt.Println(myID, " ", cn)
		}
	}
	return theNonce, ok
}

func gotcancel() bool {
	select {
	case <-waitForCancel:
		return true
	default:
		return false
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	waitForCancel = make(chan struct{})
}

var myID uint32

func main() {
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := cpb.NewCoinClient(conn)

	// Contact the server and print out its response.
	name := *user
	// if len(os.Args) > 1 {
	// 	name = os.Args[1]
	// }
	r, err := c.Login(context.Background(), &cpb.LoginRequest{Name: name})
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	log.Printf("Login successful. Assigned id: %d\n", r.Id)

	myID = r.Id

	for {
		fmt.Printf("Fetching work %s ..\n", name)

		getWork(c, name) // main work done here

		fmt.Printf("-----------------------\n")
		waitForCancel = make(chan struct{}) // reset channel
	}
}
