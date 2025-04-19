package tcp

import (
	"encoding/binary"
	"github.com/charmbracelet/log"
	"github.com/imhinotori/ConnectServer/internal/common"
	"github.com/imhinotori/ConnectServer/internal/configuration"
	"github.com/imhinotori/ConnectServer/internal/packets"
	"net"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	listen   net.Listener
	ipCount  sync.Map // ip->int
	maxPerIP int
	svSlice  []common.ServerInfo
}

func New(port, maxIP int, servers map[string]configuration.Server) (*Server, error) {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

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

	return &Server{listen: l, maxPerIP: maxIP, svSlice: svrs}, nil
}

func (s *Server) Listen() {
	for {
		conn, _ := s.listen.Accept()
		ip, portStr, _ := net.SplitHostPort(conn.RemoteAddr().String())
		log.Printf("[TCP] nueva conexión %s:%s", ip, portStr)

		// --- sección segura contra nil ---
		var count int
		if v, ok := s.ipCount.Load(ip); ok {
			count = v.(int)
		}

		if count >= s.maxPerIP {
			log.Printf("[TCP] %s excede MaxIP (%d); cerrada", ip, s.maxPerIP)
			conn.Close()
			continue
		}
		s.ipCount.Store(ip, count+1)

		// deshabilitamos Nagle para que no agrupe ni retrase
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetNoDelay(true)
			conn = tcpConn
		}

		log.Printf("[TCP] nueva conexión %s:%s", ip, portStr)

		go s.handle(conn, ip)
	}
}

// handle gestiona la conexión de un cliente MU 0.97k
func (s *Server) handle(conn net.Conn, ip string) {
	defer conn.Close()

	buf := make([]byte, 512)

	// 2) Intento de handshake de versión (timeout 1s)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			log.Printf("[TCP] sin handshake de versión, continuo")
		} else {
			log.Printf("[TCP] error en handshake: %v", err)
			return
		}
	} else if n >= 5 && buf[0] == 0xC1 && buf[2] == packets.HeadVer {
		maj, min := buf[3], buf[4]
		log.Printf("[TCP] versión cliente %d.%d", maj, min)
	} else {
		log.Printf("[TCP] paquete inesperado (%d bytes): % X", n, buf[:n])
	}
	conn.SetReadDeadline(time.Time{}) // quita deadline

	// 3) Envío de Init
	initPkt := []byte{0xC1, 0x05, packets.HeadInit, 0xFD, 0x00}
	log.Printf("[TCP] enviando Init % X", initPkt)
	conn.Write(initPkt)

	for {
		buf := make([]byte, 512)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		p := buf[:n]
		if len(p) < 4 || p[0] != 0xC1 || p[2] != packets.HeadMain {
			continue
		}
		switch p[3] {
		case packets.SubBasic:
			log.Printf("[TCP] petición BASIC list")
			s.sendCustomList(conn)
			s.sendList(conn)
		case packets.SubCustom:
			log.Printf("[TCP] petición CUSTOM list")
			s.sendCustomList(conn)
		case 0x03: // INFO
			if len(p) >= 6 {
				code := binary.LittleEndian.Uint16(p[4:6])
				log.Printf("[TCP] petición INFO server %d", code)
				s.sendServerInfo(conn, code)
			}
		default:
			log.Printf("[TCP] subcódigo desconocido: 0x%02X", p[3])
		}
	}
}

func (s *Server) sendServerInfo(c net.Conn, code uint16) {
	log.Printf("[TCP] enviando info de servidor %d", code) // <-- log
	for _, sv := range s.svSlice {
		if sv.Code == code {
			ipBytes := make([]byte, 16)
			copy(ipBytes, sv.IP)
			out := []byte{0xC1, 23, packets.HeadMain, packets.SubInfo}
			out = append(out, ipBytes...)
			out = append(out, byte(sv.Port&0xFF), byte(sv.Port>>8))
			c.Write(out)
			return
		}
	}
}
