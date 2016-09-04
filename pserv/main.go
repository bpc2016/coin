package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"coin/cpb"

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
	mu       sync.Mutex //TODO make what follows a struct too?
	nextID   int
	loggedIn map[string]int
}

var users logger

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) {
	users.mu.Lock()
	if _, ok := users.loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	users.nextID++
	users.loggedIn[in.Name] = users.nextID
	users.mu.Unlock()
	return &cpb.LoginReply{Id: uint32(users.nextID)}, nil
}

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

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	fmt.Printf("GetWork req: %+v\n", in)

	getSet.mu.Lock()
	if getSet.on.numMiners == 0 {
		waitFor.mu.Lock()
		waitFor.validStop = make(chan struct{})
		waitFor.mu.Unlock()
	}
	getSet.on.numMiners++ // we simply count miners to signal search start
	if getSet.on.numMiners == enoughWorkers {
		close(getSet.on.theLine)
		go getNewBlocks() // models watching the entire network, timeout our search
		getSet.on.numMiners = 0
	}
	getSet.mu.Unlock()

	<-getSet.on.theLine // all must wait

	if getSet.on.numMiners == 0 { // we have just closed
		getSet.mu.Lock()
		getSet.on.theLine = make(chan struct{})
		getSet.mu.Unlock()
	}
	work := fetchWork(in) // work assigned this miner
	return &cpb.GetWorkReply{Work: work}, nil
}

// prepares the candidate block and also provides user specific coibase data
func fetchWork(in *cpb.GetWorkRequest) *cpb.Work { // TODO -this should return err as well
	return &cpb.Work{Coinbase: in.Name, Block: []byte("hello world")}
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	watchNet.mu.Lock()
	if !weWonOutAlready() {
		close(watchNet.weWon) // cancel previous getNewBlocks
	}
	watchNet.mu.Unlock()
	checked := true // TODO - accommodate possible mistaken solution
	fmt.Printf("We won!: %+v\n", soln)
	endRun()
	return &cpb.AnnounceReply{Ok: checked}, nil
}

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	<-waitFor.validStop // wait for valid solution  OR timeout
	fmt.Printf("OUT: %s\n", in.Name)
	return &cpb.GetCancelReply{Ok: true}, nil
}

// these structs control cancellation of searching
type stoper struct {
	mu        sync.Mutex
	validStop chan struct{} // closed if search cancelled
}

var waitFor stoper

func endRun() {
	waitFor.mu.Lock()
	if !closedAlready() {
		close(waitFor.validStop)
	}
	waitFor.mu.Unlock()
	fmt.Printf("New race!\n--------------------\n")
}

func closedAlready() bool {
	select {
	case <-waitFor.validStop:
		return true
	default:
		return false
	}
}

// these structs around the weWon channel control interaction with the network
type external struct {
	mu    sync.Mutex
	weWon chan struct{} // closed if we do win
}

var watchNet external

func weWonOutAlready() bool {
	select {
	case <-watchNet.weWon:
		return true
	default:
		return false
	}
}

// getNewBlocks watches the network for external winners and stops searah if we exceed period secs
func getNewBlocks() {
	watchNet.weWon = make(chan struct{})
	tick := time.Tick(1 * time.Second)
	period := 37 // seconds after which we cancel search
	for i := 0; i < period; i++ {
		<-tick // continue, but check every second
		if weWonOutAlready() {
			return // because we won, endRun called elsewhere
		}
	}
	// otherwise reach this after period seconds
	endRun()
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1

	getSet.on.theLine = make(chan struct{})
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	cpb.RegisterCoinServer(s, &server{})
	s.Serve(lis)
}
