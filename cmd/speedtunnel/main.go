package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/SpeedwiT/SpeedTunnel/internal/bot"
	"github.com/SpeedwiT/SpeedTunnel/internal/config"
	"github.com/SpeedwiT/SpeedTunnel/internal/forward"
	"github.com/SpeedwiT/SpeedTunnel/internal/tunnel"
	"github.com/SpeedwiT/SpeedTunnel/internal/utils"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch cmd {
	case "server":
		runServer(os.Args[2:])
	case "client":
		runClient(os.Args[2:])
	case "run":
		runAll()
	case "bot":
		runBot()
	case "version", "-v", "--version":
		fmt.Println("SpeedTunnel", config.Version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Speed Tunnel v` + config.Version + ` - by @Speedw_IT

Usage:
  speedtunnel server  - run kharej server (from config)
  speedtunnel client  - run iran client (from config)
  speedtunnel run     - auto detect role from config and run
  speedtunnel bot     - run telegram bot only

Config: ` + config.GetConfigPath() + `
Github: https://github.com/SpeedwiT/SpeedTunnel
Channel: @Speedw_IT  Support: @SpeedwIT`)
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	_ = fs.Parse(args)
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if len(cfg.Tunnels) == 0 {
		log.Fatal("no tunnels in config. Use menu to create one.")
	}
	// run bot in background
	if cfg.BotToken != "" {
		go bot.New(cfg.BotToken, cfg.BotAdminID).Start()
	}
	for _, t := range cfg.Tunnels {
		if !t.Enabled {
			continue
		}
		if t.Role != "kharej" && t.Role != "server" {
			continue
		}
		t := t
		go runKharejTunnel(t)
	}
	waitSignal()
}

func runClient(args []string) {
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	_ = fs.Parse(args)
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if len(cfg.Tunnels) == 0 {
		log.Fatal("no tunnels in config.")
	}
	if cfg.BotToken != "" {
		go bot.New(cfg.BotToken, cfg.BotAdminID).Start()
	}
	for _, t := range cfg.Tunnels {
		if !t.Enabled {
			continue
		}
		if t.Role != "iran" && t.Role != "client" {
			continue
		}
		t := t
		go runIranTunnel(t)
	}
	waitSignal()
}

func runAll() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if len(cfg.Tunnels) == 0 {
		log.Println("no tunnels configured")
		waitSignal()
		return
	}
	hasIran, hasKharej := false, false
	for _, t := range cfg.Tunnels {
		if !t.Enabled {
			continue
		}
		if t.Role == "iran" || t.Role == "client" {
			hasIran = true
		}
		if t.Role == "kharej" || t.Role == "server" {
			hasKharej = true
		}
	}
	if cfg.BotToken != "" {
		go bot.New(cfg.BotToken, cfg.BotAdminID).Start()
	}
	for _, t := range cfg.Tunnels {
		if !t.Enabled {
			continue
		}
		t := t
		if t.Role == "iran" || t.Role == "client" {
			go runIranTunnel(t)
		} else {
			go runKharejTunnel(t)
		}
	}
	log.Printf("[main] running %d tunnels (iran=%v kharej=%v)", len(cfg.Tunnels), hasIran, hasKharej)
	waitSignal()
}

func runKharejTunnel(t config.TunnelConfig) {
	controlAddr := fmt.Sprintf("0.0.0.0:%d", t.ControlPort)
	// This server also forwards: ListenPort on kharej -> via tunnel -> iran local service
	// So kharej listens on ListenPort for users, and control port for IRAN
	addrInfo := fmt.Sprintf("tunnel %s (%s) control %d forward 0.0.0.0:%d -> iran 127.0.0.1:%d [%s sni=%s]", t.Name, t.ID, t.ControlPort, t.ListenPort, t.RemotePort, t.Transport, t.SNI)
	log.Printf("[kharej] starting %s", addrInfo)

	var muxRef *tunnel.Multiplexer
	server := tunnel.NewServer(controlAddr, t.Secret, t.Transport, t.SNI, func(mux *tunnel.Multiplexer, conn net.Conn) {
		muxRef = mux
		// for kharej: start forwarder that listens on ListenPort and forwards via mux
		fwd, err := forward.NewForwarder(fmt.Sprintf("0.0.0.0:%d", t.ListenPort), mux)
		if err != nil {
			log.Printf("[kharej %s] forward listen failed: %v", t.ID, err)
			return
		}
		log.Printf("[kharej %s] forwarding 0.0.0.0:%d -> tunnel", t.ID, t.ListenPort)
		go fwd.Serve()
		// mux onNew is handled inside forwarder (OpenStream) – no need for extra
		_ = muxRef
	})
	if err := server.ListenAndServe(); err != nil {
		log.Printf("[kharej %s] server error: %v", t.ID, err)
	}
}

func runIranTunnel(t config.TunnelConfig) {
	remoteControl := net.JoinHostPort(t.RemoteAddr, strconv.Itoa(t.ControlPort))
	localService := net.JoinHostPort("127.0.0.1", strconv.Itoa(t.RemotePort))
	log.Printf("[iran] tunnel %s (%s) -> %s local %s [%s sni=%s]", t.Name, t.ID, remoteControl, localService, t.Transport, t.SNI)
	client := tunnel.NewClient(remoteControl, t.Secret, t.Transport, t.SNI, func(mux *tunnel.Multiplexer) {
		mux.SetOnNewConn(func(s *tunnel.Stream) {
			forward.HandleRemoteStream(s, localService)
		})
	})
	client.ConnectWithRetry()
}

func runBot() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	b := bot.New(cfg.BotToken, cfg.BotAdminID)
	b.Start()
}

func waitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("shutting down...")
}

func parsePorts(s string) []int {
	var res []int
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, err := strconv.Atoi(p); err == nil {
			res = append(res, v)
		}
	}
	return res
}

var _ = utils.RandomID
