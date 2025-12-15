package main

import (
	"fmt"
	"net"
	"net/rpc"
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
	bank := new(Bank)
	bank.accounts = make(map[string]int)

	rpc.Register(bank)

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Listener error:", err)
		return
	}
	fmt.Println("Server listening on port 1234")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection accept error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
