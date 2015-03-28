package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"time"
)

var num_clients = flag.Int("c", 100, "Number of concurrent clients")
var num_requests = flag.Int("n", 5, "Requests per client to perform")
var etcd_addr = flag.String("t", "http://localhost:4001", "Address of target etcd to write to")
var write_key = flag.String("k", "dfltkey", "Write key to etcd")
var write_val = flag.String("v", "dfltval", "Write value to etcd")
var key_ttl = flag.Uint64("e", 300, "ttl of write key in seconds")

func write_etcd(clis []*etcd.Client, req int) {
	start := time.Now().UnixNano()
	for i := 0; i < req; i++ {
		for _, cli := range clis {
			if _, err := cli.Set(*write_key, *write_val, *key_ttl); err != nil {
				fmt.Printf("etcd set error: %v\n", err)
			}
		}
	}
	end := time.Now().UnixNano()
	fmt.Printf("time elapsed %d ms\n", (end-start)/1e6)
}

func main() {
	flag.Parse()

	clis := prepare_clients(*etcd_addr, *num_clients)
	write_etcd(clis, *num_requests)
}
