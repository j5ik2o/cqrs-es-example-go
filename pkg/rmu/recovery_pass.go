package rmu

import (
	"context"
	"errors"
)

// RecoveryPassResult summarises one resync and pending-task replay.
type RecoveryPassResult struct {
	Resync       ResyncResult
	Recovery     RecoveryPendingResult
	ResyncFailed bool
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
	resyncResult, resyncErr := p.resyncer.Resync(ctx)
	recoveryResult, recoveryErr := p.recovery.RecoverPending(ctx)
	return RecoveryPassResult{
		Resync:       resyncResult,
		Recovery:     recoveryResult,
		ResyncFailed: resyncErr != nil,
	}, errors.Join(resyncErr, recoveryErr)
}

// RunScheduled replays pending tasks without scanning the head table. The
// schedule is the safety net for a lost SQS notification, so the durable tasks
// are the only work it needs. A full scan every few minutes would read the
// whole head table for no reason; the notification and the manual recoverRmu
// command keep the full reconciliation path.
func (p *RecoveryPass) RunScheduled(ctx context.Context) (RecoveryPassResult, error) {
	recoveryResult, err := p.recovery.RecoverPending(ctx)
	return RecoveryPassResult{Recovery: recoveryResult}, err
}
