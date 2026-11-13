-- create "board_backfill_map" table
CREATE TABLE `board_backfill_map` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `source_table` varchar(255) NOT NULL,
  `source_id` int unsigned NOT NULL,
  `thread_id` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_board_backfill_map_deleted_at` (`deleted_at`),
  UNIQUE INDEX `idx_board_backfill_source` (`source_table`, `source_id`),
  INDEX `idx_board_backfill_thread` (`thread_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "board_posts" table
CREATE TABLE `board_posts` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `thread_id` int unsigned NOT NULL,
  `user_id` int unsigned NOT NULL,
  `ex_user_name` varchar(255) NULL,
  `content` longtext NOT NULL,
  `is_op` bool NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  INDEX `idx_board_posts_deleted_at` (`deleted_at`),
  INDEX `index_board_posts_on_thread_id` (`thread_id`),
  INDEX `index_board_posts_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "board_ride_information" table
CREATE TABLE `board_ride_information` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `thread_id` int unsigned NOT NULL,
  `offering_ride` bool NOT NULL,
  `source` varchar(255) NOT NULL,
  `destination` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_board_ride_information_deleted_at` (`deleted_at`),
  UNIQUE INDEX `index_board_ride_information_on_thread_id` (`thread_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "board_threads" table
CREATE TABLE `board_threads` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `type` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `user_id` int unsigned NOT NULL,
  `ex_user_name` varchar(255) NULL,
  `replies_enabled` bool NOT NULL DEFAULT 1,
  `resolved` bool NULL,
  `starts_at` datetime NULL,
  `ends_at` datetime NULL,
  `last_active` datetime NOT NULL,
  `reply_count` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  INDEX `idx_board_threads_deleted_at` (`deleted_at`),
  INDEX `index_board_threads_on_last_active` (`last_active`),
  INDEX `index_board_threads_on_type` (`type`),
  INDEX `index_board_threads_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
