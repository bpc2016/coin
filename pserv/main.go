package main

import (
	"coin/cpb"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port          = ":50051"
	enoughWorkers = 3
)

// server is used to implement cpb.CoinServer.
type server struct{}

// logger type is for the users login details
type logger struct {
	sync.Mutex //anonymous
	nextID     int
	loggedIn   map[string]int
}

var users logger

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) {
	users.Lock()
	defer users.Unlock()
	if _, ok := users.loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	users.nextID++
	users.loggedIn[in.Name] = users.nextID
	return &cpb.LoginReply{Id: uint32(users.nextID)}, nil
}

/*
// these structs control starting clients on the search
type starts struct {
	numMiners int
	theLine   chan struct{}
}

type starter struct {
	mu sync.Mutex
	on starts
}

var getSet starter
*/

type abort struct {
	sync.Mutex
	numMiners int
	theLine   chan struct{}
}

var getwork abort

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	fmt.Printf("GetWork req: %+v\n", in)
	// getSet.mu.Lock()
	// if getSet.on.numMiners == 0 {
	// getwork.Lock()
	// if getwork.numMiners == 0 {
	// 	endrun.New()
	// }
	// getwork.numMiners++
	// getSet.on.numMiners++ // we simply count miners to signal search start
	// if getSet.on.numMiners == enoughWorkers {
	// if getwork.numMiners == enoughWorkers {
	// close(getSet.on.theLine)
	// close(getwork.theLine)
	// go getNewBlocks() // models watching the entire network, timeout our search
	// getwork.numMiners = 0
	// getSet.on.numMiners = 0
	// }
	// getwork.Unlock()
	// getSet.mu.Unlock()
	gate <- in.Name // register
	//<-getSet.on.theLine // all must wait
	<-getwork.theLine // all must wait

	// if getSet.on.numMiners == 0 { // we have just closed
	// if getwork.numMiners == 0 { // we have just closed
	// 	getwork.Lock()
	// 	getwork.theLine = make(chan struct{})
	// 	getwork.Unlock()
	// 	// getSet.mu.Lock()
	// 	// getSet.on.theLine = make(chan struct{})
	// 	// getSet.mu.Unlock()
	// }
	work := fetchWork(in) // work assigned this miner
	return &cpb.GetWorkReply{Work: work}, nil
}

var gate chan string // unbuffered

func handleGate() {
	for {
		gate = make(chan string)
		fmt.Println("Gate ready ...")
		for i := 0; i < enoughWorkers; i++ {
			fmt.Printf("(%d) registered %s\n", i, <-gate)
		}
		close(getwork.theLine)
		go getNewBlocks() // models watching the entire network, timeout our search
		getwork.theLine = make(chan struct{})
		endrun.Revive()
	}
}

// prepares the candidate block and also provides user specific coibase data
func fetchWork(in *cpb.GetWorkRequest) *cpb.Work { // TODO -this should return err as well
	return &cpb.Work{Coinbase: in.Name, Block: []byte("hello world")}
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	newblock.Cancel() // cancel previous getNewBlocks
	checked := true   // TODO - accommodate possible mistaken solution
	fmt.Printf("We won!: %+v\n", soln)
	endRun()
	return &cpb.AnnounceReply{Ok: checked}, nil
}

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	<-endrun.Chan() //<-waitFor.validStop // wait for valid solution  OR timeout
	fmt.Printf("OUT: %s\n", in.Name)
	return &cpb.GetCancelReply{Ok: true}, nil
}

var endrun cpb.Abort

func endRun() {
	endrun.Cancel() // cancel waiting for a valid stop
	fmt.Printf("New race!\n--------------------\n")
}

var newblock cpb.Abort

// getNewBlocks watches the network for external winners and stops searah if we exceed period secs
func getNewBlocks() {
	newblock.New() // set up a new abort channel
	select {
	case <-newblock.Chan():
		return
	case <-time.After(17 * time.Second): // drop to endRun
	}
	// otherwise reach this after 17 seconds
	endRun()
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1

	getwork.theLine = make(chan struct{})
	// getSet.on.theLine = make(chan struct{})
	endrun.New()
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go handleGate()

	s := grpc.NewServer()
	cpb.RegisterCoinServer(s, &server{})
	s.Serve(lis)

}
