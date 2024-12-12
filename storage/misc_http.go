package storage

import (
	"fmt"
	"net/http"
)

func singleStringParam(r *http.Request, p string) (string, bool, error) {
	x, ok := r.URL.Query()[p]
	if !ok {
		return "", false, nil
	}

	if len(x) != 1 {
		return "", false, fmt.Errorf("%q specified multiple times", p)
	}

	return x[0], true, nil

}
