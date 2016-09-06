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

var waitForCancel chan struct{}

func getWork(c cpb.CoinClient, name string) {
	// get ready, get set ... this needs to block
	r, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: name})
	if err != nil {
		log.Fatalf("could not get work: %v", err)
	}
	log.Printf("Got work %+v\n", r.Work)
	go getCancel(c, name)
	// search blocks
	theNonce, ok := search(r.Work)
	if ok {
		fmt.Printf("%d ... sending solution \n", myID)
		annouceWin(c, theNonce, r.Work.Coinbase)
	}
	<-waitForCancel
}

// annouceWin is what causes the server to issue a cancellation
func annouceWin(c cpb.CoinClient, nonce uint32, coinbase string) {
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: toWin(nonce, coinbase)})
	if err != nil {
		log.Fatalf("could not announce win: %v", err)
	}
	if r.Ok { // it's possible that my winning nonce was late!
		// log.Print("I won!\n")
		fmt.Printf("== %d == FOUND it (%d)\n", myID, nonce)
	}
}

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string) { //tODO make use of ctx.Done()
	if _, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name}); err != nil {
		log.Fatalf("could not request cancellation: %v", err)
	}
	log.Printf("Got cancellation!\n")
	close(waitForCancel) // assume that we got an ok=true
}

// search tosses two dice waiting for a double 5. exit on cancel or win
func search(work *cpb.Work) (uint32, bool) {
	var theNonce uint32
	var ok bool
	tick := time.Tick(1 * time.Second) // spin wheels
	for cn := 0; cn < 100; cn++ {
		a, b := toss(), toss()
		if a == b && a == 5 { // a win?
			theNonce = uint32(cn)
			ok = true
			break
		}
		// check for a stop order
		if gotcancel() {
			break
		}
		<-tick
		fmt.Println(myID, " ", cn)
	}
	// fmt.Printf("[%d] .. BYE\n", myID)
	return theNonce, ok
}

// dice
func toss() int {
	return rand.Intn(6)
}

// convert nonce/coinbase to a cpb.Win object for annouceWin
func toWin(nonce uint32, coinbase string) *cpb.Win {
	return &cpb.Win{Coinbase: coinbase, Nonce: nonce}
}

func gotcancel() bool {
	select {
	case <-waitForCancel:
		return true
	default:
		return false
	}
}

// tell server that we are ready
func announceReady(user string) {
	fmt.Printf("LOGOUT: %s ..\n-----------------------\n", user)
	// reset channel
	waitForCancel = make(chan struct{})
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	waitForCancel = make(chan struct{})
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

	// TODO - convert the context to have a timeout here

	// search // for k := 0; ; k++ {

	for {
		fmt.Printf("fetching work %s ..\n", name)
		getWork(c, name)
		announceReady(name)
	}
}
