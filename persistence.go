package main

import (
	"encoding/gob"
	"fmt"
	"os"
)

type PaxosState struct {
	MinProposal      int
	AcceptedProposal int
	AcceptedValue    interface{}
}

//helper to save the state to disk

func (px *PxosPeer) persist() {
	filename := fmt.Sprintf("paxos_%d.log", px.me)

	//Open file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error saving state: %v\n", err)
		return
	}
	defer file.Close()

	//create state object
	state := PaxosState{
		MinProposal:      px.minProposal,
		AcceptedProposal: px.acceptedProp,
		AcceptedValue:    px.acceptedVal,
	}

	// Encode and Write
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(state)
	if err != nil {
		fmt.Printf("Error encoding state: %v\n", err)
	}
}

// helper to load state from disk on startup
func (px *PxosPeer) readPersist() {
	filename := fmt.Sprintf("paxos_%d.log", px.me)

	//Open file
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return // file does not exist, no need to read
	}
	if err != nil {
		fmt.Printf("Error reading state: %v\n", err)
		return
	}
	defer file.Close()

	// Decode and Read
	decoder := gob.NewDecoder(file)
	var state PaxosState
	err = decoder.Decode(&state)
	if err != nil {
		fmt.Printf("Error decoding state: %v\n", err)
		return
	}

	px.minProposal = state.MinProposal
	px.acceptedProp = state.AcceptedProposal
	px.acceptedVal = state.AcceptedValue

	fmt.Printf("[Recovery] Node %d restored state: MinProp %d, AcceptedProp %d\n", px.me, px.minProposal, px.acceptedProp)
}
