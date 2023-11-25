package util

import (
	"fmt"
	"io"
	"net/http"
)

func HTTPResponseError(resp *http.Response) error {
	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("url: %s - status: %d - body:\n%s", resp.Request.URL.String(), resp.StatusCode, string(b))
}
