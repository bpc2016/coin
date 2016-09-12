package main

import (
	cpb "coin/service"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port    = ":50051"
	timeOut = 14
)

var (
	numMiners = flag.Int("m", 3, "number of miners")
	debug     = flag.Bool("d", false, "debug mode")
)

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

var start sync.WaitGroup // var start chan struct{}

// var signIn chan string

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	if *debug {
		fmt.Printf("Work requested: %+v\n", in)
	}
	s.signIn <- in.Name // register
	start.Wait()        //<-start        // all must wait
	return &cpb.GetWorkReply{Work: fetchWork(in.Name)}, nil
}

var block string

// prepares the candidate block and also provides user specific coibase data
func fetchWork(name string) *cpb.Work { // TODO -this should return err as well
	return &cpb.Work{Coinbase: name, Block: []byte(block)}
}

type win struct {
	who  string
	data cpb.Win
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	won := s.vetWin(*soln.Win)
	return &cpb.AnnounceReply{Ok: won}, nil
}

// extAnnounce is the analogue of 'Announce'
func (s *server) extAnnounce() {
	select {
	case <-settled.ch:
		return
	case <-time.After(timeOut * time.Second):
		s.vetWin(cpb.Win{Coinbase: "external", Nonce: 0}) // bogus
		return
	}
}

// var runover chan struct{}
var runover sync.WaitGroup

// var signOut chan string

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	s.signOut <- in.Name // register
	runover.Wait()       //<-runover          // wait for valid solution  OR timeout
	return &cpb.GetCancelReply{Ok: true}, nil
}

var settled struct {
	sync.Mutex
	ch chan struct{}
}

// vetWins handle wins - all are directed here
func (s *server) vetWin(thewin cpb.Win) bool {
	settled.Lock()
	defer settled.Unlock()
	select {
	case <-settled.ch: // closed if already have a winner
		return false
	default:
		fmt.Printf("Winner is: %+v\n", thewin)
		close(settled.ch) // until call for new run resets this one
		for i := 0; i < *numMiners; i++ {
			<-s.signOut
		}
		runover.Done() //close(runover) // issue cancellation to our clients
		start.Add(1)   //start = make(chan struct{})
		block = fmt.Sprintf("BLOCK: %v", time.Now())

		fmt.Printf("\n--------------------\nNew race!\n")
		return true
	}
}

// server is used to implement cpb.CoinServer.
type server struct {
	signIn  chan string
	signOut chan string
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := new(server)

	s.signIn = make(chan string)
	s.signOut = make(chan string)
	block = fmt.Sprintf("BLOCK: %v", time.Now())
	start.Add(1) //start = make(chan struct{})

	go func() {
		for {
			for i := 0; i < *numMiners; i++ {
				<-s.signIn
			}
			runover.Add(1) //runover = make(chan struct{})
			settled.Lock()
			settled.ch = make(chan struct{})
			settled.Unlock()
			start.Done()       // close(start)       // start our miners
			go s.extAnnounce() // start external miners
		}
	}()

	g := grpc.NewServer()
	cpb.RegisterCoinServer(g, s)
	g.Serve(lis)
}
