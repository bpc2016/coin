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
	debug   = flag.Bool("d", false, "debug mode")
	server  = flag.Int("s", 0, "server is 50051+s - will include full URL later")
	timeOut = flag.Int("o", 14, "timeout for EXTERNAL")
)

// annouceWin is what causes the server to issue a cancellation
func annouceWin(c cpb.CoinClient, nonce uint32, coinbase string) bool {
	win := &cpb.Win{Coinbase: coinbase, Nonce: nonce}
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: win})
	if err != nil {
		log.Fatalf("could not announce win: %v", err)
	}
	return r.Ok
}

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string, look chan struct{}, quit chan struct{}) {
	res, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name})
	if err != nil {
		log.Fatalf("could not request cancellation: %v", err)
	}
	if *debug {
		fmt.Printf("cancel from server %+v\n", res.Index)
	}
	look <- struct{}{} // stop search
	quit <- struct{}{} // quit loop
}

// getResult makes a blocking request to the server
func getResult(c cpb.CoinClient, name string) {
	res, err := c.GetResult(context.Background(), &cpb.GetResultRequest{Name: name})
	if err != nil {
		log.Fatalf("could not request result: %v", err)
	}
	fmt.Printf("%v\n", res.Solution)
}

// search here models external net: timeout after timeOut seconds
func search(work *cpb.Work, look chan struct{}) (uint32, bool) {
	var theNonce uint32
	var ok bool
	tick := time.Tick(1 * time.Second) // spin wheels
	for cn := 0; ; cn++ {
		if cn >= *timeOut { // a win? - for the frontend, this is just a timeout
			theNonce = uint32(cn)
			ok = true
			break
		}
		// check for a stop order
		select {
		case <-look:
			goto done // if so ... break out of this cycle, return (with ok=false!)
		default: // continue
		}
		// wait for a second here ...
		<-tick
		if *debug {
			fmt.Println(myID, " ", cn)
		}
	}

done:
	return theNonce, ok
}

var myID uint32

func main() {
	flag.Parse()

	address := fmt.Sprintf("localhost:%d", 50051+*server) //"localhost:" + *server
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := cpb.NewCoinClient(conn)

	name := "EXTERNAL" // appears as EXTERNAL to servers
	// Contact the server and print out its response.
	r, err := c.Login(context.Background(), &cpb.LoginRequest{Name: name})
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	log.Printf("Login successful. Assigned id: %d\n", r.Id)

	myID = r.Id

	// main cycle
	for {
		// frontend issues the blocks
		newBlock := fmt.Sprintf("BLOCK: %v", time.Now()) // next block
		if _, err := c.IssueBlock(context.Background(), &cpb.IssueBlockRequest{Block: newBlock}); err != nil {
			log.Fatalf("could not issue block: %v", err)
		}
		// frontend handles results
		go getResult(c, name)
		// get ready, get set ... this needs to block
		r, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: name})
		if err != nil {
			log.Fatalf("could not get work: %v", err)
		}
		if *debug {
			log.Printf("...\n")
		}
		look := make(chan struct{}, 1) // for search
		quit := make(chan struct{}, 1) // for this loop
		// look out for  cancellation
		go getCancel(c, name, look, quit)
		// search blocks
		theNonce, ok := search(r.Work, look)
		if ok {
			fmt.Printf("%d ... sending solution (%d) \n", myID, theNonce)
			win := annouceWin(c, theNonce, r.Work.Coinbase)
			if win { // it's possible that my winning nonce was late!
				fmt.Printf("== %d == FOUND -> %d\n", myID, theNonce)
			}
		}

		<-quit // wait for cancellation from server

		fmt.Printf("-----------------------\n")
	}

}
