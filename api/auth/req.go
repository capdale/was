package authapi

import (
	"io"
	"net/http"
)

func GetBody(client *http.Client, endpoint string) (b []byte, err error) {
	resp, err := client.Get(endpoint)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	return
}
