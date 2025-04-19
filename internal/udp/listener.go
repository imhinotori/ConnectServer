package udp

import (
	"encoding/binary"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/imhinotori/ConnectServer/internal/common"
	"github.com/imhinotori/ConnectServer/internal/configuration"
	"net"
	"time"
)

type Listener struct {
	addr   *net.UDPAddr
	socket *net.UDPConn
	svMap  map[uint16]*common.ServerInfo
}

func New(port int, servers map[string]configuration.Server) (*Listener, error) {
	svrs := make([]common.ServerInfo, 0, len(servers))
	for _, s := range servers {
		server := common.ServerInfo{
			Code:  s.Code,
			Name:  s.Name,
			IP:    s.Address,
			Port:  uint16(s.Port),
			Show:  !s.Hidden,
			Users: 0,
		}
		svrs = append(svrs, server)
	}

	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	socket, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	m := make(map[uint16]*common.ServerInfo)
	for i, s := range svrs {
		m[s.Code] = &svrs[i]
	}

	return &Listener{
		addr:   addr,
		socket: socket,
		svMap:  m,
	}, nil
}

func (l *Listener) Run() {
	buf := make([]byte, 32)
	for {
		n, _, _ := l.socket.ReadFromUDP(buf)
		if n < 7 || buf[0] != 0xC1 || buf[2] != 0x01 {
			continue
		}

		code := binary.LittleEndian.Uint16(buf[3:5])
		users := binary.LittleEndian.Uint16(buf[5:7])

		if s, ok := l.svMap[code]; ok {
			s.Users = users
			log.Printf("[HB] server=%d users=%d", code, users) // <- nuevo log
		}
	}
}

// heartbeat opcional para depurar
func (l *Listener) DumpEach(d time.Duration) {
	t := time.NewTicker(d)
	for range t.C {
		for _, s := range l.svMap {
			log.Infof("[%d] %s users=%d", s.Code, s.Name, s.Users)
		}
	}
}
