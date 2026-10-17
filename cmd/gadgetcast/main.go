// Command gadgetcast runs the Pi-side server: REST API, event push, WebRTC
// signaling, and the embedded Studio UI.
package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/adeelahmad/gadgetcast/internal/config"
	"github.com/adeelahmad/gadgetcast/internal/gadget"
	"github.com/adeelahmad/gadgetcast/internal/server"
	"github.com/adeelahmad/gadgetcast/web"
)

const (
	defaultTLSPort = "8443"
	defaultDevAddr = "127.0.0.1:8080"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gadgetcast:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) > 1 && os.Args[1] == "gadget" {
		return runGadget(os.Args[2:])
	}
	var (
		configPath = flag.String("config", "", "path to the JSON config file (required unless -dev)")
		addrFlag   = flag.String("addr", "", "listen address override (default :8443, or "+defaultDevAddr+" with -dev)")
		dataDir    = flag.String("data", "data", "data directory (pipelines are stored under <data>/pipelines)")
		dev        = flag.Bool("dev", false, "development mode: plain HTTP on localhost, config optional")
		showVer    = flag.Bool("version", false, "print the version and exit")
	)
	flag.Parse()
	if *showVer {
		fmt.Println(server.Version)
		return nil
	}

	logger := log.New(os.Stderr, "gadgetcast ", log.LstdFlags)

	cfg, haveCfg, err := loadConfig(*configPath, *dev)
	if err != nil {
		return err
	}

	pin := ""
	if haveCfg {
		if pin, err = resolvePIN(cfg.PIN.Source); err != nil && !*dev {
			return err
		}
	}
	if pin == "" {
		if !*dev {
			return errors.New("no PIN available from pin.source")
		}
		pin = randomPIN()
		logger.Printf("dev mode: pairing PIN is %s", pin)
	}

	addr := *addrFlag
	if addr == "" {
		addr = listenAddr(cfg, *dev)
	}
	if *dev && !isLoopback(addr) {
		logger.Printf("warning: -dev serves plain HTTP; %s is not a loopback address", addr)
	}

	static, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		return err
	}
	opts := server.Options{
		DataDir:         *dataDir,
		PIN:             pin,
		PINAttemptLimit: cfg.PIN.AttemptLimit,
		Static:          static,
		Logger:          logger,
	}
	if *dev {
		opts.OriginPatterns = []string{"localhost:*", "127.0.0.1:*"}
	}
	srv, err := server.New(opts)
	if err != nil {
		return err
	}

	hs := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ErrorLog:          logger,
	}
	hs.RegisterOnShutdown(srv.Close)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		if *dev {
			logger.Printf("v%s listening on http://%s (dev, data=%s)", server.Version, addr, *dataDir)
			errc <- hs.ListenAndServe()
			return
		}
		logger.Printf("v%s listening on https://%s (data=%s)", server.Version, addr, *dataDir)
		errc <- hs.ListenAndServeTLS(cfg.HTTPS.CertPath, cfg.HTTPS.KeyPath)
	}()

	select {
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}
	logger.Printf("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Close()
	if err := hs.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

// loadConfig reads and validates the config. Outside -dev the config is
// required and must pass config.Validate (HTTPS with cert+key, PIN source, …).
func loadConfig(path string, dev bool) (config.Config, bool, error) {
	if path == "" {
		if !dev {
			return config.Config{}, false, errors.New("-config is required (or use -dev for local UI development)")
		}
		cfg, err := config.Parse([]byte(`{}`))
		return cfg, false, err
	}
	doc, err := os.ReadFile(path)
	if err != nil {
		return config.Config{}, false, fmt.Errorf("read config: %w", err)
	}
	cfg, err := config.Parse(doc)
	if err != nil {
		return config.Config{}, false, err
	}
	if !dev {
		if err := cfg.Validate(); err != nil {
			return config.Config{}, false, err
		}
	}
	return cfg, true, nil
}

// resolvePIN reads the PIN from its configured source: "env:NAME",
// "file:PATH", or a literal value.
func resolvePIN(source string) (string, error) {
	switch {
	case source == "":
		return "", nil
	case strings.HasPrefix(source, "env:"):
		name := strings.TrimPrefix(source, "env:")
		v := strings.TrimSpace(os.Getenv(name))
		if v == "" {
			return "", fmt.Errorf("pin.source: environment variable %s is empty", name)
		}
		return v, nil
	case strings.HasPrefix(source, "file:"):
		b, err := os.ReadFile(strings.TrimPrefix(source, "file:"))
		if err != nil {
			return "", fmt.Errorf("pin.source: %w", err)
		}
		return strings.TrimSpace(string(b)), nil
	default:
		return source, nil
	}
}

func randomPIN() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

func listenAddr(cfg config.Config, dev bool) string {
	if dev {
		return defaultDevAddr
	}
	if cfg.Bind.Mode == config.BindTailscaleOnly && cfg.Bind.Target != "" {
		if _, _, err := net.SplitHostPort(cfg.Bind.Target); err == nil {
			return cfg.Bind.Target
		}
		return net.JoinHostPort(cfg.Bind.Target, defaultTLSPort)
	}
	return ":" + defaultTLSPort
}

func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// runGadget handles "gadgetcast gadget up|down": it creates or removes the
// configfs USB composite device (webcam + mic + speaker). The boot unit
// gadgetcast-gadget.service runs it before the server starts.
func runGadget(args []string) error {
	fl := flag.NewFlagSet("gadget", flag.ContinueOnError)
	root := fl.String("configfs", gadget.DefaultConfigFS, "configfs usb_gadget directory")
	udcDir := fl.String("udc-dir", gadget.DefaultUDCDir, "directory listing USB device controllers")
	noVideo := fl.Bool("no-video", false, "omit the UVC webcam function")
	noAudio := fl.Bool("no-audio", false, "omit the UAC2 mic/speaker function")
	usage := errors.New("usage: gadgetcast gadget up|down [flags]")
	if len(args) == 0 {
		return usage
	}
	cmd := args[0]
	if err := fl.Parse(args[1:]); err != nil {
		return err
	}
	spec := gadget.DefaultSpec()
	spec.Video, spec.Audio = !*noVideo, !*noAudio
	if b, err := os.ReadFile("/proc/device-tree/serial-number"); err == nil {
		if sn := strings.Trim(string(b), "\x00\n "); sn != "" {
			spec.Serial = sn
		}
	}
	switch cmd {
	case "up":
		udc, err := gadget.FindUDC(*udcDir)
		if err != nil {
			return err
		}
		if err := gadget.Up(gadget.OS{}, *root, spec, udc); err != nil {
			return err
		}
		fmt.Printf("USB gadget bound to %s (video=%t audio=%t)\n", udc, spec.Video, spec.Audio)
		return nil
	case "down":
		return gadget.Down(gadget.OS{}, *root, spec)
	default:
		return usage
	}
}
