package pea

import (
	"fmt"
	"os"
	"testing"

	"github.com/fr-str/httpea/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	// This is the entry point for the test
	f, err := os.OpenFile("/home/user/code/httpea/debug.log",
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		panic(err)
	}
	config.LogFile = f
	fmt.Fprintln(f, "------------------------- dupa -------------------------")
}

func TestShellExpr(t *testing.T) {
	tests := []struct {
		in       string
		expr     string
		expected string
	}{
		{
			in:       `{"token":"udpaudpaudp"}`,
			expr:     `jq -r .token`,
			expected: "udpaudpaudp\n",
		},
	}

	for _, tt := range tests {
		out, _ := shellExpr(tt.expr)(tt.in)
		assert.Equalf(t, tt.expected, out, "got %v", out)
	}
}
