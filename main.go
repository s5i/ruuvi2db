package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"time"

	"github.com/s5i/ruuvi2db/bluetooth"
	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/database"
	"github.com/s5i/ruuvi2db/http"
	"github.com/s5i/ruuvi2db/protocol"
)

var (
	fLicenses           = flag.Bool("licenses", false, "When true, print attached licenses and exit.")
	fVersion            = flag.Bool("version", false, "When true, print version and exit.")
	fConfigPathOverride = flag.String("config_path", "", "Path to config file.")
	fCreateConfig       = flag.Bool("create_config", false, "If true, create example config file and exit.")
	fRewriteDB          = flag.Bool("rewrite_db", false, "If true, rewrite database and exit.")
)

func main() {
	flag.Parse()

	if *fLicenses {
		fmt.Fprintln(os.Stderr, Licenses)
		return
	}

	if *fVersion {
		fmt.Fprintln(os.Stderr, version())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()

	cfgPath := config.Path(*fConfigPathOverride)
	if *fCreateConfig {
		createConfig(cfgPath)
	}

	cfg := readConfig(cfgPath)

	setupDebugLogs(cfg)
	setupHumanNames(cfg)

	buffer := data.NewBuffer()
	db := database.NewDB()

	if *fRewriteDB {
		if err := db.Rewrite(); err != nil {
			os.Exit(exitDBRewriteFailed)
		}
		os.Exit(exitOK)
	}

	wg.Add(1)
	go func() {
		runBluetooth(ctx, cfg, buffer)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		runDB(ctx, cfg, db)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		refreshLoop(ctx, cfg, buffer, db)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		runHTTP(ctx, cfg, db)
		wg.Done()
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	fmt.Fprintln(os.Stderr, "Caught SIGINT, quitting...")
	cancel()
}

func createConfig(path string) {
	if err := config.CreateExample(path); err != nil {
		fmt.Fprintf(os.Stderr, "Aborting: %v\n", err)
		os.Exit(exitCreateConfigFailed)
	}
	fmt.Fprintf(os.Stderr, "Wrote example config to %s\n", path)
	os.Exit(exitOK)
}

func readConfig(path string) *config.Config {
	cfg, err := config.Read(path)
	if err == nil {
		return cfg
	}
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Config file %s does not exist.\n", path)
		fmt.Fprintln(os.Stderr, "To create it, run:")
		fmt.Fprintf(os.Stderr, "%s --create_config [--config_path=...]\n", os.Args[0])
		os.Exit(exitReadConfigFailed)
	}
	fmt.Fprintf(os.Stderr, "Aborting: %v\n", err)
	os.Exit(exitReadConfigFailed)
	return nil
}

func setupDebugLogs(cfg *config.Config) {
	if !cfg.Debug.DumpBinaryLogs {
		log.SetOutput(ioutil.Discard)
	}
}

func setupHumanNames(cfg *config.Config) {
	for _, t := range cfg.Devices.RuuviTag {
		data.RegisterHumanName(t.MAC, t.HumanName)
	}
}

func runBluetooth(ctx context.Context, cfg *config.Config, buffer *data.Buffer) {
	if err := bluetooth.Run(ctx, int(cfg.Bluetooth.HCIID), func(addr string, datagram []byte) {
		res, err := protocol.ParseDatagram(datagram, addr)
		if err != nil {
			return
		}
		if !(data.HasHumanName(res.Address) || cfg.General.LogUnknownDevices) {
			return
		}
		buffer.Push(*res)
	}); err != nil {
		log.Printf("bluetooth.Run failed: %v", err)
		fmt.Fprintln(os.Stderr, "Can't run bluetooth; please grant necessary capabilities:")
		fmt.Fprintf(os.Stderr, `$ sudo setcap "cap_net_raw,cap_net_admin=ep" "$(which %s)"`+"\n", os.Args[0])
		os.Exit(exitRunBluetoothFailed)
	}
}

func runDB(ctx context.Context, cfg *config.Config, db db) {
	if err := db.Run(ctx, cfg); err != nil {
		log.Printf("db.Run failed: %v", err)
		os.Exit(exitRunDBFailed)
	}
}

func refreshLoop(ctx context.Context, cfg *config.Config, buffer *data.Buffer, db db) {
	for {
		if cfg.Debug.DumpReadings {
			buffer.Print()
		}
		db.Push(buffer.PullAll(time.Now()))

		select {
		case <-time.After(cfg.General.LogRate):
		case <-ctx.Done():
			return
		}
	}
}

func runHTTP(ctx context.Context, cfg *config.Config, db http.DB) {
	if err := http.Run(ctx, cfg, db); err != nil {
		log.Printf("http.Run failed: %v", err)
		os.Exit(exitRunHTTPFailed)
	}
}

func version() string {
	rev := "???????"
	t := "????-??-??T??:??:??Z"
	mod := ""

	bi, ok := debug.ReadBuildInfo()

	if ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" {
				rev = fmt.Sprintf("%s???????", s.Value)[:7]
			}
			if s.Key == "vcs.time" {
				t = s.Value
			}
			if s.Key == "vcs.modified" && s.Value == "true" {
				mod = " (modified)"
			}
		}
	}
	return fmt.Sprintf("rev: %s (%s)%s", rev, t, mod)
}

type db interface {
	Push(points []data.Point)
	Run(ctx context.Context, cfg *config.Config) error
}

//go:embed LICENSES_THIRD_PARTY
var Licenses string

const (
	exitOK = iota
	exitCreateConfigFailed
	exitReadConfigFailed
	exitRunBluetoothFailed
	exitRunHTTPFailed
	exitRunDBFailed
	exitDBRewriteFailed
)
