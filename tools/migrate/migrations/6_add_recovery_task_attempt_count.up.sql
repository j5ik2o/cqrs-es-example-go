ALTER TABLE `read_model_recovery_tasks`
    ADD COLUMN `attempt_count` int unsigned NOT NULL DEFAULT 0 AFTER `status`,
    ADD INDEX `idx_read_model_recovery_tasks_scan`
        (`store_id`, `projection_name`, `projection_version`, `status`, `attempt_count`, `created_at`);
