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
	debug = flag.Bool("d", false, "debug mode")
	// server   = flag.Int("s", 0, "server is 50051+s - will include full URL later")
	servers = flag.Int("s", 1, "number of servers  each at 50051+i, i =0 ,,,")
	timeOut = flag.Int("o", 14, "timeout for EXTERNAL")
)

type coinServer cpb.CoinClient

// search here models external net: timeout after timeOut seconds
func search(look chan struct{}) (uint32, bool) {
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
			fmt.Println("EXT ", cn)
		}
	}

done:
	return theNonce, ok
}

// login to each server
func login(backends []cpb.CoinClient, name string) {
	// Contact each server and print out its response.
	for i, c := range backends {
		r, err := c.Login(context.Background(), &cpb.LoginRequest{Name: name})
		if err != nil {
			log.Fatalf("could not login to server %d : %v", i, err)
		}
		log.Printf("Login to server %d successful. Assigned id: %d\n", i, r.Id)
	}
}

// issue new block
func issueBlock(backends []cpb.CoinClient) {
	newBlock := fmt.Sprintf("BLOCK: %v", time.Now()) // next block
	for i, c := range backends {
		if _, err := c.IssueBlock(context.Background(), &cpb.IssueBlockRequest{Block: newBlock}); err != nil {
			log.Fatalf("could not issue block to server %d: %v", i, err)
		}
	}
}

// sign up with server c
func signUp(backends []cpb.CoinClient, name string) {
	// get ready, get set ... this needs to block at each server
	for i, c := range backends {
		_, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: name})
		if err != nil {
			log.Fatalf("could not get work on server %d : %v", i, err)
		}
	}
}

// getCancel makes a blocking request to the server
func getCancel(backends []cpb.CoinClient, name string, look chan struct{}, quit chan struct{}) {
	for i, c := range backends {
		res, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name})
		if err != nil {
			log.Fatalf("could not request cancellation on server %d : %v", i, err)
		}
		if *debug {
			fmt.Printf("cancel from server %d\n", res.Index)
		}
		look <- struct{}{} // stop search
		quit <- struct{}{} // quit loop
	}
}

// getResult makes a blocking request to the server
func getResult(backends []cpb.CoinClient, name string) {
	for i, c := range backends {
		res, err := c.GetResult(context.Background(), &cpb.GetResultRequest{Name: name})
		if err != nil {
			log.Fatalf("could not request result from server %d: %v", i, err)
		}
		fmt.Printf("%v\n", res.Winner)
		// then spread it around!
		for j, co := range backends {
			if j == i {
				continue
			}
			annouceWin(co, res.Winner.Nonce, res.Winner.Coinbase)
		}
	}
}

// annouceWin is what causes the server to issue a cancellation
func annouceWin(c coinServer, nonce uint32, coinbase string) bool {
	win := &cpb.Win{Coinbase: coinbase, Nonce: nonce}
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: win})
	if err != nil {
		log.Fatalf("could not announce win: %v", err)
	}
	return r.Ok
}

func main() {
	flag.Parse()

	// address := fmt.Sprintf("localhost:%d", 50051+*server) //"localhost:" + *server
	// Set up a connection to the server.
	// conn, err := grpc.Dial(address, grpc.WithInsecure())
	// if err != nil {
	// 	log.Fatalf("did not connect: %v", err)
	// }
	// defer conn.Close()

	//for _, addr := range strings.Split(*backends, ",")

	var backends []cpb.CoinClient

	for index := 0; index < *servers; index++ {
		addr := fmt.Sprintf("localhost:%d", 50051+index)
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		defer conn.Close()
		client := cpb.NewCoinClient(conn)
		backends = append(backends, client)
	}

	//c := cpb.NewCoinClient(conn)

	name := "EXTERNAL" // appears as EXTERNAL to servers

	login(backends, name) // login to backends

	// main cycle
	for {
		// frontend issues the blocks
		issueBlock(backends)
		// get on the start line with each server's other clients
		signUp(backends, name)
		if *debug {
			log.Printf("...\n")
		}

		look := make(chan struct{}, *servers) // for search
		quit := make(chan struct{}, *servers) // for this loop
		// look out for  cancellation
		go getCancel(backends, name, look, quit)
		// frontend handles results
		go getResult(backends, name)
		// search blocks
		theNonce, ok := search(look) // search here doesnt involve block!
		if ok {
			fmt.Printf("EXT ... sending solution (%d) \n", theNonce)
			win := true
			for _, c := range backends {
				win = win && annouceWin(c, theNonce, "EXTERNAL")
			}
			if win { // it's possible that my winning nonce was late!
				fmt.Printf("== EXT == FOUND -> %d\n", theNonce)
			}
		}

		for i := 0; i < *servers; i++ {
			<-quit // wait for cancellation from each server
		}

		fmt.Printf("-----------------------\n")
	}

}
