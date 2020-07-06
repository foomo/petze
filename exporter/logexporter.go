package exporter

import (
	log "github.com/sirupsen/logrus"
	"github.com/dreadl0ck/petze/watch"
)

func LogResultHandler(result watch.Result) {
	logger := log.WithFields(log.Fields{
		"service_id": result.ID,
		"runtime":    result.RunTime,
		"timeout":    result.Timeout,
	})

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			if err.Comment != "" {
				logger = logger.WithField("comment", err.Comment)
			}
			logger.WithField("type", err.Type).Error(err.Error)
		}
	} else {
		logger.Info("run completed without errors")
	}

}
