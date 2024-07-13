package pea

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fr-str/httpea/internal/log"
)

type Response struct {
	*http.Response
	Duration      time.Duration
	BodyExports   []Export
	HeaderExports []Export
}

type Client struct {
	*http.Client
	Auth map[string]func() (*Response, error)
	Env  map[string]string
}

func NewClient() *Client {
	c := &Client{
		Client: &http.Client{},
		Auth:   map[string]func() (*Response, error){},
		Env:    map[string]string{},
	}
	c.loadEnv()
	return c
}

func (c *Client) loadEnv() {
	osEnv := os.Environ()
	for _, v := range osEnv {
		b, a, _ := strings.Cut(v, "=")
		c.Env[b] = a
	}
	c.LoadAuto()
}

func (c *Client) Request(file string) (*Response, error) {
	p, err := GetRequestDataFromFile(file, c.Env)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest(p.Method, p.Host, bytes.NewBuffer([]byte(p.Body)))
	if err != nil {
		return nil, err
	}
	r.Header = p.Headers
	q := r.URL.Query()
	for _, v := range strings.Split(p.Query, "\n") {
		k, v, f := strings.Cut(v, "=")
		if !f {
			return nil, fmt.Errorf("invalid query value: '%s'", v)
		}
		q.Add(k, v)
	}
	r.URL.RawQuery = q.Encode()
	log.Debug("[dupa] r.URL.RawQuery: ", r.URL.RawQuery)

	ts := time.Now()
	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}

	log.Debug("request ", p.Method, p.Host, p.Query, p.Headers)

	re := &Response{}
	re.Response = resp
	duration := time.Since(ts).Round(time.Microsecond)
	re.Duration = duration
	re.BodyExports = p.BodyExports
	re.HeaderExports = p.HeaderExports

	if resp != nil && resp.StatusCode == http.StatusForbidden {
		a, ok := c.Auth[strconv.Itoa(http.StatusForbidden)]
		if !ok {
			return re, c.doExports(re)
		}

		resp, err := a()
		if err != nil {
			return nil, err
		}

		if err := c.doExports(resp); err != nil {
			return nil, err
		}

		resp, err = c.Request(file)
		if err != nil {
			return nil, err
		}
		re = resp
	}

	return re, c.doExports(re)
}

func (c *Client) doExports(resp *Response) error {
	// body := m.ReqView.body
	buff := bytes.NewBuffer([]byte{})
	io.Copy(buff, resp.Body)
	resp.Body = io.NopCloser(buff)
	body := buff.String()

	errs := []error{}
	log.Debug("resp.BodyExports: ", len(resp.BodyExports))
	for _, expr := range resp.BodyExports {
		v, err := expr.Expr(body)
		if err != nil {
			errs = append(errs, err)
		}

		c.Env[expr.Name] = v
		log.Debug("expr.EnvName[]: ", expr.Name, v)
	}

	log.Debug("[dupa] resp.HeaderExports: ", len(resp.HeaderExports))
	for _, expr := range resp.HeaderExports {
		v, err := expr.Expr(body)
		if err != nil {
			errs = append(errs, err)
		}
		log.Debug(v, expr.Name)
		c.Env[expr.Name] = resp.Header.Get(v)
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (c *Client) LoadAuto() {
	data := getAutoDataFromFile()
	for _, d := range data {
		c.Auth[d[0]] = func() (*Response, error) {
			return c.Request(d[1])
		}
	}
}
