package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/s5i/ruuvi2db/bluetooth"
	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
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
	ctx := context.Background()

	cfgPath := config.Path(*configPathOverride)
	if *createConfig {
		createConfigAndExit(cfgPath)
	}

	cfg := readConfigOrExit(cfgPath)

	setupDebugLogs(cfg)
	setupHumanNames(cfg)

	buffer := data.NewBuffer(int(cfg.GetGeneral().BufferSize))
	go runBluetooth(ctx, cfg, buffer)

	outputs := setupOutputs(ctx, cfg)
	go refreshLoop(ctx, cfg, buffer, outputs)

	select {}
}

func createConfigAndExit(path string) {
	if err := config.CreateExample(path); err != nil {
		fmt.Printf("Aborting: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote example config to %s\n", path)
	os.Exit(0)
}

func readConfigOrExit(path string) *config.Config {
	cfg, err := config.Read(path)
	if err == nil {
		return cfg
	}
	if os.IsNotExist(err) {
		fmt.Printf("Config file %s does not exist.\n", path)
		fmt.Println("To create it, run:")
		fmt.Printf("%s --create_config [--config_path=...]\n", os.Args[0])
		os.Exit(2)
	}
	fmt.Printf("Aborting: %v\n", err)
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
		fmt.Println("Can't run bluetooth; please grant necessary capabilities:")
		fmt.Printf(`$ sudo setcap "cap_net_raw,cap_net_admin=ep" "$(which %s)"`+"\n", os.Args[0])
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
