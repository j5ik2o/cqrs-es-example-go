ALTER TABLE `read_model_recovery_tasks`
    DROP INDEX `idx_read_model_recovery_tasks_scan`,
    DROP COLUMN `attempt_count`;
