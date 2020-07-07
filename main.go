package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/s5i/ruuvi2db/bluetooth"
	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
	"github.com/s5i/ruuvi2db/db/bolt"
	"github.com/s5i/ruuvi2db/db/influx"
	"github.com/s5i/ruuvi2db/db/iowriter"
	"github.com/s5i/ruuvi2db/protocol"
)

var (
	configPathOverride = flag.String("config_path", "", "Path to config file.")
	createConfig       = flag.Bool("create_config", false, "If true, create example config file and exit.")
)

func main() {
	flag.Parse()
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

	buffer := data.NewBuffer(int(cfg.GetGeneral().BufferSize))
	wg.Add(1)
	go func() {
		runBluetooth(ctx, cfg, buffer)
		wg.Done()
	}()

	outputs := setupOutputs(ctx, cfg)
	wg.Add(1)
	go func() {
		refreshLoop(ctx, cfg, buffer, outputs)
		wg.Done()
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
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
	if !cfg.GetGeneral().EnableDebugLogs {
		log.SetOutput(ioutil.Discard)
	}
}

func setupHumanNames(cfg *config.Config) {
	for _, t := range cfg.GetDevices().RuuviTag {
		data.RegisterHumanName(t.GetMac(), t.GetHumanName())
	}
}

func runBluetooth(ctx context.Context, cfg *config.Config, buffer *data.Buffer) {
	if err := bluetooth.Run(ctx, int(cfg.GetBluetooth().HciId), func(addr string, data []byte) {
		res, err := protocol.ParseDatagram(data, addr)
		if err != nil {
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

func setupOutputs(ctx context.Context, cfg *config.Config) []db.Interface {
	outputs := []db.Interface{}

	if cfg.GetGeneral().LogToStdout {
		outputs = append(outputs, iowriter.NewStdout())
	}

	if cfg.GetGeneral().LogToInflux {
		db := influx.NewDB()
		go func() {
			if err := db.RunWithConfig(ctx, cfg); err != nil {
				log.Fatalf("influx: db.RunWithConfig failed: %v", err)
			}
		}()
		outputs = append(outputs, db)
	}

	if cfg.GetGeneral().LogToBolt {
		db := bolt.NewDB()
		go func() {
			if err := db.RunWithConfig(ctx, cfg); err != nil {
				log.Fatalf("bolt: db.RunWithConfig failed: %v", err)
			}
		}()
		outputs = append(outputs, db)
	}

	return outputs
}

func refreshLoop(ctx context.Context, cfg *config.Config, buffer *data.Buffer, outputs []db.Interface) {
	for {
		pts := buffer.PullAll(time.Now())
		sort.Slice(pts, func(i, j int) bool {
			return strings.Compare(pts[i].Name(), pts[j].Name()) < 0
		})
		for _, out := range outputs {
			out.Push(pts)
		}
		select {
		case <-time.After(time.Duration(cfg.GetGeneral().MaxRefreshRateSec) * time.Second):
		case <-ctx.Done():
			return
		}
	}
}
