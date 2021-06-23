package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(os.Args[1], nats.Name("JS sub test"), nats.UserCredentials(os.Args[2]))
	defer nc.Close()
	if err != nil {
		fmt.Printf("nats connect: %v\n", err)
		return
	}
	js, err := nc.JetStream(nats.APIPrefix("from.test.API"))
	if err != nil {
		fmt.Printf("JetStream: %v\n", err)
		if js == nil {
			return
		}
	}
	s, err := js.PullSubscribe("test", "DUR", nats.Bind("aggregate", "DUR"))
	if err != nil {
		fmt.Printf("PullSubscribe: %v\n", err)
		return
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("starting\n")
	for {
		select {
		case <-shutdown:
			return
		default:
			if m, err := s.Fetch(1, nats.MaxWait(time.Second)); err != nil {
				fmt.Println(err)
			} else {

				if meta, err := m[0].Metadata(); err == nil {
					fmt.Printf("%+v\n", meta)
				}
				fmt.Println(string(m[0].Data))

				if err := m[0].Ack(); err != nil {
					fmt.Printf("ack error: %+v\n", err)
				}
			}
		}
	}
}
