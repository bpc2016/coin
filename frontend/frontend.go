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
	servers = flag.Int("s", 1, "number of servers  each at 50051+i, i =0 ,,,")
	timeOut = flag.Int("o", 14, "timeout for EXTERNAL")
)

// 'search' here models external net: timeout after timeOut seconds
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

// login to server c, returns a id
func login(c cpb.CoinClient, name string) uint32 {
	// Contact the server and print out its response.
	r, err := c.Login(context.Background(), &cpb.LoginRequest{Name: name})
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	log.Printf("Login successful. Assigned id: %d\n", r.Id)
	return r.Id
}

// sign up with server c
func signUp(c cpb.CoinClient, name string) *cpb.Work {
	// get ready, get set ... this needs to block at each server
	r, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: name})
	if err != nil {
		log.Fatalf("could not get work: %v", err)
	}

	if *debug {
		log.Printf("Got work %+v\n", r.Work)
	}
	return r.Work
}

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string, look chan struct{}, quit chan struct{}) {
	if _, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name}); err != nil {
		log.Fatalf("could not request cancellation: %v", err)
	}
	look <- struct{}{} // stop search
	quit <- struct{}{} // quit loop
}

var backends []cpb.CoinClient

// getResult makes a blocking request to the server
func getResult(c cpb.CoinClient, name string) {
	res, err := c.GetResult(context.Background(), &cpb.GetResultRequest{Name: name})
	if err != nil {
		log.Fatalf("could not request result: %v", err)
	}
	if res.Winner.Coinbase != "EXTERNAL" {
		fmt.Printf("%d, %s, %d\n", res.Index, res.Winner.Coinbase, res.Winner.Nonce)
		for index, c := range backends {
			if uint32(index) == res.Index {
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
	if err != nil {
		log.Fatalf("could not announce win: %v", err)
	}
	return r.Ok
}

func main() {
	flag.Parse()

	for index := 0; index < *servers; index++ {
		addr := fmt.Sprintf("localhost:%d", 50051+index)
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		defer conn.Close()
		client := cpb.NewCoinClient(conn)
		login(client, "EXTERNAL")
		backends = append(backends, client)
	}

	for {
		look := make(chan struct{}, *servers)            // for search
		quit := make(chan struct{}, *servers)            // for this loop
		newBlock := fmt.Sprintf("BLOCK: %v", time.Now()) // next block
		for _, c := range backends {                     // will need to use teh index!!
			go func(c cpb.CoinClient, newBlock string, look chan struct{}, quit chan struct{}) {
				if _, err := c.IssueBlock(context.Background(), &cpb.IssueBlockRequest{Block: newBlock}); err != nil {
					log.Fatalf("could not issue block: %v", err)
				}
				// frontend handles results
				go getResult(c, "EXTERNAL")
				// get ready, get set ... this needs to block
				_, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: "EXTERNAL"})
				if err != nil {
					log.Fatalf("could not get work: %v", err)
				}
				if *debug {
					log.Printf("...\n")
				}
				// in parallel - seek cancellation
				go getCancel(c, "EXTERNAL", look, quit)
			}(c, newBlock, look, quit)
		}
		// 'search' blocks - the *sole* External one
		theNonce, ok := search(look)
		if ok {
			win := true
			for _, c := range backends {
				win = win && annouceWin(c, theNonce, "EXTERNAL") // our 'coinbase' nonce = 14 is from here
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
