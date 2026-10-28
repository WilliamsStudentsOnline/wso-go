-- create "feedback" table
CREATE TABLE `feedback` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NULL,
  `preferred_name` varchar(255) NULL,
  `priority` varchar(255) NOT NULL DEFAULT 'none',
  `url` varchar(255) NULL,
  `notes` longtext NOT NULL,
  `admin_notes` longtext NULL,
  `issue` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_feedback_deleted_at` (`deleted_at`),
  INDEX `index_feedback_on_priority` (`priority`),
  INDEX `index_feedback_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
