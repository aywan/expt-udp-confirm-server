package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"log"
	"net"
)

const maxPacketSize = 512
const headerSize = 20
const maxPacketDataSize = maxPacketSize - headerSize

const ok uint16 = 1
const wrongLength uint16 = 2
const wrongCrc uint16 = 4

func main() {
	pc, err := net.ListenPacket("udp", ":18086")
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	ch, wg := startPool()

	for {
		buf := make([]byte, maxPacketSize)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}
		go serve(pc, addr, buf[:n], ch)
	}

	close(ch)
	wg.Wait()
}

func serve(pc net.PacketConn, addr net.Addr, buf []byte, ch pktSndChan) {
	var crc uint32
	msg := &Packet{}
	response := make([]byte, 8)

	reader := bytes.NewReader(buf)

	binary.Read(reader, binary.LittleEndian, &msg.Pid)
	binary.Read(reader, binary.LittleEndian, &msg.Server)
	binary.Read(reader, binary.LittleEndian, &msg.Num)
	binary.Read(reader, binary.LittleEndian, &msg.Total)
	binary.Read(reader, binary.LittleEndian, &crc)
	binary.Read(reader, binary.LittleEndian, &msg.Length)

	//log.Print("receive id=", msg.Pid, "/", msg.Server, " packet=", msg.Num, "/", msg.Total, " length=", msg.Length, " crc=", crc)
	msg.Data = make([]byte, msg.Length)
	n, _ := reader.Read(msg.Data)
	if n < int(msg.Length) {
		binary.LittleEndian.PutUint16(response, wrongLength)
	} else if crc32.ChecksumIEEE(msg.Data) != crc {
		binary.LittleEndian.PutUint16(response, wrongCrc)
	} else {
		ch <- msg
		repCrc := crc % msg.Pid
		binary.LittleEndian.PutUint16(response[0:2], ok)
		binary.LittleEndian.PutUint16(response[2:4], msg.Num)
		binary.LittleEndian.PutUint32(response[4:8], repCrc)
	}
	pc.WriteTo(response, addr)
}
