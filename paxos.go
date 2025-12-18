package main

import (
	"fmt"
	"net/rpc"
	"sync"
)

type PrepareArgs struct {
	ProposalNumber int
	NodeID         int
}

type PrepareReply struct {
	Promise     bool
	HighestSeen int
}

type PxosPeer struct {
	mu          sync.Mutex
	me          int
	minProposal int
	peers       []string // List of peer addresses
}

func MakePaxosPeer(me int, peers []string) *PxosPeer {
	return &PxosPeer{
		me:          me,
		minProposal: -1,
		peers:       peers, // Initialize with provided peer list
	}
}

func (px *PxosPeer) Prepare(args *PrepareArgs, reply *PrepareReply) error {
	px.mu.Lock()
	defer px.mu.Unlock()

	fmt.Printf("[Paxos %d] Received Prepare with ProposalNumber %d from Node %d\n", px.me, args.ProposalNumber, args.NodeID)
	if args.ProposalNumber > px.minProposal {
		px.minProposal = args.ProposalNumber
		reply.Promise = true
		fmt.Printf("[Paxos %d] Promised for ProposalNumber %d\n", px.me, args.ProposalNumber)
	} else {
		reply.Promise = false
		reply.HighestSeen = px.minProposal
		fmt.Printf("[Paxos %d] Rejected Prepare for ProposalNumber %d, highest seen is %d\n", px.me, args.ProposalNumber, px.minProposal)
	}
	return nil
}

func (px *PxosPeer) SendPrepare(serverAddress string, args *PrepareArgs, reply *PrepareReply) bool {
	// connect to the peer
	client, err := rpc.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Printf("[Paxos %d] Failed to connect to %s: %v\n", px.me, serverAddress, err)
		return false
	}
	defer client.Close()

	//call prepare function
	err = client.Call("PxosPeer.Prepare", args, reply)
	if err != nil {
		fmt.Printf("[Paxos %d] RPC call to %s failed: %v\n", px.me, serverAddress, err)
		return false
	}
	return true
}

func (px *PxosPeer) RunPaxos(value interface{}) {
	currentProposalNumber := px.minProposal + 1
	for {
		//Setup Args
		args := &PrepareArgs{
			ProposalNumber: currentProposalNumber,
			NodeID:         px.me,
		}

		//count promises
		promisesReceived := 0
		// Ask everyone : looping through all peers
		for _, peerAddress := range px.peers {
			var reply PrepareReply
			ok := px.SendPrepare(peerAddress, args, &reply)
			if ok {
				if reply.Promise {
					promisesReceived++
				} else {
					if reply.HighestSeen >= currentProposalNumber {
						fmt.Printf("[Paxos %d] Updating proposal number from %d to %d based on rejection from peer\n", px.me, currentProposalNumber, reply.HighestSeen+1)
						currentProposalNumber = reply.HighestSeen + 1
					}
				}
			}
		}
		if promisesReceived > len(px.peers)/2 {
			fmt.Printf("WON THE ELECTION!! Proceeding to Accept Phase...")
			//(Next steps : Send Accept)
			return
		}
		fmt.Printf("Failes to get quorom, Retrying.....")
	}
}
