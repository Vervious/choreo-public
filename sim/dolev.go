package main

import (
	"math/rand"
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
	numNodes                 int
	uniformEdgeReliability   float64
	configDayMS              int
	configDaySkewStdDev      int // in MS
	configEveryDaySkewStdDev int
	crashProbability         float64
	configPartitionPerc      int // 0 - 100
	configPartitionTwoPerc   int
	constantMessageDelay     int // in MS
	roundTime                int // in MS
	roundTimeSkew            int
}

func (sn *SimulatedNetwork) Init() {
	sn.mu = &sync.Mutex{}
	sn.mailboxes = make(map[int](chan Message))
	sn.configDayMS = 25
	sn.configDaySkewStdDev = 0
	sn.configEveryDaySkewStdDev = 0
	sn.crashProbability = 0
	sn.constantMessageDelay = 0
	sn.roundTime = 10
	sn.roundTimeSkew = 0
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
	if id/sn.numNodes < sn.configPartitionPerc {
		return 0
	}
	if id/sn.numNodes < sn.configPartitionTwoPerc {
		return 1
	}
	return 2
}

func (sn *SimulatedNetwork) Broadcast(m Message) {
	fromPartition := sn.partitionNum(m.from)
	sn.mu.Lock()
	defer sn.mu.Unlock()
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
		mb := mailbox
		mo := m
		go func() {
			// now, delay every message
			if sn.constantMessageDelay > 0 {
				time.Sleep(time.Duration(sn.constantMessageDelay*10.0) * time.Millisecond) // delay is specified in cs
			}
			select {
			case mb <- mo:
			default:
			}
		}()
	}
}

