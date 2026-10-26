-- create "moderator_tasks" table
CREATE TABLE `moderator_tasks` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `requesting_user_id` int unsigned NULL,
  `task_type` varchar(255) NOT NULL,
  `task_fields` text NOT NULL,
  `subject_key` varchar(255) NOT NULL,
  `resolved_at` datetime NULL,
  `resolution` varchar(255) NULL,
  `resolved_by_user_id` int unsigned NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_moderator_tasks_deleted_at` (`deleted_at`),
  INDEX `index_moderator_tasks_on_requesting_user_id` (`requesting_user_id`),
  INDEX `index_moderator_tasks_on_resolved_at` (`resolved_at`),
  INDEX `index_moderator_tasks_on_resolved_by_user_id` (`resolved_by_user_id`),
  INDEX `index_moderator_tasks_on_subject_key` (`subject_key`),
  INDEX `index_moderator_tasks_on_task_type` (`task_type`),
  INDEX `index_moderator_tasks_on_task_type_subject_key_resolved_at` (`task_type`, `subject_key`, `resolved_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
