package main

import (
	"flag"
	"fmt"
	"github.com/amyangfei/redlock-go/redlock"
	"github.com/coreos/go-etcd/etcd"
	"strings"
	"time"
)

var num_clients = flag.Int("c", 100, "Number of concurrent clients")
var num_requests = flag.Int("n", 5, "Requests per client to perform")
var etcd_addr = flag.String("t", "http://localhost:4001", "Address of target etcd to write to")
var write_key = flag.String("k", "dfltkey", "Write key to etcd")
var write_val = flag.String("v", "dfltval", "Write value to etcd")
var key_ttl = flag.Uint64("e", 300, "ttl of write key in seconds")
var redis_addr = flag.String("r", "tcp://127.0.0.1:6379;tcp://127.0.0.1:6380;tcp://127.0.0.1:6381", "multi redis master addr")

func writer(addrs []string, cli *etcd.Client, count int, back chan int) {
	lock, err := redlock.NewRedLock(addrs)
	if err != nil {
		panic(err)
	}
	lock.SetRetryCount(1000)

	incr := 0
	for i := 0; i < count; i++ {
		expiry, err := lock.Lock("foo", 1000)
		if err != nil {
			fmt.Println(err)
		} else {
			if expiry > 500 {
				if _, err := cli.Set(*write_key, *write_val, *key_ttl); err != nil {
					fmt.Printf("etcd set error: %v\n", err)
				}

				incr += 1
				lock.UnLock()
			}
		}
	}
	back <- incr
}

func main() {
	flag.Parse()

	addrs := strings.Split(*redis_addr, ";")
	clis := prepare_clients(*etcd_addr, *num_clients)
	c := make(chan int, *num_clients)
	start := time.Now().UnixNano()
	for _, cli := range clis {
		go writer(addrs, cli, *num_requests, c)
	}
	for i := 0; i < *num_clients; i++ {
		<-c
	}
	end := time.Now().UnixNano()
	fmt.Printf("time elapsed %d ms\n", (end-start)/1e6)
}
