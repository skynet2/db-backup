package main

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setupZeroLog() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		sp := strings.Split(file, "/")

		segments := 4

		if len(sp) == 0 { // just in case
			segments = 0
		}

		if segments > 0 && segments > len(sp) {
			segments = len(sp) - 1
		}

		return fmt.Sprintf("%s:%v", strings.Join(sp[segments:], "/"), line)
	}

	log.Logger = log.Logger.With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.DefaultContextLogger = &log.Logger
}
