package main

import (
	"encoding/gob"
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
	paxos    *PxosPeer
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

	fmt.Printf("Request Received: %s sends $%d to %s\n", args.From, args.Amount, args.To)
	b.paxos.RunPaxos(args)

	if b.accounts[args.From] >= args.Amount {
		b.accounts[args.From] -= args.Amount
		b.accounts[args.To] += args.Amount
		*reply = true
		fmt.Printf("[Server] Transferred $%d from %s to %s\n", args.Amount, args.From, args.To)
	} else {
		*reply = false
		fmt.Printf("[Server] Transfer failed: Insufficient funds in %s\n", args.From)
	}
	return nil
}

func (b *Bank) Get(user string, balance *int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	*balance = b.accounts[user]
	return nil
}

func main() {
	//register gob types
	gob.Register(TransferArgs{})
	//parse command line arguments
	portPtr := flag.String("port", "8001", "Port to listen on")
	idPtr := flag.Int("id", 0, "My Node ID (0,1,2,...)")
	peersPtr := flag.String("peers", "localhost:8001,localhost:8002,localhost:8003", "Comma-separated list of peer addresses")
	flag.Parse()

	peers := strings.Split(*peersPtr, ",")

	//initialize paxos
	px := MakePaxosPeer(*idPtr, peers)

	bank := &Bank{
		accounts: make(map[string]int),
		paxos:    px,
	}

	//register both bank and paxos RPCs
	rpc.Register(bank)
	rpc.Register(px)

	listener, err := net.Listen("tcp", ":"+*portPtr)
	if err != nil {
		fmt.Println("Listener error:", err)
		return
	}
	fmt.Printf("Node %d listening on port %s", *idPtr, *portPtr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection accept error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
