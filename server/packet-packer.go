package main

import (
	"sync"
)

const workerCount = 8

type pktSndChan chan<- *Packet
type pktRcvChan <-chan *Packet
type msgSndChan chan<- *Message
type msgRcvChan chan<- *Message

var messages = make(map[uint64]*Message)
var messageMtx = sync.Mutex{}

func startPool() (pktSndChan, *sync.WaitGroup) {
	ch := make(chan *Packet)

	wg := &sync.WaitGroup{}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go packWorker(ch, wg)
	}

	return ch, wg
}

func packWorker(ch pktRcvChan, wg *sync.WaitGroup) {
	for packet := range ch {
		id := packet.getId()
		msg, ok := messages[id]
		if ok {
			msg.Merge(packet)
		} else {
			messageMtx.Lock()
			msg, ok := messages[id]
			if ok {
				msg.Merge(packet)
			} else {
				msg := newMessage(packet)
				messages[id] = msg
			}
			messageMtx.Unlock()
		}
		msg = messages[id]
		if msg.IsDone() {
			messageMtx.Lock()
			delete(messages[id])

			messageMtx.Unlock()
		}
	}
	wg.Done()
}
