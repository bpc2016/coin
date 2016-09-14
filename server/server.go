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

var (
	index     = flag.Int("index", 0, "RPC port is 50051+index") //; debug port is 36661+index")
	numMiners = flag.Int("miners", 3, "number of miners")       // DOESNT include the external one
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

type blockdata struct {
	sync.Mutex
	data string
}

var block blockdata // models teh block information - basis of 'work'

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	if *debug {
		fmt.Printf("Work request: %+v\n", in)
	}
	s.signIn <- in.Name // register
	s.start.Wait()      // all must wait, start when this is Done()
	block.Lock()
	work := &cpb.Work{Coinbase: in.Name, Block: []byte(block.data)}
	block.Unlock()
	return &cpb.GetWorkReply{Work: work}, nil
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	won := s.vetWin(*soln.Win)
	return &cpb.AnnounceReply{Ok: won}, nil
}

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	s.signOut <- in.Name // register
	s.stop.Wait()        // wait for valid solution  OR timeout
	return &cpb.GetCancelReply{Ok: true}, nil
}

// vetWins handles wins - all are directed here
func (s *server) vetWin(thewin cpb.Win) bool {
	s.search.Lock()
	defer s.search.Unlock()
	won := false
	select {
	case <-s.search.ch: // closed if already have a winner
	default: // should vet the solution, return false otherwise!!
		close(s.search.ch) // until call for new run resets this one
		resultchan <- fmt.Sprintf("Winner: %+v\n", thewin)
		for i := 0; i < *numMiners; i++ {
			<-s.signOut
		}
		s.stop.Done()  // issue cancellation to our clients
		s.start.Add(1) // reset getwork start
		won = true
	}
	return won
}

type privchannel struct {
	sync.Mutex
	ch chan struct{}
}

// server is used to implement cpb.CoinServer.
type server struct {
	signIn  chan string    // for registering users in getwork
	signOut chan string    // for registering leaving users in getcancel
	search  privchannel    // search.ch is closed when we have dclared a winner
	start   sync.WaitGroup // this is the tricky one, stop here for company
	stop    sync.WaitGroup
}

//==========================================================================
var blockchan chan string

// this will implemented by cpb.server
func issueBlock() {
	//time.Sleep(8 * time.Second)
	blockchan <- fmt.Sprintf("BLOCK: %v", time.Now()) // comes from client = frontend
}

func (s *server) IssueBlock(ctx context.Context, in *cpb.IssueBlockRequest) (*cpb.IssueBlockReply, error) {
	return &cpb.IssueBlockReply{Ok: true}, nil
}

var resultchan chan string

// // this will implement dpb.server
// func getResult() {
// 	for {
// 		result := <-resultchan // wait for a result
// 		fmt.Print(result)      // send this back to client
// 		issueBlock()           ///BOGUS - this happens at the frontend in response ..
// 	}
// }

func (s *server) GetResult(ctx context.Context, in *cpb.GetResultRequest) (*cpb.GetResultReply, error) {
	result := <-resultchan // wait for a result
	fmt.Print(result)      // send this back to client
	issueBlock()           ///BOGUS - this happens at the frontend in response ..
	return &cpb.GetResultReply{Solution: result}, nil
}

//===========================================================================

var haveExternal bool

// this comes from this server's role as a client to frontend
func getNewBlock() {
	temp := <-blockchan // note that this will block if EXTERNAL absent
	block.Lock()
	block.data = temp //= fmt.Sprintf("BLOCK: %v", time.Now())
	block.Unlock()
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1
}

func main() {
	flag.Parse()

	*numMiners++ // to include the EXTERNAL

	port := fmt.Sprintf(":%d", 50051+*index)
	lis, err := net.Listen("tcp", port) // RPC port - localhost?
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := new(server)

	s.signIn = make(chan string)
	s.signOut = make(chan string)
	s.start.Add(1) // get work start

	blockchan = make(chan string, 1) // buffered
	resultchan = make(chan string)   //, 1) // buffered

	issueBlock() ///BOGUS

	go func() {
		for {
			getNewBlock()
			for i := 0; i < *numMiners; i++ { // loop blocks here until miners are ready
				<-s.signIn
			}
			fmt.Printf("\n--------------------\nNew race!\n")
			s.stop.Add(1)                     // prep channel for getcancels
			s.search.ch = make(chan struct{}) // reset this channel
			s.start.Done()                    // start our miners
		}
	}()

	// go getResult() // this is for later server implementation

	g := grpc.NewServer()
	cpb.RegisterCoinServer(g, s)
	g.Serve(lis)
}
