package tcp

import (
	"github.com/imhinotori/ConnectServer/internal/packets"
	"log"
	"net"
)

func (s *Server) sendCustomList(conn net.Conn) {
	cnt := len(s.svSlice)
	size := 7 + cnt*34 // 1(type)+2(size)+1(head)+1(sub)+2(count)+cnt*(2+32)
	buf := make([]byte, 0, size)

	// PSWMSG_HEAD
	buf = append(buf,
		packets.MsgTypePSW,
		byte(size>>8), byte(size&0xFF), // size HB, LB
		packets.HeadMain, packets.SubCustom, // head, subh
		byte(cnt>>8), byte(cnt&0xFF), // count HB, LB
	)

	// entries: WORD ServerCode (LE), char[32] Name
	for _, sv := range s.svSlice {
		buf = append(buf,
			byte(sv.Code&0xFF), byte(sv.Code>>8),
		)
		nameBuf := make([]byte, 32)
		copy(nameBuf, sv.Name)
		buf = append(buf, nameBuf...)
	}

	log.Printf("[TCP] ▶ customList %d bytes", len(buf))
	conn.Write(buf)
}

// sendList emula CCServerListRecv (PSWMSG_HEAD + BYTE count + entries)
func (s *Server) sendList(conn net.Conn) {
	cnt := byte(len(s.svSlice))
	size := 5 + int(cnt)*3 // 1+2+1+1 + cnt*(2+1)
	buf := []byte{
		packets.MsgTypePSW,
		byte(size >> 8), byte(size & 0xFF), // size HB, LB
		packets.HeadMain, packets.SubBasic, // head, subh
		cnt, // count
	}

	// entries: WORD ServerCode (LE), BYTE UserPercent
	for _, sv := range s.svSlice {
		buf = append(buf,
			byte(sv.Code&0xFF), byte(sv.Code>>8),
			byte(calcPercent(sv.Users)),
		)
	}

	log.Printf("[TCP] ▶ basicList %d bytes", len(buf))
	conn.Write(buf)
}

func calcPercent(users uint16) int {
	const maxUsers = 1000
	p := int(users) * 100 / maxUsers
	if p > 100 {
		p = 100
	}
	return p
}
