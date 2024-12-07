package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/s5i/goutil/shutdown"
	"github.com/s5i/goutil/version"
	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/http"
	"github.com/s5i/ruuvi2db/licenses"
	"github.com/s5i/ruuvi2db/reader/bluetooth"
	"github.com/s5i/ruuvi2db/reader/protocol"
	"github.com/s5i/ruuvi2db/storage/database/bolt"
)

var (
	fLicenses           = flag.Bool("licenses", false, "When true, print attached licenses and exit.")
	fVersion            = flag.Bool("version", false, "When true, print version and exit.")
	fConfigPathOverride = flag.String("config_path", "", "Path to config file.")
	fCreateConfig       = flag.Bool("create_config", false, "If true, create example config file and exit.")
	fAllowRewriteDB     = flag.Bool("allow_rewrite_db", false, "Allows database to be rewritten to a newer schema.")
)

func main() {
	flag.Parse()

	if *fLicenses {
		fmt.Fprintln(os.Stderr, licenses.Merged())
		os.Exit(0)
	}

	if *fVersion {
		fmt.Fprintln(os.Stderr, version.Get())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go shutdown.OnSignal(os.Interrupt, cancel)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	cfgPath := config.Path(*fConfigPathOverride)
	if *fCreateConfig {
		createConfig(cfgPath)
	}

	cfg := readConfig(cfgPath)

	setupDebugLogs(cfg)
	setupHumanNames(cfg)

	buffer := data.NewBuffer(cfg.General.MaxDatapointStaleness)
	db := bolt.New()

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
		log.SetOutput(io.Discard)
	}
}

func setupHumanNames(cfg *config.Config) {
	for _, t := range cfg.Devices.RuuviTag {
		data.RegisterHumanName(t.MAC, t.HumanName)
	}
}

func runBluetooth(ctx context.Context, cfg *config.Config, buffer *data.Buffer) {
	if err := bluetooth.Run(ctx, func(addr string, mfID uint16, datagram []byte) {
		res, err := protocol.ParseDatagram(mfID, datagram, addr)
		if err != nil {
			return
		}
		if !(data.HasHumanName(res.Address) || cfg.General.LogUnknownDevices) {
			return
		}
		buffer.Push(res)
	}, cfg.Bluetooth.WatchdogTimeout); err != nil {
		log.Printf("bluetooth.Run failed: %v", err)
		fmt.Fprintln(os.Stderr, "Can't run bluetooth; please grant necessary capabilities:")
		fmt.Fprintf(os.Stderr, `$ sudo setcap "cap_net_raw,cap_net_admin=ep" "$(which %s)"`+"\n", os.Args[0])
		os.Exit(exitRunBluetoothFailed)
	}
}

func runDB(ctx context.Context, cfg *config.Config, db *bolt.DB) {
	if err := db.Run(ctx, &bolt.Config{
		Path:            cfg.Database.Path,
		RetentionWindow: cfg.Database.RetentionWindow,
		AllowRewrite:    *fAllowRewriteDB,
	}); err != nil {
		log.Printf("db.Run failed: %v", err)
		os.Exit(exitRunDBFailed)
	}
}

func refreshLoop(ctx context.Context, cfg *config.Config, buffer *data.Buffer, db *bolt.DB) {
	for {
		if cfg.Debug.DumpReadings {
			buffer.Print()
		}
		ts := time.Now().Truncate(cfg.General.LogRate)
		data := buffer.PullAll()
		for i := range data {
			data[i].Timestamp = ts
		}
		db.PushPoints(data)

		select {
		case <-time.After(cfg.General.LogRate):
		case <-ctx.Done():
			return
		}
	}
}

func runHTTP(ctx context.Context, cfg *config.Config, db *bolt.DB) {
	if err := http.Run(ctx, cfg, db); err != nil {
		log.Printf("http.Run failed: %v", err)
		os.Exit(exitRunHTTPFailed)
	}
}

const (
	exitOK = iota
	exitCreateConfigFailed
	exitReadConfigFailed
	exitRunBluetoothFailed
	exitRunHTTPFailed
	exitRunDBFailed
	exitDBRewriteFailed
)
