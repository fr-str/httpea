package config

import (
	"io"

	"github.com/fr-str/httpea/internal/util/env"
)

// Database configuration parameters.
var (
	Debug      = env.Get("HTTPEA_DEBUG", false)
	FileFolder = env.Get("HTTPEA_FOLDER", "pea")
)
var DebugLogFile io.Writer = io.Discard
