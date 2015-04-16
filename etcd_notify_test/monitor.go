package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"strconv"
	"strings"
	"time"
)

var EtcdAddr = flag.String("m", "http://localhost:4001", "string array of etcd address, seperated by ';'")

var KeyBase string = "/update/t"

const (
	EtcdActionGet              = "get"
	EtcdActionCreate           = "create"
	EtcdActionSet              = "set"
	EtcdActionUpdate           = "update"
	EtcdActionDelete           = "delete"
	EtcdActionCompareAndSwap   = "compareAndSwap"
	EtcdActionCompareAndDelete = "compareAndDelete"
	EtcdActionExpire           = "expire"
)

type TestSuit struct {
	running    bool
	start      int64 // start time
	end        int64 // end time
	count      int64 // update count
	latencySum int64 // sum of each update operation's latency
}

func getEtcdClient(machines []string) *etcd.Client {
	return etcd.NewClient(machines)
}

// return true to stop the benchmark
func processNotify(data *etcd.Response, stop chan bool, ts *TestSuit) bool {
	if data.Action == EtcdActionDelete {
		ts.end = time.Now().UnixNano()
		return true
	} else if data.Action == EtcdActionUpdate {
		now := time.Now().UnixNano()
		if !ts.running {
			ts.running = true
			ts.start = now
		}
		ts.count++
		if lstart, err := strconv.ParseInt(data.Node.Value, 10, 0); err != nil {
			panic(err)
		} else {
			latency := now - lstart
			// fmt.Printf("latency: %.3f ms\n", float64(latency)/1e6)
			ts.latencySum += latency
		}
	}
	return false
}

func DataNodeWatcher(c *etcd.Client, prefix string, ts *TestSuit) {
	receiver := make(chan *etcd.Response)
	stop := make(chan bool)

	recursive := true
	go c.Watch(prefix, 0, recursive, receiver, stop)

	for {
		select {
		case data := <-receiver:
			if processNotify(data, stop, ts) {
				return
			}
		}
	}
}

func main() {
	machines := strings.Split(*EtcdAddr, ";")
	c := getEtcdClient(machines)
	ts := &TestSuit{
		running:    false,
		count:      0,
		latencySum: 0,
	}
	DataNodeWatcher(c, KeyBase, ts)
	fmt.Printf("count: %d, throughput: %.1f avg latency: %.1f ms\n",
		ts.count, float64(ts.count)/float64(ts.end-ts.start)*1e9,
		float64(ts.latencySum)/1e6/float64(ts.count))
}
