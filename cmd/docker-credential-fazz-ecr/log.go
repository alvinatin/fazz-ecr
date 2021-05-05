package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/payfazz/go-errors/v2"
)

var logFile io.Writer

func init() {
	path := os.Getenv("FAZZ_ECR_LOG_FILE")
	if path != "" {
		if f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600); err == nil {
			logFile = f
		}
	}
	if logFile == nil {
		logFile = io.Discard
	}
}

func log(err error) error {
	if err == nil {
		return nil
	}

	fmt.Fprintf(logFile, "%v\n%s\n\n", time.Now(), errors.Format(err))

	return err
}
