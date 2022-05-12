package caldav

import (
	"fmt"
	"net/http"

	"github.com/emersion/go-webdav"
)

type RedirectionTraverser struct {
	user, password string
}

func NewRedirectionTraverser() *RedirectionTraverser {
	return &RedirectionTraverser{}
}

func (tr *RedirectionTraverser) SetAuth(user, password string) {
	tr.user = user
	tr.password = password
}

func (tr *RedirectionTraverser) GetLastLocation(method string, url string) (string, error) {
	lastLocation := ""

	var client webdav.HTTPClient

	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			lastLocation = req.URL.String()
			return nil
		},
	}

	if tr.user != "" && tr.password != "" {
		client = webdav.HTTPClientWithBasicAuth(client, tr.user, tr.password)
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}

	_, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed resolving .well-known: %w", err)
	}

	if lastLocation == "" {
		return "", fmt.Errorf("server provided no Location header")
	}

	return lastLocation, nil
}