// we assume each node knows the address of every other node
// In addition, we are implementing HONEST DOLEV's - there is no f + 1 threshold; as long as
// we get 1 sync message we are good to go.
func SpawnDolevNode(id int, sn *SimulatedNetwork, numTotalNodes int, ps *PublicSwarmUpdateServer, pinkslip chan struct{}) {

	state_crashed := false

	// we run a centisecond (cs) granular clock. (we can't exactly visualize ms granularity...)
	// clockSkew := int((rand.Float64()*1.0 - 1.0) * float64(sn.roundTimeSkew) / float64(sn.configDayMS) / 10.0) // configDayMS is actually in cs, and there are 10 ms in 1 cs
	clockSkew := 1000000 * float64(sn.roundTimeSkew) * 10.0 / float64(sn.configDayMS) / 10.0 * float64(id) * 10.0 // id is temp
	cachedSkewConfig := sn.roundTimeSkew
	cachedDay := sn.configDayMS

	clockDur := time.Duration(10000000+clockSkew) * time.Nanosecond
	clockTicker := time.NewTicker(clockDur)
	clockVal := 0
	roundNum := 0
	lastAdjustedClockVal := 0

	// skewedRoundLength := sn.roundTime + int((rand.Float64()*2.0-1.0)*float64(sn.roundTimeSkew))
	// skewedDayLength := sn.configDayMS + int((rand.Float64()*2.0-1.0)*float64(sn.configDaySkewStdDev))

	// cachedDTS := sn.configDaySkewStdDev

	// proposedDay := 0
	// correctedDay := 0
	// currentRound := 0
	myMailbox := sn.MailboxFor(id)
	// dur := int(math.Exp(math.Log(float64(sn.configDayMS)) + rand.NormFloat64()*float64(sn.configDaySkewStdDev)))
	// if dur < 1 {
	// 	dur = 1
	// }
	// D := time.Duration(dur) * time.Millisecond
	// ticker := time.NewTicker(D)

	// roundDur := sn.roundTime + int((rand.Float64()*2.0-1.0)*float64(sn.roundTimeSkew)) // in ms
	// roundDMS := time.Duration(roundDur) * time.Millisecond
	// roundTicker := time.NewTicker(roundDMS)

	// knownProposals := make([]int, numTotalNodes)
	// var crashRecovery <-chan time.Time

	// prevDur := dur
	for {
		select {
		// case <-crashRecovery:
		// 	// if nil, this case never triggers
		// 	state_crashed = false
		// 	// fire our timer
		// 	D = time.Duration(prevDur) * time.Millisecond
		// 	ticker = time.NewTicker(D)
		case <-clockTicker.C:
			// fmt.Printf("TICK%d;%d\n", clockVal, id)
			// advance our clock
			clockVal += 1
			newDay := false

			if clockVal%sn.configDayMS == 0 {
				// new day!
				// send a sync/new-day message with the current clock time
				// fmt.Println("Node reached new day: ", id)
				lastAdjustedClockVal = clockVal
				m := Message{}
				m.from = id
				m.name = "new-day"
				m.contents = clockVal
				sn.Broadcast(m)
				newDay = true
			}
			if clockVal%sn.roundTime == 0 {
				// new round!
				roundNum = clockVal / sn.roundTime

				// send a message to test msg-delivery-guarantee
				m := Message{}
				m.from = id
				m.name = "checkround"
				m.contents = roundNum
				sn.Broadcast(m)

				if newDay {
					ps.BroadcastSwarmUpdate(SwarmUpdate{id, roundNum, 1})
				} else {
					ps.BroadcastSwarmUpdate(SwarmUpdate{id, roundNum, 0})
				}

				if sn.roundTimeSkew != cachedSkewConfig || sn.configDayMS != cachedDay {
					cachedSkewConfig = sn.roundTimeSkew
					cachedDay = sn.configDayMS
					// clockSkew = int((rand.Float64()*1.0 - 1.0) * float64(sn.roundTimeSkew) / float64(sn.configDayMS) / 10.0)
					clockSkew = 1000000 * float64(sn.roundTimeSkew) * 10.0 / float64(sn.configDayMS) / 10.0 * float64(id) * 10.0 // id is temp
					// if id == 1 {
					// 	fmt.Println("hello")
					// 	fmt.Println(clockSkew)
					// }
					clockTicker.Stop()
					clockDur = time.Duration(10000000+clockSkew) * time.Nanosecond
					clockTicker = time.NewTicker(clockDur)
				}

				// // check if our clocks changed at all (ok this is actually dynamic...)
				// // actually... let's skew the actual clocks.
				// if sn.roundTimeSkew != cachedRTS {
				// 	skewedRoundLength = sn.roundTime + int((rand.Float64()*2.0-1.0)*float64(sn.roundTimeSkew))
				// 	cachedRTS = sn.roundTimeSkew

				// 	clockTicker.Stop()
				// 	clockDur = time.Duration(10) * time.Millisecond
				// 	clockTicker := time.NewTicker(clockDur)
				// }
				// if sn.configDaySkewStdDev != cachedDTS {
				// 	// ignore the day skew for now
				// 	skewedDayLength = sn.configDayMS + int((rand.Float64()*2.0-1.0)*float64(sn.configDaySkewStdDev))
				// 	cachedDTS = sn.configDaySkewStdDev
				// }

				// if skewedRoundLength < 1 {
				// 	skewedRoundLength = 1
				// }
				// if skewedDayLength < 1 {
				// 	skewedDayLength = 1
				// }
			}

		// case <-ticker.C:
		// 	if state_crashed {
		// 		break
		// 	}
		// 	proposedDay += 1
		// 	// send a sync message
		// 	m := Message{}
		// 	m.from = id
		// 	m.name = "proposed-day"
		// 	m.contents = proposedDay
		// 	// fmt.Println(m)
		// 	sn.Broadcast(m)
		// 	ps.BroadcastSwarmUpdate(SwarmUpdate{id, proposedDay, 0})

		// 	// check if new D or dynamic skew
		// 	dur := sn.configDayMS + int(rand.NormFloat64()*float64(sn.configDaySkewStdDev))
		// 	dur = dur + int(rand.NormFloat64()*float64(sn.configEveryDaySkewStdDev))
		// 	if dur < 1 {
		// 		dur = 1
		// 	}
		// 	if dur != prevDur {
		// 		ticker.Stop()
		// 		D = time.Duration(dur) * time.Millisecond
		// 		ticker = time.NewTicker(D)
		// 		prevDur = dur
		// 	}
		case m := <-myMailbox:
			// nodes are alive are always processing new day messages
			if state_crashed {
				break
			}
			// we have a message!
			switch m.name {
			case "checkround":
				if m.contents < roundNum {
					// we didn't wait long enough. For some reason, new-day messages trigger this prematurely.
					// fmt.Printf("Delta: \t%d\n", roundNum-m.contents)
					ps.NoteMessageDelivery(false, roundNum-m.contents)
				} else if m.contents > roundNum {
					// this is fine, our clock is slower than the sender
					ps.NoteMessageDelivery(true, roundNum-m.contents)
				} else {
					ps.NoteMessageDelivery(true, roundNum-m.contents)
				}
			case "proposed-day":
				// HONEST DOLEV's - every sync message is basically a new-day message
				// if m.contents <= knownProposals[m.from] {
				// 	break
				// }
				// knownProposals[m.from] = m.contents
				// // check if a majority of known proposals have advanced past some time
				// majorityCheck := make([]int, len(knownProposals))
				// copy(majorityCheck, knownProposals)
				// sort.Sort(sort.Reverse(sort.IntSlice(majorityCheck)))
				// majorityDay := majorityCheck[len(majorityCheck)/2]

				// if majorityDay <= correctedDay {
				// 	break
				// }
				// m.contents = majorityDay
				fallthrough
			case "new-day":
				if m.contents <= lastAdjustedClockVal {
					// this is unrelated, but we don't verify any proofs yet
					break
				}
				clockVal = m.contents
				lastAdjustedClockVal = clockVal
				roundNum = clockVal / sn.roundTime
				ps.BroadcastSwarmUpdate(SwarmUpdate{id, roundNum, 1})
				// gossip new day
				m := Message{}
				m.from = id
				m.name = "new-day"
				m.contents = clockVal
				sn.Broadcast(m)
				// correctedDay = m.contents
				// proposedDay = correctedDay
				// ticker.Stop()
				// roundTicker.Stop()

				// ps.BroadcastSwarmUpdate(SwarmUpdate{id, proposedDay, 1})
				// // gossip new-day
				// m := Message{}
				// m.from = id
				// m.name = "new-day"
				// m.contents = correctedDay
				// sn.Broadcast(m)

				// // check if we crashed...
				// if rand.Float64() < sn.crashProbability {
				// 	// edge failure
				// 	state_crashed = true
				// 	// recover after 1 second for now
				// 	G := time.Duration(1000) * time.Millisecond
				// 	crashRecoveryT := time.NewTimer(G)
				// 	crashRecovery = crashRecoveryT.C
				// 	break
				// }

				// // check if new D or dynamic skew
				// dur := sn.configDayMS + int(rand.NormFloat64()*float64(sn.configDaySkewStdDev))
				// dur = dur + int(rand.NormFloat64()*float64(sn.configEveryDaySkewStdDev))
				// if dur < 1 {
				// 	dur = 1
				// }
				// D = time.Duration(dur) * time.Millisecond
				// ticker = time.NewTicker(D)
				// roundTicker = time.NewTicker(roundDMS)
				// prevDur = dur
			}
		case <-pinkslip:
			return
		}
	}
}
