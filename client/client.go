package main

import (
	"fmt"
	"log"
	"net/rpc"
)

// Define the data structures used in RPC
type PutArgs struct {
	User   string
	Amount int
}

type TransferArgs struct {
	From   string
	To     string
	Amount int
}

// Helper function to connect to a specific port and run a transaction
func runTransaction(port string, from string, to string) {
	fmt.Printf("\n--- Connecting to Shard Node at %s ---\n", port)

	client, err := rpc.Dial("tcp", "localhost:"+port)
	if err != nil {
		fmt.Println("❌ Connection Failed:", err)
		return
	}
	defer client.Close()

	// 1. Setup Accounts (Put)
	var reply bool

	// Create "From" Account
	fmt.Printf("   -> Creating Account: %s ($100)\n", from)
	putArgs := PutArgs{User: from, Amount: 100}
	err = client.Call("Bank.Put", &putArgs, &reply)
	if err != nil {
		log.Println("      Put Error:", err)
	}

	// Create "To" Account
	fmt.Printf("   -> Creating Account: %s ($0)\n", to)
	bobArgs := PutArgs{User: to, Amount: 0}
	err = client.Call("Bank.Put", &bobArgs, &reply)
	if err != nil {
		log.Println("      Put Error:", err)
	}

	// 2. Execute Transfer
	fmt.Printf("   -> Attempting Transfer: %s sends $50 to %s\n", from, to)
	transferArgs := TransferArgs{From: from, To: to, Amount: 50}

	err = client.Call("Bank.Transfer", &transferArgs, &reply)
	if err != nil {
		// RPC Error
		log.Println("      ❌ RPC Error:", err)
		return
	}

	if reply {
		fmt.Println("      ✅ Success! Transaction Committed.")
	} else {
		fmt.Println("      ⛔ Failed! Server rejected the transaction (Wrong Shard or Insufficient Funds).")
	}

	// 3. Verify Balance (Optional)
	var balance int
	err = client.Call("Bank.Get", from, &balance)
	if err == nil {
		fmt.Printf("   -> %s's New Balance: $%d\n", from, balance)
	}
}

func main() {
	// TEST 1: Intra-Shard Transaction on Shard 1 (A-M)
	// Alice and Bob start with 'A' and 'B', so they belong to Shard 1.
	// We connect to Node 0 of Shard 1 (Port 8001).
	runTransaction("8001", "Alice", "Bob")

	// TEST 2: Intra-Shard Transaction on Shard 2 (N-Z)
	// Zelda and Xander start with 'Z' and 'X', so they belong to Shard 2.
	// We connect to Node 0 of Shard 2 (Port 9001).
	runTransaction("9001", "Zelda", "Xander")

	// TEST 3: Invalid Shard Access (Boundary Check)
	// We try to process "Zelda" on Shard 1 (Port 8001).
	// This SHOULD FAIL if your isMyKey() logic is working correctly.
	fmt.Println("\n--- TEST: Sending 'Zelda' to Shard 1 (Should Fail) ---")
	runTransaction("8001", "Zelda", "Xander")
}
