package deploy_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testHeadStreamARN = "arn:aws:dynamodb:ap-northeast-1:000000000000:table/head/stream/2026-01-01T00:00:00.000"
	testQueueURL      = "http://localhost:4566/000000000000/cqrs-es-example-go-rmu-failure"
)

func TestApplyScriptsDoNotCallAWSInDryRun(t *testing.T) {
	requireTool(t, "bash")
	requireTool(t, "jq")

	dir := scriptDir(t)
	binDir := t.TempDir()
	marker := filepath.Join(t.TempDir(), "aws-called")
	fakeAWS := "#!/bin/sh\n: > \"$FAKE_AWS_MARKER\"\nexit 99\n"
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "aws"), []byte(fakeAWS), 0o755))

	env := append(os.Environ(),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"FAKE_AWS_MARKER="+marker,
	)

	output, code := runScript(t, dir, env, "apply.sh",
		"--rmu-function", "test-rmu",
		"--head-stream-arn", testHeadStreamARN,
		"--failure-queue-url", testQueueURL,
		"--recovery-function", "test-recovery")
	require.Equal(t, 0, code, output)
	require.Contains(t, output, "dry-run: no AWS calls were made.")
	require.Contains(t, output, "batch 100, bisect true, ReportBatchItemFailures, retries 3, max age 86400s")
	require.Contains(t, output, "retention 1209600s, visibility 960s")
	require.Contains(t, output, "rate(5 minutes), ENABLED")

	output, code = runScript(t, dir, env, "apply-alarms.sh",
		"--rmu-function", "test-rmu",
		"--recovery-function", "test-recovery",
		"--failure-queue-url", testQueueURL)
	require.Equal(t, 0, code, output)
	require.Contains(t, output, "dry-run: no AWS calls were made.")
	require.Contains(t, output, "filter:     rmu-recovery-pending-tasks -> RecoveryPendingTasks from $.pending")
	require.Contains(t, output, "alarm:      rmu-recovery-pending-tasks")

	_, err := os.Stat(marker)
	require.True(t, os.IsNotExist(err), "a dry run must not invoke the aws CLI")
}

func TestApplyScriptsRequireTheirFlags(t *testing.T) {
	requireTool(t, "bash")

	dir := scriptDir(t)
	for _, script := range []string{"apply.sh", "apply-alarms.sh"} {
		output, code := runScript(t, dir, os.Environ(), script)
		require.Equalf(t, 2, code, "%s: %s", script, output)
		require.Contains(t, output, "missing required option", script)
	}
}

func TestApplyScriptsHaveValidShellSyntax(t *testing.T) {
	requireTool(t, "bash")

	dir := scriptDir(t)
	for _, script := range []string{"apply.sh", "apply-alarms.sh"} {
		command := exec.Command("bash", "-n", script)
		command.Dir = dir
		output, err := command.CombinedOutput()
		require.NoErrorf(t, err, "%s: %s", script, output)
	}
}

func TestApplyScriptsAreExecutable(t *testing.T) {
	dir := scriptDir(t)
	for _, script := range []string{"apply.sh", "apply-alarms.sh"} {
		info, err := os.Stat(filepath.Join(dir, script))
		require.NoError(t, err)
		require.NotZero(t, info.Mode()&0o111, "%s must be executable", script)
	}
}

func scriptDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	return dir
}

func requireTool(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s is required for this test", name)
	}
}

// runScript runs a deploy script with bash and returns its combined output and
// exit code. A non-zero exit is a result, not a test failure.
func runScript(t *testing.T, dir string, env []string, script string, args ...string) (string, int) {
	t.Helper()
	command := exec.Command("bash", append([]string{script}, args...)...)
	command.Dir = dir
	command.Env = env
	output, err := command.CombinedOutput()
	if err == nil {
		return string(output), 0
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("run %s: %v", script, err)
	}
	return string(output), exitErr.ExitCode()
}
