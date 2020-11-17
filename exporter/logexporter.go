package exporter

import (
	"github.com/foomo/petze/watch"
	"github.com/sirupsen/logrus"
)

func LogServiceResultHandler(serviceResult watch.ServiceResult) {
	logger := logrus.WithFields(logrus.Fields{
		"service_id": serviceResult.ID,
		"runtime":    serviceResult.RunTime,
		"timeout":    serviceResult.Timeout,
	})

	if len(serviceResult.Errors) > 0 {
		for _, err := range serviceResult.Errors {
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
		"host_id": hostResult.ID,
		"rtt":     hostResult.RunTime,
		"timeout": hostResult.Timeout,
	})

	if len(hostResult.Errors) > 0 {
		for _, err := range hostResult.Errors {
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
