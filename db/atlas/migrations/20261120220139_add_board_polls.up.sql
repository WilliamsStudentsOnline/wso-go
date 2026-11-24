-- create "board_polls" table
CREATE TABLE `board_polls` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `thread_id` int unsigned NOT NULL,
  `closes_at` datetime NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_board_polls_deleted_at` (`deleted_at`),
  UNIQUE INDEX `index_board_polls_on_thread_id` (`thread_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "board_poll_options" table
CREATE TABLE `board_poll_options` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `poll_id` int unsigned NOT NULL,
  `position` int NOT NULL,
  `label` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_board_poll_options_deleted_at` (`deleted_at`),
  INDEX `index_board_poll_options_on_poll_id` (`poll_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "board_poll_votes" table
CREATE TABLE `board_poll_votes` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `poll_id` int unsigned NOT NULL,
  `user_id` int unsigned NOT NULL,
  `option_id` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_board_poll_votes_deleted_at` (`deleted_at`),
  INDEX `index_board_poll_votes_on_option_id` (`option_id`),
  UNIQUE INDEX `index_board_poll_votes_on_poll_id_and_user_id` (`poll_id`, `user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
