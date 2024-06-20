package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fr-str/httpea/internal/config"
	"github.com/fr-str/httpea/internal/log"
)

// PrintJSON marshals the given value to JSON with an indented format and prints it.
func PrintJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(err, v)
		return
	}

	log.Debug("\n", string(b))
}

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

func GetPeaFilePath(f string) string {
	return filepath.Join(config.FileFolder, fmt.Sprintf("%s.pea", f))
}

func ReadPeaFile(f string) string {
	b, err := os.ReadFile(GetPeaFilePath(f))
	if err != nil {
		log.Debug("read read err: ", err)
	}
	return string(b)
}

// Must asserts that the given value is not an error.
// If it is, it logs the error and exits the program.
func Must[T any](v T, err error) T {
	if err != nil {
		panic("must failed" + err.Error())
	}
	return v
}

func PrettyJSON(b []byte) ([]byte, error) {
	cmd := exec.Command("jq", "-C")
	cmd.Stdin = bytes.NewBuffer(b)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return out, nil

}
