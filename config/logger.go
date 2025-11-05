package config

import (
    "log"
    "os"
)

// This file will handle log writer and rotating log files functionality
func SetupLogger() {
    // TODO: Implement log writer and rotating log files
    log.SetOutput(os.Stdout)
    log.SetFlags(log.LstdFlags | log.Lshortfile)
}
