package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	cpb "coin/service"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "busiso"
)

// annouceWin is what causes the server to issue a cancellation
func annouceWin(c cpb.CoinClient, nonce uint32, coinbase string) {
	win := &cpb.Win{Coinbase: coinbase, Nonce: nonce}
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: win})
	if err != nil {
		log.Fatalf("could not announce win: %v", err)
	}
	if r.Ok { // it's possible that my winning nonce was late!
		fmt.Printf("== %d == FOUND it (%d)\n", myID, nonce)
	}
}

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string, waitForCancel chan struct{}) {
	if _, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name}); err != nil {
		log.Fatalf("could not request cancellation: %v", err)
	}
	close(waitForCancel) // assume that we got an ok=true
}

// dice
func toss() int {
	return rand.Intn(6)
}

// search tosses two dice waiting for a double 5. exit on cancel or win
func search(work *cpb.Work, waitForCancel chan struct{}) (uint32, bool) {
	var theNonce uint32
	var ok bool
	tick := time.Tick(1 * time.Second) // spin wheels
	for cn := 0; ; cn++ {
		a, b := toss(), toss() // toss ...
		if a == b && a == 5 {  // a win?
			theNonce = uint32(cn)
			ok = true
			break
		}

		select {
		case <-waitForCancel:
			goto done //return true
		default: // continue
		}

		<-tick
		fmt.Println(myID, " ", cn)
	}
done:
	return theNonce, ok
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

var myID uint32

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := cpb.NewCoinClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	r, err := c.Login(context.Background(), &cpb.LoginRequest{Name: name})
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	log.Printf("Login successful. Assigned id: %d\n", r.Id)

	myID = r.Id

	for {
		fmt.Printf("Fetching work %s ..\n", name)
		waitForCancel := make(chan struct{})
		// go out for work
		r, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: name})
		if err != nil {
			log.Fatalf("could not get work: %v", err)
		}
		log.Printf("Got work %+v\n", r.Work)
		// listen for a cancellation
		go getCancel(c, name, waitForCancel)
		// search blocks
		theNonce, ok := search(r.Work, waitForCancel)
		// a good place to check whether we are cancelled when we have a solution too
		if ok {
			fmt.Printf("%d ... sending solution (%d) \n", myID, theNonce)
			annouceWin(c, theNonce, r.Work.Coinbase)
		}

		<-waitForCancel // even if we have a solution
		fmt.Printf("-----------------------\n")
	}
}
