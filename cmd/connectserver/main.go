package main

import (
	"github.com/charmbracelet/log"
	"github.com/imhinotori/ConnectServer/internal/configuration"
	"github.com/imhinotori/ConnectServer/internal/tcp"
	"github.com/imhinotori/ConnectServer/internal/udp"
	"go.uber.org/fx"
)

func main() {

	fx.New(
		fx.Provide(
			configuration.Load,
			udp.New,
			tcp.New,
		),
		fx.Invoke(func(
			udpL *udp.Listener,
			tcpS *tcp.Server,
		) {
			log.Info("ConnectServer GO listo")
		}),
	).Run()

	config, err := configuration.Load()
	if err != nil {
		panic(err)
	}

	log.Info("cfg", "cfg", config)

	udpL, err := udp.New(config.ConnectServerInfo.ConnectServerPortUDP, config.Servers)
	if err != nil {
		log.Fatal(err)
	}
	go udpL.Run()

	tcpS, err := tcp.New(config.ConnectServerInfo.ConnectServerPortTCP, config.ConnectServerInfo.MaxIpConnection, config.Servers)
	if err != nil {
		log.Fatal(err)
	}
	go tcpS.Listen()

	log.Info("ConnectServer GO listo")
	select {} // bloquea para siempre
}
