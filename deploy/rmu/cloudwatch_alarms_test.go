package deploy_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// alarmConfig mirrors deploy/rmu/cloudwatch-alarms.json. The test rejects
// unknown fields so a typo cannot silently drop an alarm or a metric filter.
type alarmConfig struct {
	Version       int    `json:"version"`
	Namespace     string `json:"namespace"`
	MetricFilters []struct {
		Name         string `json:"name"`
		Pattern      string `json:"pattern"`
		MetricName   string `json:"metricName"`
		MetricValue  string `json:"metricValue"`
		DefaultValue int    `json:"defaultValue"`
	} `json:"metricFilters"`
	Alarms []struct {
		Name               string `json:"name"`
		Description        string `json:"description"`
		Namespace          string `json:"namespace"`
		MetricName         string `json:"metricName"`
		Statistic          string `json:"statistic"`
		PeriodSeconds      int    `json:"periodSeconds"`
		EvaluationPeriods  int    `json:"evaluationPeriods"`
		DatapointsToAlarm  int    `json:"datapointsToAlarm"`
		Threshold          int    `json:"threshold"`
		ComparisonOperator string `json:"comparisonOperator"`
		TreatMissingData   string `json:"treatMissingData"`
		Dimensions         []struct {
			Name      string `json:"name"`
			ValueFrom string `json:"valueFrom"`
		} `json:"dimensions"`
	} `json:"alarms"`
}

func TestAlarmsCoverTheRecoveryFailureModes(t *testing.T) {
	cfg := loadAlarmConfig(t)

	require.Equal(t, 1, cfg.Version)
	require.NotEmpty(t, cfg.Namespace)

	alarms := make(map[string]bool, len(cfg.Alarms))
	for _, alarm := range cfg.Alarms {
		require.NotEmpty(t, alarm.Name)
		require.NotEmpty(t, alarm.Description)
		require.NotEmpty(t, alarm.Namespace)
		require.NotEmpty(t, alarm.MetricName)
		require.Contains(t, []string{"Sum", "Maximum", "Average"}, alarm.Statistic)
		require.Contains(t, []string{"GreaterThanThreshold", "GreaterThanOrEqualToThreshold"}, alarm.ComparisonOperator)
		require.Contains(t, []string{"notBreaching", "breaching", "missing", "ignore"}, alarm.TreatMissingData)
		require.GreaterOrEqual(t, alarm.PeriodSeconds, 60)
		require.LessOrEqual(t, alarm.PeriodSeconds, 86400)
		require.GreaterOrEqual(t, alarm.EvaluationPeriods, 1)
		require.GreaterOrEqual(t, alarm.DatapointsToAlarm, 0)
		if alarm.DatapointsToAlarm > 0 {
			require.LessOrEqual(t, alarm.DatapointsToAlarm, alarm.EvaluationPeriods)
		}
		for _, dimension := range alarm.Dimensions {
			require.NotEmpty(t, dimension.Name)
			require.Contains(t, []string{"rmuFunction", "recoveryFunction", "failureQueueName"}, dimension.ValueFrom)
		}
		alarms[alarm.Name] = true
	}

	// The issue asks for alerts on a task backlog, retry exhaustion, and a
	// failed resync. Retry exhaustion shows up as a failed RMU invocation and
	// as a growing failure queue; both are covered.
	require.True(t, alarms["rmu-recovery-pending-tasks"])
	require.True(t, alarms["rmu-function-errors"])
	require.True(t, alarms["rmu-failure-queue-backlog"])
	require.True(t, alarms["rmu-recovery-resync-failures"])
}

func TestMetricFiltersReadTheCompletionLogLine(t *testing.T) {
	cfg := loadAlarmConfig(t)

	require.NotEmpty(t, cfg.MetricFilters)
	metricNames := make(map[string]bool, len(cfg.MetricFilters))
	for _, filter := range cfg.MetricFilters {
		require.NotEmpty(t, filter.Name)
		require.Contains(t, filter.Pattern, "recovery pass finished")
		require.NotEmpty(t, filter.MetricName)
		require.True(t, strings.HasPrefix(filter.MetricValue, "$."), filter.MetricValue)
		require.GreaterOrEqual(t, filter.DefaultValue, 0)
		require.False(t, metricNames[filter.MetricName])
		metricNames[filter.MetricName] = true
	}
	for _, required := range []string{
		"RecoveryPendingTasks", "RecoveryDeferredPasses", "RecoveryTaskFailures", "RecoveryResyncFailures",
	} {
		require.True(t, metricNames[required], "missing metric filter for %s", required)
	}
}

func loadAlarmConfig(t *testing.T) alarmConfig {
	t.Helper()
	raw, err := os.ReadFile("cloudwatch-alarms.json")
	require.NoError(t, err)
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var cfg alarmConfig
	require.NoError(t, decoder.Decode(&cfg))
	return cfg
}
