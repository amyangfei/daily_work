package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var EtcdAddr = flag.String("m", "http://localhost:4001", "string array of etcd address, seperated by ';'")

var KeyBase string = "/update/t"
var StopKey string = "/update/s/stop"

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
	stop       bool
	start      int64 // start time
	end        int64 // end time
	count      int64 // update count
	failcount  int64 // None response count
	latencySum int64 // sum of each update operation's latency
}

func getEtcdClient(machines []string) *etcd.Client {
	return etcd.NewClient(machines)
}

// return true to stop the watcher
func processNotify(data *etcd.Response, stop chan bool, ts *TestSuit) bool {
	if data == nil {
		ts.failcount++
		return true
	}
	if data.Action == EtcdActionUpdate {
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
			ts.latencySum += latency
		}
	}
	return false
}

// return true to stop the watcher, return false to end the monitor.
func UpdateWatcher(machines []string, prefix string, ts *TestSuit) bool {
	c := getEtcdClient(machines)
	receiver := make(chan *etcd.Response)
	stop := make(chan bool)

	go func() {
		if _, err := c.Watch(prefix, 0, true, receiver, stop); err != nil {
			ts.end = time.Now().UnixNano()
			ts.stop = true
		}
	}()

	for {
		select {
		case data := <-receiver:
			if processNotify(data, stop, ts) {
				// omit clean work for benchmark accuracy
				return true
			}
		case <-time.After(time.Second * 1):
			if ts.stop {
				stop <- true
				return false
			}
		}
	}
}

func StopWatcher(machines []string, prefix string, ts *TestSuit) {
	c := getEtcdClient(machines)
	_, err := c.Watch(prefix, 0, false, nil, nil)
	if err != nil {
		fmt.Printf("stop watcher return with error: %v", err)
	}
	ts.end = time.Now().UnixNano()
	ts.stop = true
}

func HandleSignal(ts *TestSuit) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM,
		syscall.SIGINT, syscall.SIGSTOP)

	s := <-c
	switch s {
	case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
		ts.stop = true
		ts.end = time.Now().UnixNano()
		Output(ts)
		os.Exit(0)
	default:
		return
	}
}

func Output(ts *TestSuit) {
	fmt.Printf(
		"count: %d, fail count: %d, throughput: %.1f avg latency: %.1f ms\n",
		ts.count, ts.failcount, float64(ts.count)/float64(ts.end-ts.start)*1e9,
		float64(ts.latencySum)/1e6/float64(ts.count))
}

func main() {
	machines := strings.Split(*EtcdAddr, ";")
	ts := &TestSuit{
		running:    false,
		stop:       false,
		count:      0,
		failcount:  0,
		latencySum: 0,
	}

	go HandleSignal(ts)

	defer Output(ts)

	go StopWatcher(machines, StopKey, ts)

	var stop bool = true
	for stop {
		stop = UpdateWatcher(machines, KeyBase, ts)
	}
}
