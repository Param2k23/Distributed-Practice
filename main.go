package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"strings"
	"sync"
)

type Bank struct {
	mu       sync.Mutex
	accounts map[string]int
}

type TransferArgs struct {
	From   string
	To     string
	Amount int
}

type PutArgs struct {
	User   string
	Amount int
}

// RPC Methods
func (b *Bank) Put(args *PutArgs, reply *bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.accounts[args.User] = args.Amount
	fmt.Printf("[Server] Deposited %d to %s\n", args.Amount, args.User)
	*reply = true
	return nil
}

func (b *Bank) Transfer(args *TransferArgs, reply *bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.accounts[args.From] < args.Amount {
		return fmt.Errorf("insufficient funds in %s's account", args.From)
	}

	//Move Money
	b.accounts[args.From] -= args.Amount
	b.accounts[args.To] += args.Amount
	fmt.Printf("[Server] Transferred %d from %s to %s\n", args.Amount, args.From, args.To)
	*reply = true
	return nil
}

func (b *Bank) Get(user string, balance *int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	*balance = b.accounts[user]
	return nil
}

func main() {
	//parse command line arguments
	portPtr := flag.String("port", "8001", "Port to listen on")
	idPtr := flag.Int("id", 0, "My Node ID (0,1,2,...)")
	peersPtr := flag.String("peers", "localhost:8001,localhost:8002,localhost:8003", "Comma-separated list of peer addresses")
	flag.Parse()

	peers := strings.Split(*peersPtr, ",")

	bank := new(Bank)
	bank.accounts = make(map[string]int)

	//initialize paxos
	px := MakePaxosPeer(*idPtr, peers)

	//register both bank and paxos RPCs
	rpc.Register(bank)
	rpc.Register(px)

	listener, err := net.Listen("tcp", ":"+*portPtr)
	if err != nil {
		fmt.Println("Listener error:", err)
		return
	}
	fmt.Printf("Node %d listening on port %s", *idPtr, *portPtr)

	//If i am the node 0 , i will try to be the leader (for testing only)
	if *idPtr == 0 {
		go func() {
			fmt.Println("Node 0 trying to be leader by sending Prepare to all peers")
			px.RunPaxos(nil)
		}()
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection accept error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
