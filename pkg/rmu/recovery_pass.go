package rmu

import (
	"context"
	"errors"
)

// RecoveryPassResult summarises one resync and pending-task replay.
type RecoveryPassResult struct {
	Resync    ResyncResult
	Completed int
}

// RecoveryPass is the recovery job behind both the one-shot recoverRmu command
// and the SQS-triggered recoverRmuWorker. It reconciles every head from the
// journal and then replays the pending tasks a permanent failure recorded.
type RecoveryPass struct {
	resyncer *Resyncer
	recovery *RecoveryService
}

func NewRecoveryPass(resyncer *Resyncer, recovery *RecoveryService) *RecoveryPass {
	return &RecoveryPass{resyncer: resyncer, recovery: recovery}
}

// Run executes both steps even when the resync fails. Pending tasks are
// independent work, and a transient scan failure must not block them.
func (p *RecoveryPass) Run(ctx context.Context) (RecoveryPassResult, error) {
	var result RecoveryPassResult
	resyncResult, resyncErr := p.resyncer.Resync(ctx)
	result.Resync = resyncResult
	completed, recoveryErr := p.recovery.RecoverPending(ctx)
	result.Completed = completed
	return result, errors.Join(resyncErr, recoveryErr)
}
