package main

import (
	cpb "coin/service"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

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
	sync.Mutex
	nextID   int
	loggedIn map[string]int
}

type blockdata struct {
	sync.Mutex
	data string
}

type lockable struct {
	sync.Mutex
	winnerFound bool
	ch          chan struct{}
}

var (
	users   logger
	block   blockdata // models the block information - basis of 'work'
	run     lockable
	signIn  chan string // for registering users in getwork
	signOut chan string // for registering leaving users in getcancel
	stop    sync.WaitGroup
)

// Login implements cpb.CoinServer login OMIT
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) { // HL
	users.Lock()
	defer users.Unlock()
	if _, ok := users.loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	users.nextID++
	users.loggedIn[in.Name] = users.nextID
	return &cpb.LoginReply{Id: uint32(users.nextID)}, nil
}

// nigol OMIT

// GetWork implements cpb.CoinServer, synchronises start of miners, hands out work gw OMIT
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	debugF("Work request: %+v\n", in)
	signIn <- in.Name // HL
	<-run.ch          // HL

	block.Lock()
	work := &cpb.Work{Coinbase: in.Name, Block: []byte(block.data)}
	block.Unlock()
	return &cpb.GetWorkReply{Work: work}, nil
}

// wg OMIT
// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	run.Lock()
	defer run.Unlock()
	if run.winnerFound {
		return &cpb.AnnounceReply{Ok: false}, nil
	}
	// we have a winner
	run.winnerFound = true // HL
	resultchan <- *soln.Win
	for i := 0; i < *numMiners; i++ {
		<-signOut
	}
	run.ch = make(chan struct{}) // HL
	stop.Done()                  // HL
	return &cpb.AnnounceReply{Ok: true}, nil
}

// cancel OMIT
// GetCancel broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	signOut <- in.Name // HL
	stop.Wait()        // HL
	serv := *index
	return &cpb.GetCancelReply{Index: uint32(serv)}, nil // HL
}

// lecnac OMIT

// server is used to implement cpb.CoinServer.
type server struct{}

var blockchan chan string

// IssueBlock receives the new block from Conductor : implements cpb.CoinServer
func (s *server) IssueBlock(ctx context.Context, in *cpb.IssueBlockRequest) (*cpb.IssueBlockReply, error) {
	blockchan <- in.Block
	return &cpb.IssueBlockReply{Ok: true}, nil
}

// this comes from this server's role with Conductor as client
func getNewBlock() {
	temp := <-blockchan // note that this will block if EXTERNAL absent
	block.Lock()
	block.data = temp
	block.Unlock()
}

var resultchan chan cpb.Win //string

// GetResult sends back win to Conductor : implements cpb.CoinServer
func (s *server) GetResult(ctx context.Context, in *cpb.GetResultRequest) (*cpb.GetResultReply, error) {
	result := <-resultchan                             // wait for a result
	fmt.Printf("sendresult: %d, %v\n", *index, result) // send this back to client
	return &cpb.GetResultReply{Winner: &result, Index: uint32(*index)}, nil
}

// utilities
func fatalF(message string, err error) {
	if err != nil {
		log.Fatalf(message+": %v", err)
	}
}

func debugF(format string, args ...interface{}) {
	if *debug {
		log.Printf(format, args...)
	}
}

func main() {
	flag.Parse()
	users.loggedIn = make(map[string]int)
	users.nextID = -1
	*numMiners++ // to include the Conductor (EXTERNAL)

	port := fmt.Sprintf(":%d", 50051+*index)
	lis, err := net.Listen("tcp", port)
	fatalF("failed to listen", err)

	signIn = make(chan string, *numMiners)  // register incoming miners
	signOut = make(chan string, *numMiners) // register miners receipt of cancel instructions
	blockchan = make(chan string, 1)        // transfer block data
	run.ch = make(chan struct{})            // signal to start mining
	resultchan = make(chan cpb.Win)         // transfer solution data
	// loop OMIT
	go func() {
		for {
			run.Lock()
			run.winnerFound = false
			run.Unlock()
			getNewBlock()                     // HL
			for i := 0; i < *numMiners; i++ { // loop blocks here until miners are ready
				<-signIn
			}

			fmt.Printf("\n--------------------\nNew race!\n")
			stop.Add(1)   // HL
			close(run.ch) // HL
		}
	}()
	// pool OMIT
	s := new(server)
	g := grpc.NewServer()
	cpb.RegisterCoinServer(g, s)
	g.Serve(lis)
}
