package main

import (
	"math/rand"
	"sort"
	"sync"
	"time"
)

// here's an actual dolev implementation simulation

// use strings since it's fast to hack around with
type Message struct {
	from     int
	name     string
	contents int
}

type SimulatedNetwork struct {
	mu                       *sync.Mutex
	mailboxes                map[int](chan Message)
	uniformEdgeReliability   float64
	configDayMS              int
	configDaySkewStdDev      int // in MS
	configEveryDaySkewStdDev int
	crashProbability         float64
	configPartitionPerc      int // 0 - 100
	configPartitionTwoPerc   int
}

func (sn *SimulatedNetwork) Init() {
	sn.mu = &sync.Mutex{}
	sn.mailboxes = make(map[int](chan Message))
	sn.configDayMS = 30
	sn.configDaySkewStdDev = 0
	sn.configEveryDaySkewStdDev = 0
	sn.crashProbability = 0
}

func (sn *SimulatedNetwork) MailboxFor(id int) chan Message {
	sn.mu.Lock()
	defer sn.mu.Unlock()
	if sn.mailboxes[id] == nil {
		// not buffered; drop the message if no one receiving
		// actually buffer 1 to pre-empt random sequencing bugs
		sn.mailboxes[id] = make(chan Message, 1)
	}
	return sn.mailboxes[id]
}

func (sn *SimulatedNetwork) partitionNum(id int) int {
	if id < sn.configPartitionPerc {
		return 0
	}
	if id < sn.configPartitionTwoPerc {
		return 1
	}
	return 2
}

func (sn *SimulatedNetwork) Broadcast(m Message) {
	sn.mu.Lock()
	defer sn.mu.Unlock()
	fromPartition := sn.partitionNum(m.from)
	for k, mailbox := range sn.mailboxes {
		if k == m.from {
			continue
		}
		// check if different partitions
		if fromPartition != sn.partitionNum(k) {
			// delivery failure
			continue
		}
		if rand.Float64() > sn.uniformEdgeReliability {
			// edge failure
			continue
		}
		select {
		case mailbox <- m:
		default:
		}
	}
}

// we assume each node knows the address of every other node
func SpawnDolevNode(id int, sn *SimulatedNetwork, numTotalNodes int, ps *PublicSwarmUpdateServer, pinkslip chan struct{}) {

	state_crashed := false

	proposedDay := 0
	correctedDay := 0
	myMailbox := sn.MailboxFor(id)
	dur := sn.configDayMS + int(rand.NormFloat64()*float64(sn.configDaySkewStdDev))
	if dur < 1 {
		dur = 1
	}
	D := time.Duration(dur) * time.Millisecond
	ticker := time.NewTicker(D)
	knownProposals := make([]int, numTotalNodes)
	var crashRecovery <-chan time.Time

	prevDur := dur
	for {
		select {
		case <-crashRecovery:
			// if nil, this case never triggers
			state_crashed = false
			// fire our timer
			D = time.Duration(prevDur) * time.Millisecond
			ticker = time.NewTicker(D)
		case <-ticker.C:
			if state_crashed {
				break
			}
			proposedDay += 1
			// send a sync message
			m := Message{}
			m.from = id
			m.name = "proposed-day"
			m.contents = proposedDay
			// fmt.Println(m)
			sn.Broadcast(m)
			ps.BroadcastSwarmUpdate(SwarmUpdate{id, proposedDay, 0})

			// check if new D or dynamic skew
			dur := sn.configDayMS + int(rand.NormFloat64()*float64(sn.configDaySkewStdDev))
			dur = dur + int(rand.NormFloat64()*float64(sn.configEveryDaySkewStdDev))
			if dur < 1 {
				dur = 1
			}
			if dur != prevDur {
				ticker.Stop()
				D = time.Duration(dur) * time.Millisecond
				ticker = time.NewTicker(D)
				prevDur = dur
			}
		case m := <-myMailbox:
			if state_crashed {
				break
			}
			// we have a message!
			switch m.name {
			case "proposed-day":
				if m.contents <= knownProposals[m.from] {
					break
				}
				knownProposals[m.from] = m.contents
				// check if a majority of known proposals have advanced past some time
				majorityCheck := make([]int, len(knownProposals))
				copy(majorityCheck, knownProposals)
				sort.Sort(sort.Reverse(sort.IntSlice(majorityCheck)))
				majorityDay := majorityCheck[len(majorityCheck)/2]

				if majorityDay <= correctedDay {
					break
				}
				m.contents = majorityDay
				// if id < sn.configPartitionPerc {
				// 	fmt.Println("EXCITING shit")
				// 	fmt.Println(m.from)
				// }
				fallthrough
			case "new-day":
				if m.contents <= correctedDay {
					// we don't verify any proofs yet
					break
				}
				correctedDay = m.contents
				proposedDay = correctedDay
				ticker.Stop()

				ps.BroadcastSwarmUpdate(SwarmUpdate{id, proposedDay, 1})
				// gossip new-day
				m := Message{}
				m.from = id
				m.name = "new-day"
				m.contents = correctedDay
				sn.Broadcast(m)

				// check if we crashed...
				if rand.Float64() < sn.crashProbability {
					// edge failure
					state_crashed = true
					// recover after 1 second for now
					G := time.Duration(1000) * time.Millisecond
					crashRecoveryT := time.NewTimer(G)
					crashRecovery = crashRecoveryT.C
					break
				}

				// check if new D or dynamic skew
				dur := sn.configDayMS + int(rand.NormFloat64()*float64(sn.configDaySkewStdDev))
				dur = dur + int(rand.NormFloat64()*float64(sn.configEveryDaySkewStdDev))
				if dur < 1 {
					dur = 1
				}
				D = time.Duration(dur) * time.Millisecond
				ticker = time.NewTicker(D)
				prevDur = dur
			}
		case <-pinkslip:
			return
		}
	}
}
