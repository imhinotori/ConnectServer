package configuration

type Configuration struct {
	ConnectServerInfo ConnectServerInfo `toml:"Connect_Server_Info"`
	Servers           map[string]Server `toml:"Servers"`
}

type ConnectServerInfo struct {
	ConnectServerPortTCP int `toml:"TCP_PORT"`
	ConnectServerPortUDP int `toml:"UDP_PORT"`
	MaxIpConnection      int `toml:"Max_Ip_Connection"`
}

type Server struct {
	Code    uint16 `toml:"Code"`
	Name    string `toml:"Name"`
	Address string `toml:"Address"`
	Port    int    `toml:"Port"`
	Type    string `toml:"Type"`
	Hidden  bool   `toml:"Hidden"`
}
