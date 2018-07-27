package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/Vervious/eventsource"
)

type SwarmUpdate struct {
	NodeID    int
	ClockTime int
	Corrected int // 1 or 0, default 0
}

func JSONStr(su []SwarmUpdate) string {
	serialized, err := json.Marshal(su)
	if err != nil {
		panic("Serializing swarm update failed")
	}
	return string(serialized)
}

type PublicSwarmUpdateServer struct {
	mu                   *sync.Mutex
	stream               *eventsource.Stream
	lastSentId           int
	cachedSwarmUpdates   []SwarmUpdate
	cachedBroadcastTimer *time.Timer

	// we also handle toolbox
	sn *SimulatedNetwork
}

func (ps *PublicSwarmUpdateServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseMultipartForm(1000)
	toolboxUpdates := r.MultipartForm.Value
	log.Println(toolboxUpdates)
	// none of these sn writes are actually thread-safe/atomic,
	// but we will take the risk for now.
	if toolboxUpdates["edgeInput"] != nil {
		val := toolboxUpdates["edgeInput"][0]
		f, _ := strconv.ParseFloat(val, 64)
		f = f / 100.0
		fmt.Printf("new edge reliability: %f", f)
		ps.sn.uniformEdgeReliability = f
	}
	if toolboxUpdates["meanBaseClockSkew"] != nil {
		val := toolboxUpdates["meanBaseClockSkew"][0]
		dayMS, _ := strconv.Atoi(val)
		fmt.Printf("new base clock freq: %d", dayMS)
		ps.sn.configDayMS = dayMS
	}
	if toolboxUpdates["baseClockSkew"] != nil {
		val := toolboxUpdates["baseClockSkew"][0]
		clockSkew, _ := strconv.Atoi(val)
		fmt.Printf("new natural clock skew: %d", clockSkew)
		ps.sn.configDaySkewStdDev = clockSkew
	}
	if toolboxUpdates["dynClockSkew"] != nil {
		val := toolboxUpdates["dynClockSkew"][0]
		clockSkew, _ := strconv.Atoi(val)
		fmt.Printf("new random clock skew: %d", clockSkew)
		ps.sn.configEveryDaySkewStdDev = clockSkew
	}
	if toolboxUpdates["crashedProb"] != nil {
		val := toolboxUpdates["crashedProb"][0]
		f, _ := strconv.ParseFloat(val, 64)
		f = f / 100.0
		fmt.Printf("new crash proabbility: %f", f)
		ps.sn.crashProbability = f
	}
	if toolboxUpdates["partitionOnePerc"] != nil {
		val := toolboxUpdates["partitionOnePerc"][0]
		partitionSize, _ := strconv.Atoi(val)
		fmt.Printf("new partition perc: %d", partitionSize)
		ps.sn.configPartitionPerc = partitionSize
	}
	if toolboxUpdates["partitionTwoPerc"] != nil {
		val := toolboxUpdates["partitionTwoPerc"][0]
		partitionSize, _ := strconv.Atoi(val)
		fmt.Printf("new partition 2 perc: %d", partitionSize)
		ps.sn.configPartitionTwoPerc = partitionSize + ps.sn.configPartitionPerc
	}
	fmt.Fprintf(w, "success")
}

func (ps *PublicSwarmUpdateServer) ListenAndServe(port string) {
	ps.mu = &sync.Mutex{}
	ps.cachedSwarmUpdates = make([]SwarmUpdate, 0)
	ps.stream = eventsource.NewStream()

	fmt.Println("Sim server is now running on port 8910.")

	ps.stream.ClientConnectHook(func(r *http.Request, c *eventsource.Client) {
		fmt.Println("Received connection from", r.Host)
	})
	http.Handle("/toolbox", ps)
	http.Handle("/ev", ps.stream)
	http.ListenAndServe(":8910", nil)
}

func (ps *PublicSwarmUpdateServer) BroadcastSwarmUpdate(su SwarmUpdate) {
	// i'm not sure if eventsource is concurrent-safe, so serialize for now
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.cachedSwarmUpdates = append(ps.cachedSwarmUpdates, su)
	if ps.cachedBroadcastTimer == nil {
		// about 60Hz
		D := time.Duration(17) * time.Millisecond
		ps.cachedBroadcastTimer = time.NewTimer(D)
		go ps.BroadcastCachedSwarmUpdates()
	}
}

// only one of these should run at a time
func (ps *PublicSwarmUpdateServer) BroadcastCachedSwarmUpdates() {
	<-ps.cachedBroadcastTimer.C

	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.cachedBroadcastTimer.Stop()
	ps.cachedBroadcastTimer = nil

	e := eventsource.DataEvent(JSONStr(ps.cachedSwarmUpdates))
	e.ID(fmt.Sprintf("%d", ps.lastSentId)) // eventsource also has a factory that can do this, but keep it simple
	ps.stream.Broadcast(e)
	ps.lastSentId += 1
	ps.cachedSwarmUpdates = make([]SwarmUpdate, 0)
}

func main() {
	NUM_NODES := 100
	// this can be changed by toolbox so is just a starting value
	UNIFORM_EDGE_RELIABILITY := 0.8 // prob success

	ps := &PublicSwarmUpdateServer{}
	sn := &SimulatedNetwork{}
	ps.sn = sn
	sn.uniformEdgeReliability = UNIFORM_EDGE_RELIABILITY
	sn.Init() // I really wish go let you set default memory values
	pinkslip := listen_kill_signal()
	for i := 0; i < NUM_NODES; i++ {
		go SpawnDolevNode(i, sn, NUM_NODES, ps, pinkslip)
	}

	ps.ListenAndServe(":8910") // blocks
}

// listen to SIGINT and terminate gracefully
func listen_kill_signal() chan struct{} {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	pinkslip := make(chan struct{})
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		close(pinkslip)
		time.Sleep(1)
		panic("killed")
	}()
	return pinkslip
}
