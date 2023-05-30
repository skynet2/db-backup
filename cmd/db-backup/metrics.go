package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	successTotalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "db_backup_total_success",
		Help: "The total number of success backups",
	}, []string{"job_name"})

	failTotalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "db_backup_total_errors",
		Help: "The total number of errors during backups",
	}, []string{"job_name"})

	successPerDbCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "db_backup_per_db_success",
		Help: "The total number of success backups [per db]",
	}, []string{"job_name", "db_name"})

	failPerDbCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "db_backup_per_db_errors",
		Help: "The total number of errors during backups [per db]",
	}, []string{"job_name", "db_name"})

	isRegistered = false
)

func registerMetrics() {
	if isRegistered {
		return
	}

	prometheus.MustRegister(successTotalCounter)
	prometheus.MustRegister(failTotalCounter)
	prometheus.MustRegister(successPerDbCounter)
	prometheus.MustRegister(failPerDbCounter)
	isRegistered = true
}

func pushMetrics(prometheusPushGatewayUrl string, jobName string) error {
	if prometheusPushGatewayUrl == "" {
		return nil
	}

	prometheusJobName := "default_job"
	if jobName != "" {
		prometheusJobName = jobName
	}

	return push.New(prometheusPushGatewayUrl, prometheusJobName).
		Collector(successTotalCounter).
		Collector(failTotalCounter).
		Collector(successPerDbCounter).
		Collector(failPerDbCounter).
		Push()
}
