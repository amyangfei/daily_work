package main

import (
	"github.com/coreos/go-etcd/etcd"
)

func prepare_clients(addr string, n int) []*etcd.Client {
	clis := make([]*etcd.Client, 0)
	for i := 0; i < n; i++ {
		clis = append(clis, etcd.NewClient([]string{addr}))
	}
	return clis
}
