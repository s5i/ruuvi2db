package storage

import (
	"fmt"
	"net/http"
)

func singleStringParam(r *http.Request, p string) (string, error) {
	x, ok := r.URL.Query()[p]
	if !ok {
		return "", fmt.Errorf("%q not specified", p)
	}

	if len(x) != 1 {
		return "", fmt.Errorf("%q specified multiple times", p)
	}

	return x[0], nil
}
