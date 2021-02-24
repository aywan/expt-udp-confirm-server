package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"log"
	"net"
)

const maxPacketSize = 1400

func main() {
	pc, err := net.ListenPacket("udp", ":18086")
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	for {
		buf := make([]byte, maxPacketSize)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}
		go serve(pc, addr, buf[:n])
	}
}

func serve(pc net.PacketConn, addr net.Addr, buf []byte) {
	var pid, server uint32
	var num, total uint16
	var crc, length uint32

	reader := bytes.NewReader(buf)

	binary.Read(reader, binary.LittleEndian, &pid)
	binary.Read(reader, binary.LittleEndian, &server)
	binary.Read(reader, binary.LittleEndian, &num)
	binary.Read(reader, binary.LittleEndian, &total)
	binary.Read(reader, binary.LittleEndian, &crc)
	binary.Read(reader, binary.LittleEndian, &length)

	log.Print("receive id=", pid, "/", server, " packet=", num, "/", total, " length=", length, " crc=", crc)
	packet := make([]byte, length)
	n, _ := reader.Read(packet)
	if n < int(length) {
		log.Fatal(n, " < ", length)
	}

	if crc32.ChecksumIEEE(packet) != crc {
		log.Fatal("crc not valid")
	}
}
