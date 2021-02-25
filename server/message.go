package main

import (
	"sync/atomic"
	"time"
)

type Message struct {
	Id          uint64
	parts       []bool
	body        []byte
	totalLength int
	createAt    time.Time
	loaded      int32
	isDone      bool
}

func (m *Message) IsDone() bool {
	return m.isDone
}

func (m *Message) Merge(packet *Packet) {

	start := maxPacketDataSize * int(packet.Num)
	end := start + int(packet.Length)

	copy(m.body[start:end], packet.Data[:])

	if packet.Num == packet.Total-1 {
		m.totalLength -= maxPacketDataSize + int(packet.Length)
	}

	num := int(packet.Num)

	if !m.parts[num] {
		atomic.AddInt32(&m.loaded, 1)
		m.parts[num] = true
		if int(m.loaded) == len(m.parts) {
			m.isDone = true
		}
	}
}

func newMessage(packet *Packet) *Message {

	totalLength := int(packet.Total) * maxPacketDataSize
	msg := &Message{
		Id:          packet.getId(),
		parts:       make([]bool, packet.Total),
		body:        make([]byte, totalLength),
		totalLength: totalLength,
		createAt:    time.Now(),
	}

	msg.Merge(packet)

	return msg
}
