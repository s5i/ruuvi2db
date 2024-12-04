#!/bin/bash

cd $(dirname "$(readlink -f "$0")")
rm -rf embed

go install github.com/google/go-licenses@latest 2>/dev/null
go-licenses save github.com/s5i/ruuvi2db/... --save_path="${PWD}/embed" 2>/dev/null

for pkg in \
  "github.com/c3js/c3/LICENSE https://raw.githubusercontent.com/c3js/c3/refs/heads/master/LICENSE" \
  "github.com/d3/d3/LICENSE https://raw.githubusercontent.com/d3/d3/refs/heads/main/LICENSE" \
; do
  set -- ${pkg}
  mkdir -p embed/$(dirname ${1})
  curl ${2} > embed/${1} 2>/dev/null
done