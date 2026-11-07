package deploy_test

import (
	"bytes"
	"cqrs-es-example-go/pkg/rmu"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// deployConfig mirrors deploy/rmu/event-source-mapping.json. The test rejects
// unknown fields so a typo cannot silently leave a safety setting unset.
type deployConfig struct {
	Version     int `json:"version"`
	RMUFunction struct {
		ImageCommand   string `json:"imageCommand"`
		TimeoutSeconds int    `json:"timeoutSeconds"`
	} `json:"rmuFunction"`
	HeadStreamEventSourceMapping struct {
		BatchSize                  int      `json:"batchSize"`
		BisectBatchOnFunctionError bool     `json:"bisectBatchOnFunctionError"`
		FunctionResponseTypes      []string `json:"functionResponseTypes"`
		MaximumRetryAttempts       int      `json:"maximumRetryAttempts"`
		MaximumRecordAgeInSeconds  int      `json:"maximumRecordAgeInSeconds"`
	} `json:"headStreamEventSourceMapping"`
	Retry struct {
		MaxAttempts       int `json:"maxAttempts"`
		MaxElapsedSeconds int `json:"maxElapsedSeconds"`
	} `json:"retry"`
	FailureQueue struct {
		MessageRetentionPeriodSeconds int `json:"messageRetentionPeriodSeconds"`
		VisibilityTimeoutSeconds      int `json:"visibilityTimeoutSeconds"`
	} `json:"failureQueue"`
	RecoveryWorkerFunction struct {
		ImageCommand   string `json:"imageCommand"`
		TimeoutSeconds int    `json:"timeoutSeconds"`
	} `json:"recoveryWorkerFunction"`
	RecoveryWorkerEventSourceMapping struct {
		BatchSize             int      `json:"batchSize"`
		FunctionResponseTypes []string `json:"functionResponseTypes"`
		MaximumRetryAttempts  int      `json:"maximumRetryAttempts"`
	} `json:"recoveryWorkerEventSourceMapping"`
}

func TestEventSourceMappingMatchesHandlerContract(t *testing.T) {
	cfg := loadDeployConfig(t)

	require.Equal(t, 1, cfg.Version)
	require.Equal(t, "rmu", cfg.RMUFunction.ImageCommand)
	require.Equal(t, "recoverRmuWorker", cfg.RecoveryWorkerFunction.ImageCommand)
	require.True(t, cfg.HeadStreamEventSourceMapping.BisectBatchOnFunctionError)
	// HandleDynamoDBEvent and HandleSQSEvent both answer with
	// batchItemFailures, which Lambda honors only for this response type.
	require.Equal(t, []string{"ReportBatchItemFailures"}, cfg.HeadStreamEventSourceMapping.FunctionResponseTypes)
	require.Equal(t, []string{"ReportBatchItemFailures"}, cfg.RecoveryWorkerEventSourceMapping.FunctionResponseTypes)
}

func TestRetryBudgetFitsInsideFunctionTimeout(t *testing.T) {
	cfg := loadDeployConfig(t)

	policy := rmu.DefaultRetryPolicy()
	require.Equal(t, policy.MaxAttempts, cfg.Retry.MaxAttempts)
	require.Equal(t, policy.MaxElapsed, time.Duration(cfg.Retry.MaxElapsedSeconds)*time.Second)
	// Leave at least as much time for batch processing as the retry budget
	// spends inside the invocation.
	require.GreaterOrEqual(t, cfg.RMUFunction.TimeoutSeconds, cfg.Retry.MaxElapsedSeconds*2)
	require.LessOrEqual(t, cfg.RMUFunction.TimeoutSeconds, 900)
	require.LessOrEqual(t, cfg.RecoveryWorkerFunction.TimeoutSeconds, 900)
}

func TestMappingSettingsStayInsideServiceLimits(t *testing.T) {
	cfg := loadDeployConfig(t)

	require.GreaterOrEqual(t, cfg.HeadStreamEventSourceMapping.BatchSize, 1)
	require.LessOrEqual(t, cfg.HeadStreamEventSourceMapping.BatchSize, 10000)
	require.GreaterOrEqual(t, cfg.HeadStreamEventSourceMapping.MaximumRetryAttempts, 1)
	require.LessOrEqual(t, cfg.HeadStreamEventSourceMapping.MaximumRetryAttempts, 10000)
	// Match the DynamoDB Streams retention window; older records must be
	// recovered through the resync path instead.
	require.Equal(t, 86400, cfg.HeadStreamEventSourceMapping.MaximumRecordAgeInSeconds)

	require.GreaterOrEqual(t, cfg.RecoveryWorkerEventSourceMapping.BatchSize, 1)
	require.LessOrEqual(t, cfg.RecoveryWorkerEventSourceMapping.BatchSize, 10)
	require.GreaterOrEqual(t, cfg.RecoveryWorkerEventSourceMapping.MaximumRetryAttempts, 1)
	require.LessOrEqual(t, cfg.RecoveryWorkerEventSourceMapping.MaximumRetryAttempts, 1000)

	require.GreaterOrEqual(t, cfg.FailureQueue.MessageRetentionPeriodSeconds, 60)
	require.LessOrEqual(t, cfg.FailureQueue.MessageRetentionPeriodSeconds, 1209600)
	// A message must not become visible again while the worker is still
	// running. Recovery is idempotent, so the margin does not need to be six
	// times the function timeout.
	require.GreaterOrEqual(t, cfg.FailureQueue.VisibilityTimeoutSeconds, cfg.RecoveryWorkerFunction.TimeoutSeconds)
	require.LessOrEqual(t, cfg.FailureQueue.VisibilityTimeoutSeconds, 43200)
}

func loadDeployConfig(t *testing.T) deployConfig {
	t.Helper()
	raw, err := os.ReadFile("event-source-mapping.json")
	require.NoError(t, err)
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var cfg deployConfig
	require.NoError(t, decoder.Decode(&cfg))
	return cfg
}
