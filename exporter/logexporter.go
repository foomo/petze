package exporter

import (
	"github.com/foomo/petze/watch"
	"github.com/sirupsen/logrus"
)

func LogResultHandler(result watch.Result) {
	logger := logrus.WithFields(logrus.Fields{
		"service_id": result.ID,
		"runtime":    result.RunTime,
		"timeout":    result.Timeout,
	})

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			if err.Comment != "" {
				logger = logger.WithField("comment", err.Comment)
			}
			logger.WithFields(logrus.Fields{
				"type":     err.Type,
				"location": err.Location,
			}).Error(err.Error)
		}
	} else {
		logger.Info("run completed without errors")
	}

}
