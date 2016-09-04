package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"coin/cpb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "busiso"
)

func getWork(c cpb.CoinClient, name string) *cpb.Work {
	// get ready, get set ... this needs to block
	r, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: name})
	if err != nil {
		log.Fatalf("could not get work: %v", err)
	}
	log.Printf("Got work %+v\n", r.Work)
	return r.Work
}

func annouceWin(c cpb.CoinClient, nonce uint32, coinbase string) {
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: toWin(nonce, coinbase)})
	if err != nil {
		log.Fatalf("could not announce win: %v", err)
	}
	log.Printf("Solution verified: %+v\n", r.Ok)
}

var wait chan struct{}

func gotcancel() bool {
	select {
	case <-wait:
		return true
	default:
		return false
	}
}

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string) {
	r, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name})
	if err != nil {
		log.Fatalf("could not request cancellation: %v", err)
	}
	log.Printf("Got cancel message: %+v\n", r.Ok)
	//cancelled = r.Ok
	close(wait) // assume that we got an ok=true
}

// search tosses two dice waiting for a double 5
func search(c cpb.CoinClient, id uint32, work *cpb.Work) {
	var foundit uint32
	tick := time.Tick(1 * time.Second) // spin wheels
	for cn := 0; cn < 50; cn++ {
		a, b := toss(), toss()
		if a == b && a == 5 {
			foundit = uint32(cn) // our nonce
			break
		}
		<-tick
		fmt.Println(id, " ", cn)
		// check for a stop order
		if gotcancel() {
			break
		}
	}
	if foundit > 0 {
		fmt.Printf("== %d == FOUND it\n", id)
		annouceWin(c, foundit, work.Coinbase) // toServer <- id // declare we are done
	}
	fmt.Printf("[%d] .. BYE\n", id)
}

// TODO Move these out ??

// dice
func toss() int {
	return rand.Intn(6)
}

// convert nonce/coinbase to a cpb.Win object for annouceWin
func toWin(nonce uint32, coinbase string) *cpb.Win {
	return &cpb.Win{Coinbase: coinbase, Nonce: nonce}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	wait = make(chan struct{})
}

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

	// TODO - convert the context to have a timeout here

	// search

	fmt.Printf("fetching work %s ..\n", name)
	task := getWork(c, name)
	go getCancel(c, name)
	search(c, r.Id, task)
	// final wait
	for {
		if gotcancel() {
			break
		}
	}
	fmt.Printf("LOGOUT: %s ..\n-----------------------\n", name)

	for k := 0; k < 5; k++ {
		wait = make(chan struct{})
		fmt.Printf("fetching work %s ..\n", name)
		task = getWork(c, name)
		go getCancel(c, name)
		search(c, r.Id, task)
		// final wait
		for {
			if gotcancel() {
				break
			}
		}
		fmt.Printf("LOGOUT: %s ..\n-----------------------\n", name)
	}
}
