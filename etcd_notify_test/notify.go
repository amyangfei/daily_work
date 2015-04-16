package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var KeySize = flag.Int("k", 16, "key size in etcd, at least 16")
var NumClis = flag.Int("n", 4, "Number of concurrent clients")
var UpdateFrequency = flag.Float64("r", 1.0, "update frequency of per client")
var EtcdAddr = flag.String("m", "http://localhost:4001", "string array of etcd address, seperated by ';'")
var RunTime = flag.Int("t", 30, "run time in second")

var KeyBase string = "/update/t"
var KeyNum int = 4
var minKeySize = 16
var AttrNames []string
var letters = []rune("abcdefghijklmnopqrstuvwxyz")

type Client struct {
	name string
	cli  *etcd.Client
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func timestampVal() string {
	now := time.Now().UnixNano()
	return fmt.Sprintf("%d", now)
}

func prepareAttrs(clis []*Client, KeySize, key_num int) {
	if KeySize < minKeySize {
		KeySize = minKeySize
	}
	for i := 0; i < key_num; i++ {
		name := randSeq(KeySize - len(KeyBase) - len(clis[0].name) - 2)
		AttrNames = append(AttrNames, name)
	}

	for _, c := range clis {
		for _, aname := range AttrNames {
			key := fmt.Sprintf("%s/%s/%s", KeyBase, c.name, aname)
			if _, err := c.cli.Create(key, timestampVal(), 0); err != nil {
				panic(err)
			}
		}
	}
}

func prepareClients(machines []string, n int) []*Client {
	clis := make([]*Client, 0)
	for i := 0; i < n; i++ {
		format := fmt.Sprintf("%%0%dd", len(fmt.Sprintf("%d", n-1)))
		name := fmt.Sprintf(format, i)
		c := &Client{
			name: name,
			cli:  etcd.NewClient(machines),
		}
		clis = append(clis, c)
	}
	return clis
}

func randomEtcdUpdate(c *Client) {
	randAttr := AttrNames[rand.Intn(len(AttrNames))]
	key := fmt.Sprintf("%s/%s/%s", KeyBase, c.name, randAttr)
	newVal := timestampVal()
	if _, err := c.cli.Update(key, newVal, 0); err != nil {
		panic(err)
	}
}

func cliRoutine(cli *Client, freq float64, stop chan bool) {
	microFreq := int(freq * 1e6)
	ticker := time.NewTicker(time.Microsecond * time.Duration(microFreq))
	for {
		select {
		case <-ticker.C:
			randomEtcdUpdate(cli)
		case s := <-stop:
			if s {
				fmt.Printf("cli %s routine done\n", cli.name)
				return
			}
		}
	}
}

func updateEtcdTester(clis []*Client, freq float64, runTime int) {
	fmt.Printf("start benchmark...\n")
	timer := time.NewTimer(time.Second * time.Duration(runTime))

	stops := make([]chan bool, 0)
	for i := 0; i < len(clis); i++ {
		stops = append(stops, make(chan bool))
	}
	for idx, cli := range clis {
		go cliRoutine(cli, freq, stops[idx])
	}
	<-timer.C
	for _, stop := range stops {
		stop <- true
	}
}

func InitSignal() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM,
		syscall.SIGINT, syscall.SIGSTOP)
	return c
}

func HandleSignal(c chan os.Signal, clis []*Client) {
	s := <-c
	switch s {
	case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
		fmt.Printf("catch signal %v\n", s)
		cleanWork(clis)
		os.Exit(0)
	default:
		return
	}
}

func cleanWork(clis []*Client) {
	fmt.Printf("do clean work...\n")
	for _, c := range clis {
		k := fmt.Sprintf("%s/%s", KeyBase, c.name)
		c.cli.Delete(k, true)
	}
}

func main() {
	flag.Parse()

	machines := strings.Split(*EtcdAddr, ";")
	clis := prepareClients(machines, *NumClis)
	prepareAttrs(clis, *KeySize, KeyNum)

	defer cleanWork(clis)
	signalChan := InitSignal()
	go HandleSignal(signalChan, clis)

	updateEtcdTester(clis, *UpdateFrequency, *RunTime)
}
