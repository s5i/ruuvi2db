package licenses

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

func Merged() string {
	var files []string
	fs.WalkDir(embedded, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		files = append(files, strings.TrimPrefix(path, "embed/"))
		return nil
	})

	sort.Slice(files, func(i, j int) bool {
		if files[i] == "github.com/s5i/ruuvi2db/LICENSE" {
			return true
		}
		if files[j] == "github.com/s5i/ruuvi2db/LICENSE" {
			return false
		}
		return files[i] < files[j]
	})

	var formatted []string

	for _, f := range files {
		var text string
		b, err := embedded.ReadFile("embed/" + f)
		if err != nil {
			text = "<error reading file>"
			continue
		}
		text = string(b)
		formatted = append(formatted, fmt.Sprintf("## %s\n\n%s", f, text))
	}
	return strings.Join(formatted, fmt.Sprintf("\n%s\n\n", strings.Repeat("-", 80)))
}

//go:embed embed
var embedded embed.FS
