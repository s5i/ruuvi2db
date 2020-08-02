package http

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/db"
	"github.com/s5i/ruuvi2db/plot"
)

func Run(ctx context.Context, cfg *config.Config, dbs map[string]db.Interface) error {
	srv := http.Server{}

	listen := cfg.GetHttp().Listen
	if listen == "" {
		if os.Geteuid() == 0 {
			listen = ":80"
		} else {
			listen = ":8080"
		}
	}
	srv.Addr = listen

	srcDB := cfg.GetHttp().SourceDb
	if _, ok := dbs[srcDB]; !ok {
		return fmt.Errorf("source_db (%q) not enabled", srcDB)
	}

	src, ok := dbs[srcDB].(db.Source)
	if !ok {
		return fmt.Errorf("reading from source_db (%q) not supported", srcDB)
	}

	http.HandleFunc("/", mainPage)

	plt := plot.NewPlotter(src)
	http.Handle("/render", plt)

	go func() {
		<-ctx.Done()
		srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("srv.ListenAndServe: %v", err)
	}

	return nil
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	vals := &struct {
		Duration string
		EndTime  string
	}{}

	if x, ok := r.URL.Query()["duration"]; ok && len(x) == 1 {
		vals.Duration = x[0]
	} else {
		vals.Duration = "1h"
	}
	if x, ok := r.URL.Query()["end_time"]; ok && len(x) == 1 {
		vals.EndTime = x[0]
	} else {
		vals.EndTime = "now"
	}

	tmpl.Execute(w, vals)
}

var tmpl = template.Must(template.New("").Parse(`
<body>
  <table border=1>
    <tr><td>
      <form>
        <label for="end_time">End time ("now", Unix epoch in seconds, eg. "1596368396", or offset from time.Now, eg. "-1d"):</label><br>
        <input type="text" id="end_time" name="end_time" value="{{.EndTime}}"><br>
        <label for="duration">Duration (eg. "1d"; unitless means seconds):</label><br>
        <input type="text" id="duration" name="duration" value="{{.Duration}}"><br><br>
        <input type="submit" value="Submit">
      </form>
    </td></tr>
    <tr><td><img src="render?param=temperature&duration={{.Duration}}&end_time={{.EndTime}}"></td></tr>
    <tr><td><img src="render?param=pressure&duration={{.Duration}}&end_time={{.EndTime}}"></td></tr>
    <tr><td><img src="render?param=humidity&duration={{.Duration}}&end_time={{.EndTime}}"></td></tr>
    <tr><td><img src="render?param=battery&duration={{.Duration}}&end_time={{.EndTime}}"></td></tr>
  </table>
</body>
</html>`))
