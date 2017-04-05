package mcpeping

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	rakNetMagic = []byte{0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78}
)

// Status displays basic statistics for an MCPE server.
type Status struct {
	// The MOTD of the server.
	Description string

	// A human-friendly name for the protocol this server purports to support.
	ProtocolVersion string

	// The protocol ID this server purports to support.
	ProtocolID int

	// How many players this server purportedly contains.
	PlayersOnline int

	// How many players maximum this server reports.
	PlayersMax int
}

// Fetch tries to ping an MCPE server.
func Fetch(host string) (*Status, error) {
	conn, err := net.Dial("udp", host)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(500 * time.Millisecond))

	uc := conn.(*net.UDPConn)
	// generate and send the ping packet
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	r1 := rnd.Int63()
	r2 := rnd.Int63()

	var unconnectedPingPacket bytes.Buffer
	unconnectedPingPacket.WriteByte(0x01)
	binary.Write(&unconnectedPingPacket, binary.BigEndian, r1)
	unconnectedPingPacket.Write(rakNetMagic)
	binary.Write(&unconnectedPingPacket, binary.BigEndian, r2)
	packetBytes := unconnectedPingPacket.Bytes()

	if _, err := uc.Write(packetBytes); err != nil {
		return nil, err
	}

	buf := make([]byte, 2048)
	read, err := uc.Read(buf)
	if err != nil {
		return nil, err
	}

	if read < 1 || buf[0] != 0x1c {
		return nil, errors.New("Invalid ping response")
	}

	return deserialize(buf, read)
}

func deserialize(buf []byte, read int) (*Status, error) {
	parts := strings.Split(string(buf[19+len(rakNetMagic):read]), ";")
	pid, err := strconv.Atoi(parts[2])
	po, err := strconv.Atoi(parts[4])
	pm, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, err
	}

	return &Status{
		Description:     parts[1],
		ProtocolVersion: parts[3],
		ProtocolID:      pid,
		PlayersOnline:   po,
		PlayersMax:      pm,
	}, nil
}
