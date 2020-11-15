package exporter

import (
	"github.com/foomo/petze/watch"
	"github.com/sirupsen/logrus"
)

func LogServiceResultHandler(serviceResult watch.ServiceResult) {
	logger := logrus.WithFields(logrus.Fields{
		"service_id": serviceResult.Result.ID,
		"runtime":    serviceResult.Result.RunTime,
		"timeout":    serviceResult.Result.Timeout,
	})

	if len(serviceResult.Result.Errors) > 0 {
		for _, err := range serviceResult.Result.Errors {
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

func LogHostResultHandler(hostResult watch.HostResult) {
	logger := logrus.WithFields(logrus.Fields{
		"host_id": hostResult.Result.ID,
		"rtt":     hostResult.Result.RunTime,
		"timeout": hostResult.Result.Timeout,
	})

	if len(hostResult.Result.Errors) > 0 {
		for _, err := range hostResult.Result.Errors {
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
