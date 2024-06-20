package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/exp/constraints"
)

// types is an interface representing types that can be retrieved from environment variables.
type types interface {
	~bool | ~[]string | constraints.Ordered
}

// Get retrieves a variable from the environment. If not found, it returns the default value.
// If defaultValue is set and the variable is not found, it panics.
func Get[T types](envName string, defaultValue ...T) T {
	godotenv.Overload()
	// Get the value of the environment variable.
	value := os.Getenv(envName)

	var ret any = value
	var err error

	var def T
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}

	switch any(def).(type) {
	case string:
		ret = value

	case bool:
		ret, err = strconv.ParseBool(value)

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		ret, err = strconv.Atoi(value)
		ret = reflect.ValueOf(ret).Convert(reflect.TypeOf(def)).Interface()

	case float64:
		ret, err = strconv.ParseFloat(value, 64)

	case []string:
		if strings.Contains(value, ";") {
			ret = strings.Split(value, ";")
		} else {
			ret = strings.Split(value, ",")
		}
	}

	switch {
	case value == "" && len(defaultValue) == 0:
		// If the required variable is not set and no default value is provided, panic.
		panic(fmt.Sprintf("Required variable '%s' is not set - type: '%T'", envName, def))
	case value == "":
		// If the variable is not set but a default value is provided, return the default value.
		ret = def
	case err != nil:
		// If the variable cannot be parsed, panic.
		panic(fmt.Sprintf("Variable '%s' could not be parsed - type: '%T', value: '%s'", envName, def, value))
	}

	return ret.(T)
}
