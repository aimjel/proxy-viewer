package main

import (
	"flag"
	"github.com/aimjel/minecraft"
)

var targetAddr = flag.String("targetAddr", "localhost:25565", "Server which the proxy connects to")

func main() {
	flag.Parse()

	httpSrv := startHttpServer()

	cfg := minecraft.ProxyConfig{
		Status: minecraft.NewStatus(minecraft.Version{
			Protocol: 763,
			Text:     "1.20.1 Proxy Packet Viewer",
		}, 0, "Proxy Packet Viewer"),
		OnReceive: httpSrv.handleReceive,
	}

	go func() {
		if err := httpSrv.srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	if err := cfg.Listen("localhost:25566", *targetAddr); err != nil {
		return
	}

	select {}
}
