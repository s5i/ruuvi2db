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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/s5i/ruuvi2db/bluetooth"
	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
	"github.com/s5i/ruuvi2db/db/bolt"
	"github.com/s5i/ruuvi2db/db/iowriter"
	"github.com/s5i/ruuvi2db/http"
	"github.com/s5i/ruuvi2db/protocol"
)

var (
	licenses           = flag.Bool("licenses", false, "When true, print attached licenses and exit.")
	version            = flag.Bool("version", false, "When true, print version and exit.")
	configPathOverride = flag.String("config_path", "", "Path to config file.")
	createConfig       = flag.Bool("create_config", false, "If true, create example config file and exit.")
)

var (
	StdoutDB = "stdout"
	BoltDB   = "bolt"

	//go:embed LICENSES_THIRD_PARTY
	Licenses string
)

func main() {
	flag.Parse()

	if *licenses {
		fmt.Fprintln(os.Stderr, Licenses)
		return
	}

	if *version {
		fmt.Fprintln(os.Stderr, v())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()

	cfgPath := config.Path(*configPathOverride)
	if *createConfig {
		createConfigAndExit(cfgPath)
	}

	cfg := readConfigOrExit(cfgPath)

	setupDebugLogs(cfg)
	setupHumanNames(cfg)

	buffer := data.NewBuffer()
	wg.Add(1)
	go func() {
		runBluetooth(ctx, cfg, buffer)
		wg.Done()
	}()

	dbs := setupDBs(ctx, cfg)
	wg.Add(1)
	go func() {
		refreshLoop(ctx, cfg, buffer, dbs)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		maybeHandleHTTP(ctx, cfg, dbs[BoltDB].(db.Source))
		wg.Done()
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	fmt.Fprintln(os.Stderr, "Caught SIGINT, quitting...")
	cancel()
}

func createConfigAndExit(path string) {
	if err := config.CreateExample(path); err != nil {
		fmt.Fprintf(os.Stderr, "Aborting: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Wrote example config to %s\n", path)
	os.Exit(0)
}

func readConfigOrExit(path string) *config.Config {
	cfg, err := config.Read(path)
	if err == nil {
		return cfg
	}
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Config file %s does not exist.\n", path)
		fmt.Fprintln(os.Stderr, "To create it, run:")
		fmt.Fprintf(os.Stderr, "%s --create_config [--config_path=...]\n", os.Args[0])
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "Aborting: %v\n", err)
	os.Exit(2)
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
		os.Exit(3)
	}
}

func setupDBs(ctx context.Context, cfg *config.Config) map[string]db.Interface {
	dbs := map[string]db.Interface{}

	if cfg.Debug.DumpReadings {
		dbs[StdoutDB] = iowriter.NewStdout()
	}

	db := bolt.NewDB()
	go func() {
		if err := db.RunWithConfig(ctx, cfg); err != nil {
			log.Fatalf("bolt: db.RunWithConfig failed: %v", err)
		}
	}()
	dbs[BoltDB] = db

	return dbs
}

func refreshLoop(ctx context.Context, cfg *config.Config, buffer *data.Buffer, dbs map[string]db.Interface) {
	for {
		pts := buffer.PullAll(time.Now())
		sort.Slice(pts, func(i, j int) bool {
			return strings.Compare(pts[i].Name(), pts[j].Name()) < 0
		})
		for _, db := range dbs {
			db.Push(pts)
		}
		select {
		case <-time.After(cfg.General.LogRate):
		case <-ctx.Done():
			return
		}
	}
}

func maybeHandleHTTP(ctx context.Context, cfg *config.Config, db db.Source) {
	if err := http.Run(ctx, cfg, db); err != nil {
		log.Printf("http.Run failed: %v", err)
		os.Exit(4)
	}
}

func v() string {
	rev := "??????"
	t := "????-??-??T??:??:??Z"
	mod := ""

	bi, ok := debug.ReadBuildInfo()

	if ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" {
				rev = fmt.Sprintf("%s??????", s.Value)[:6]
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
