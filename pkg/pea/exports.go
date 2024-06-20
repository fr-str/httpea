package pea

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/fr-str/httpea/internal/log"
)

var (
	RegExportBody   = regexp.MustCompile(`@exportBody (.+) ([\w|\d]+)`)
	RegExportHeader = regexp.MustCompile(`@exportHeader (.+) ([\w|\d]+)`)
	RegShell        = regexp.MustCompile(`@sh\((.+)\)`)
	RegCode         = regexp.MustCompile(`@(\d{3}) (.+)`)
)

type Export struct {
	Name string
	Expr ExprFunc
}

type ExprFunc func(in string) (string, error)

func (p *Pea) parseExports(s string) {
	matches := RegExportBody.FindAllStringSubmatch(s, -1)
	for _, v := range matches {
		p.BodyExports = append(p.BodyExports, Export{Name: v[2], Expr: determineExpr(v[1])})
	}
	matches = RegExportHeader.FindAllStringSubmatch(s, -1)
	for _, v := range matches {
		p.HeaderExports = append(p.HeaderExports, Export{Name: v[2], Expr: determineExpr(v[1])})
	}

}

func determineExpr(s string) ExprFunc {
	if RegShell.MatchString(s) {
		return shellExpr(RegShell.FindStringSubmatch(s)[1])
	}
	return staticExpr(s)
}

func staticExpr(s string) ExprFunc {
	return func(in string) (string, error) {
		return s, nil
	}
}

func shellExpr(s string) ExprFunc {
	return func(in string) (string, error) {
		cmd := exec.Command("sh", "-c", s)

		cmd.Env = os.Environ()
		cmd.Stdin = strings.NewReader(in)
		out := bytes.NewBuffer(make([]byte, 0, 1<<8))
		cmd.Stdout = out
		cmd.Stderr = out

		log.Debug(cmd.String())
		err := cmd.Run()
		if err != nil {
			log.Debug(out.String())
			return "", fmt.Errorf("%s\n%s", err, out.String())
		}
		return strings.TrimSpace(out.String()), nil
	}
}
