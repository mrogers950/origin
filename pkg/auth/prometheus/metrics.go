package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	AuthSubsystem = "auth_subsystem"
)

var (
	authCounterTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: AuthSubsystem,
			Name:      "auth_count",
			Help:      "Counts total authentication attempts",
		}, []string{},
	)
	authCounterUser = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: AuthSubsystem,
			Name:      "auth_count_user",
			Help:      "Counts total authentication attempts, by user",
		}, []string{"user"},
	)
	authCounterResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: AuthSubsystem,
			Name:      "auth_count_result",
			Help:      "Counts total authentication attempts, by result",
		}, []string{"result"},
	)
	authCounterUserResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: AuthSubsystem,
			Name:      "auth_count_user_result",
			Help:      "Counts total authentication attempts, by user and result",
		}, []string{"user", "result"},
	)
	authCounterPath = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: AuthSubsystem,
			Name:      "auth_count_path",
			Help:      "Counts total authentication attempts, by request path",
		}, []string{"path"},
	)
	authCounterUserPath = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: AuthSubsystem,
			Name:      "auth_count_user_path",
			Help:      "Counts total authentication attempts, by user and request path",
		}, []string{"user", "path"},
	)
)

func init() {
	prometheus.MustRegister(authCounterTotal)
	prometheus.MustRegister(authCounterUser)
	prometheus.MustRegister(authCounterResult)
	prometheus.MustRegister(authCounterUserResult)
	prometheus.MustRegister(authCounterPath)
	prometheus.MustRegister(authCounterUserPath)
}

func UpdateAuthCounters(user, path, result string) {
	authCounterTotal.WithLabelValues().Inc()
	authCounterUser.WithLabelValues(user).Inc()
	authCounterResult.WithLabelValues(result).Inc()
	authCounterUserResult.WithLabelValues(user, result).Inc()
	authCounterPath.WithLabelValues(path).Inc()
	authCounterUserPath.WithLabelValues(user, path).Inc()
}
