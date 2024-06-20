package client

import (
	"bytes"
	"net/http"
	"time"

	"github.com/fr-str/httpea/internal/log"
	"github.com/fr-str/httpea/pkg/pea"
)

var client = &http.Client{}

type Response struct {
	*http.Response
	Duration      time.Duration
	BodyExports   []pea.Export
	HeaderExports []pea.Export
}

func NewRequest(file string, d pea.Pea) (*Response, error) {
	r, err := http.NewRequest(d.Method, d.Host, bytes.NewBuffer([]byte(d.Body)))
	if err != nil {
		return nil, err
	}
	r.Header = d.Headers

	ts := time.Now()
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	duration := time.Since(ts).Round(time.Microsecond)
	log.Debug("request ", d.Method, d.Host, d.Query, d.Headers)
	re := &Response{}
	re.Duration = duration
	re.Response = resp
	re.BodyExports = d.BodyExports
	re.HeaderExports = d.HeaderExports

	return re, nil
}
