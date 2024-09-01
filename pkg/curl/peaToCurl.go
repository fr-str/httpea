package curl

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fr-str/httpea/internal/util"
	"github.com/fr-str/httpea/pkg/pea"
)

func PeaToCurl(p pea.Pea) (string, error) {
	buf := strings.Builder{}
	w := buf.WriteString
	w("curl")
	w(" ")
	w(`"`)
	w(p.Host)
	if p.Query != "" {
		w(joinQuery(p.Query))
	}
	w(`"`)

	w(" ")
	w("-X " + strings.ToUpper(p.Method))

	for k, v := range p.Headers {
		w(" \\\n\t")
		w(fmt.Sprintf("-H '%s: %s'", k, strings.Join(v, ",")))
	}

	if len(p.Body) > 0 {
		w(" \\\n\t")
		w("-d ")
		w("'")
		w(string(util.Must(json.Marshal(json.RawMessage(p.Body)))))
		w("'")
	}

	return buf.String(), nil
}

func joinQuery(s string) string {
	return "?" + strings.Join(strings.Split(s, "\n"), "&")
}
