package pea

import (
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/fr-str/httpea/internal/log"
	"github.com/fr-str/httpea/internal/util"
	"github.com/fr-str/httpea/internal/util/env"
)

type Pea struct {
	Host          string
	Headers       map[string][]string
	Query         string
	Body          string
	Method        string
	HeaderExports []Export
	BodyExports   []Export
}

const (
	HOST    = "[host]"
	HEADERS = "[headers]"
	QUERY   = "[query]"
	BODY    = "[body]"
	METHOD  = "[method]"
)

// regexes
var (
	regComment  = regexp.MustCompile(`#.*\n`)
	regEnvVar   = regexp.MustCompile(`\$\{.+\}`)
	RegCategory = regexp.MustCompile(`(?m)^\[\w+\]`)
)

func GetRequestDataFromFile(file string, env map[string]string) (Pea, error) {
	d := Pea{}
	b, err := os.ReadFile(util.GetPeaFilePath(file))
	if err != nil {
		return d, err
	}

	s := string(b)
	s = regComment.ReplaceAllString(s, "")
	s = ResolveEnvVars(s, env)
	fields := getFields(s)
	d.Host = fields[HOST]
	d.Query = fields[QUERY]
	d.Body = fields[BODY]
	d.Method = fields[METHOD]
	if d.Method == "" {
		d.Method = http.MethodGet
	}
	d.Headers = map[string][]string{}
	for _, v := range strings.Split(fields[HEADERS], "\n") {
		b, a, f := strings.Cut(v, ":")
		if !f {
			// log.Debug("BAD CUT", v)
			continue
		}
		d.Headers[b] = cleanSlice(strings.Split(a, ","))
	}
	d.parseExports(s)

	return d, nil
}

func getAutoDataFromFile() [][2]string {
	data := util.ReadPeaFile("auto_pea")
	matches := RegCode.FindAllStringSubmatch(data, -1)
	out := make([][2]string, 0, len(matches))
	for _, matche := range matches {
		out = append(out, [2]string{matche[1], matche[2]})
	}

	return out
}

func ResolveEnvVars(s string, en map[string]string) string {
	s = regEnvVar.ReplaceAllStringFunc(s, func(b string) string {
		name := b[2 : len(b)-1]
		v, ok := en[name]
		if !ok {
			v = env.Get(name, b)
		}
		log.Debug(v)
		if v == "" {
			return b
		}
		return v
	})
	return s

}

func cleanSlice(s []string) []string {
	for i, v := range s {
		s[i] = strings.TrimSpace(v)
	}
	return s
}

func getFields(s string) map[string]string {
	m := map[string]string{}
	locs := RegCategory.FindAllStringIndex(s, -1)
	for i, idx := range locs {
		start := idx[1]
		end := len(s) - 1
		if i != len(locs)-1 {
			end = locs[i+1][0]
		}
		key := strings.ToLower(s[idx[0]:idx[1]])
		m[key] = strings.TrimSpace(s[start:end])
	}
	return m
}
