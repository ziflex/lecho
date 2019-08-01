package lecho

import (
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
)

var levels = map[log.Lvl]zerolog.Level{
	log.DEBUG: zerolog.DebugLevel,
	log.INFO:  zerolog.InfoLevel,
	log.WARN:  zerolog.WarnLevel,
	log.ERROR: zerolog.ErrorLevel,
	log.OFF:   zerolog.NoLevel,
}

func getLevel(level log.Lvl) (zerolog.Level, log.Lvl) {
	zlvl, found := levels[level]

	if found {
		return zlvl, level
	}

	return zerolog.NoLevel, log.OFF
}