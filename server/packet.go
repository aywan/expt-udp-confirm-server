package main

type Packet struct {
	Pid    uint32
	Server uint32
	Num    uint16
	Total  uint16
	Length uint32
	Data   []byte
}

func (p *Packet) getId() uint64 {
	return uint64(p.Pid)<<32 | uint64(p.Server)
}
