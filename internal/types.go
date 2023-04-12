package internal

import "net"

type e131Packet struct {
	root  *e131PacketRoot
	frame *e131PacketFrame
	dmp   *e131PacketDMP
}

type e131PacketRoot struct {
	PreambleSize  []byte // 2 bytes Define RLP Preamble Size.
	PostambleSize []byte // 2 bytes RLP Post-amble Size.
	PacketID      []byte // 12 bytes Identifies this packet as E1.17
	FlagsLength   []byte // 2 bytes Protocol flags and length
	Vector        []byte // 4 bytes
	CID           []byte // 16 bytes
}
type e131PacketFrame struct {
	FlagsLength []byte // 2 bytes Protocol flags and length
	Vector      []byte // 4 bytes
	SourceName  []byte // 64 bytes
	Priority    byte   // 1 byte
	SyncAddr    []byte // 2 bytes
	SeqNum      byte   // 1 byte
	Options     byte   // 1 byte
	Universe    []byte // 2 bytes
}
type e131PacketDMP struct {
	FlagsLength       []byte // 2 bytes Protocol flags and length
	Vector            byte   // 1 byte
	AddrType          byte   // 1 byte
	FirstPropertyAddr []byte // 2 bytes
	AddrInc           []byte // 2 bytes
	PropertyValCount  []byte // 2 bytes
	PropertyVal       []byte // 1-513 bytes
}

type conf struct {
	Input  confInput            `yaml:"input"`
	Output map[string]confLight `yaml:"output"`
}

type confInput struct {
	IP       net.IP `yaml:"ip"`
	Port     int    `yaml:"port"`
	Universe int    `yaml:"universe"`
}

type confLight struct {
	MACAddress string `yaml:"mac"`
	UUID       string `yaml:"uuid"`
	RedByte    int    `yaml:"redByte"`
	GreenByte  int    `yaml:"greenByte"`
	BlueByte   int    `yaml:"blueByte"`
}
