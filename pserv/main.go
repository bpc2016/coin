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

// logger type is for the users var  accounts for logins
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

type starts struct {
	countRunners int
	theLine      chan struct{}
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
	if getSet.on.countRunners == 0 { //fmt.Printf("GetWork lined=0, resetting waitFor.validStop ...\n")
		waitFor.mu.Lock()
		waitFor.validStop = make(chan struct{})
		waitFor.mu.Unlock()
	}
	getSet.on.countRunners++ //fmt.Printf("GetWork linedUp = %d\n", getSet.in.linedUp)
	if getSet.on.countRunners == enoughWorkers {
		close(getSet.on.theLine)
		go getNewBlocks() // models watching the entire network, timeout our search
		getSet.on.countRunners = 0
	}
	getSet.mu.Unlock()

	work := fetchWork(in) // get this in advance

	<-getSet.on.theLine // wait

	if getSet.on.countRunners == 0 { // we have just closed
		getSet.mu.Lock()
		getSet.on.theLine = make(chan struct{})
		getSet.mu.Unlock()
	}
	return &cpb.GetWorkReply{Work: work}, nil
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	fmt.Printf("verify solution: %+v\n", soln)
	watchNet.mu.Lock()
	if !weWonOutAlready() {
		close(watchNet.weWon) // cancel previous getNewBlocks
	}
	watchNet.mu.Unlock()
	endRun()
	checked := true // TODO - accommodate possible mistaken solution
	return &cpb.AnnounceReply{Ok: checked}, nil
}

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	// fmt.Printf("GetCancel request from %s\n", in.Name)
	<-waitFor.validStop // wait for valid solution  OR timeout

	fmt.Printf("OUT: %s\n", in.Name)
	return &cpb.GetCancelReply{Ok: true}, nil
}

type stoper struct {
	mu        sync.Mutex
	validStop chan struct{}
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

func weWonOutAlready() bool {
	select {
	case <-watchNet.weWon:
		return true
	default:
		return false
	}
}

// prepares the candidate block and also provides user specific coibase data
// TODO -this should return err as well
func fetchWork(in *cpb.GetWorkRequest) *cpb.Work {
	return &cpb.Work{Coinbase: in.Name, Block: []byte("hello world")}
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1

	getSet.on.theLine = make(chan struct{})
}

type external struct {
	mu    sync.Mutex
	weWon chan struct{} // closed if we do win
}

var watchNet external

func getNewBlocks() {
	watchNet.weWon = make(chan struct{})
	tick := time.Tick(1 * time.Second)
	period := 37 // seconds after which we cancel search
	for i := 0; i < period; i++ {
		<-tick // continue, but check every second
		//fmt.Printf("getnewblocks ... \n")
		if func() bool {
			select {
			case <-watchNet.weWon:
				return true
			default:
				return false
			}
		}() {
			return // because we won, endRun called elsewhere
		}
	}
	// otherwise reach this after period seconds
	endRun()
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
