package main

import (
	"log"
	"sync"
)

const workerCount = 8

type pktSndChan chan<- *Packet
type pktRcvChan <-chan *Packet
type msgSndChan chan<- *Message
type msgRcvChan chan<- *Message

var messages = sync.Map{} //make(map[uint64]*Message)
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
		var msg *Message
		v, ok := messages.Load(id)
		if ok {
			msg = v.(*Message)
			msg.Merge(packet)
		} else {
			messageMtx.Lock()
			v, ok := messages.Load(id)
			if ok {
				msg = v.(*Message)
				msg.Merge(packet)
			} else {
				msg = newMessage(packet)
				log.Printf("new packet %d\n", msg.Id)
				messages.Store(id, msg)
			}
			messageMtx.Unlock()
		}

		if msg.IsDone() {
			messageMtx.Lock()
			log.Printf("paket done %d\n", msg.Id)
			messages.Delete(id)

			messageMtx.Unlock()
		}
	}
	wg.Done()
}
