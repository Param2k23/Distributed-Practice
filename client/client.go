package main

import (
	"fmt"
	"log"
	"net/rpc"
)

type PutArgs struct {
	User   string
	Amount int
}

type TransferArgs struct {
	From   string
	To     string
	Amount int
}

func main() {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		fmt.Println("Dialing error:", err)
		return
	}
	fmt.Println("Client connected to server.")
	putArgs := PutArgs{User: "Alice", Amount: 100}
	var reply bool

	err = client.Call("Bank.Put", &putArgs, &reply)
	if err != nil {
		log.Fatal("Put Error", err)
		return
	}
	fmt.Println("Deposited 100 to Alice's account.")

	bobArgs := PutArgs{User: "Bob", Amount: 0}
	err = client.Call("Bank.Put", &bobArgs, &reply)
	if err != nil {
		log.Fatal("Put Error", err)
		return
	}
	fmt.Println("Created Bob's account with 0 balance.")

	transferArgs := TransferArgs{From: "Alice", To: "Bob", Amount: 50}
	fmt.Println("Transferring 50 from Alice to Bob...")
	err = client.Call("Bank.Transfer", &transferArgs, &reply)
	if err != nil {
		log.Fatal("Transfer Error", err)
		return
	}
	fmt.Println("Transferred 50 from Alice to Bob.")

	var aliceBalance int
	err = client.Call("Bank.Get", "Alice", &aliceBalance)
	if err != nil {
		log.Fatal("Get Error", err)
		return
	}
	fmt.Printf("Alice's balance: %d\n", aliceBalance)
}
