-- Create "areas_of_study" table
CREATE TABLE `areas_of_study` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `name` varchar(255) NOT NULL,
  `abbrev` varchar(4) NOT NULL,
  `department_id` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `abbrev` (`abbrev`),
  INDEX `idx_areas_of_study_deleted_at` (`deleted_at`),
  UNIQUE INDEX `name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "banned_users" table
CREATE TABLE `banned_users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `reason` varchar(255) NULL,
  `factrak` bool NULL DEFAULT 0,
  `dormtrak` bool NULL DEFAULT 0,
  `ephcatch` bool NULL DEFAULT 0,
  `bulletin_read` bool NULL DEFAULT 0,
  `bulletin_write` bool NULL DEFAULT 0,
  `ephmatch` bool NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  INDEX `idx_banned_users_deleted_at` (`deleted_at`),
  INDEX `index_banned_users_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "book_listings" table
CREATE TABLE `book_listings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `book_id` int unsigned NOT NULL,
  `user_id` int unsigned NOT NULL,
  `condition` varchar(255) NULL,
  `description` longtext NULL,
  `listing_type` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_book_listings_deleted_at` (`deleted_at`),
  INDEX `index_book_listings_on_book_id` (`book_id`),
  INDEX `index_book_listings_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "books" table
CREATE TABLE `books` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `title` varchar(255) NOT NULL,
  `subtitle` varchar(255) NULL,
  `authors` varchar(255) NULL,
  `publisher` varchar(255) NULL,
  `isbn` varchar(255) NOT NULL,
  `info_link` longtext NULL,
  `image_link` longtext NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_books_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "bulletin_rides" table
CREATE TABLE `bulletin_rides` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `body` longtext NULL,
  `date` datetime NULL,
  `offer` bool NOT NULL,
  `source` varchar(255) NOT NULL,
  `destination` varchar(255) NOT NULL,
  `user_id` int unsigned NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_bulletin_rides_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "bulletins" table
CREATE TABLE `bulletins` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `type` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `body` longtext NULL,
  `start_date` datetime NULL,
  `end_date` datetime NULL,
  `offer` bool NULL,
  `user_id` int unsigned NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_bulletins_deleted_at` (`deleted_at`),
  INDEX `index_bulletins_on_type` (`type`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "course_book" table
CREATE TABLE `course_book` (
  `course_id` int unsigned NOT NULL,
  `book_id` int unsigned NOT NULL,
  PRIMARY KEY (`course_id`, `book_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "courses" table
CREATE TABLE `courses` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `number` varchar(255) NOT NULL,
  `area_of_study_id` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_courses_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "departments" table
CREATE TABLE `departments` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_departments_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "discussions" table
CREATE TABLE `discussions` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `last_active` datetime NULL,
  `title` varchar(255) NULL,
  `ex_user_name` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_discussions_deleted_at` (`deleted_at`),
  INDEX `index_discussions_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "dorm_rooms" table
CREATE TABLE `dorm_rooms` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `dorm_id` int unsigned NOT NULL,
  `number` varchar(255) NOT NULL,
  `closet` longtext NULL,
  `flooring` varchar(255) NULL,
  `common_room_access` bool NULL,
  `common_room_desc` longtext NULL,
  `thermostat_access` bool NULL,
  `key_or_card` varchar(255) NULL,
  `noise` longtext NULL,
  `bed_adjustable` bool NULL,
  `private_bathroom` bool NULL,
  `bathroom_desc` longtext NULL,
  `picture` varchar(255) NULL,
  `room_type` varchar(255) NOT NULL,
  `faces` varchar(255) NULL,
  `hc` bool NULL,
  `floor_number` int NULL,
  `area` int NULL,
  `walkthrough` bool NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_dorm_rooms_deleted_at` (`deleted_at`),
  INDEX `index_dorm_rooms_on_dorm_id` (`dorm_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "dorms" table
CREATE TABLE `dorms` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `neighborhood_id` int unsigned NOT NULL,
  `name` varchar(255) NULL,
  `key_or_card` varchar(255) NULL,
  `description` longtext NULL,
  `built` int NULL,
  `capacity` int NOT NULL DEFAULT 0,
  `number_bathrooms` int NOT NULL DEFAULT 0,
  `number_singles` int NOT NULL DEFAULT 0,
  `number_doubles` int NOT NULL DEFAULT 0,
  `number_flex` int NOT NULL DEFAULT 0,
  `number_washers` int NOT NULL DEFAULT 0,
  `bathroom_ratio` double NULL,
  `average_single_area` int NULL,
  `average_double_area` int NULL,
  `mode_single_area` int NULL,
  `mode_double_area` int NULL,
  `loudness` double NULL,
  `wifi` double NULL,
  `location` double NULL,
  `satisfaction` double NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_dorms_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "dormtrak_reviews" table
CREATE TABLE `dormtrak_reviews` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `dorm_room_id` int unsigned NOT NULL,
  `comment` longtext NULL,
  `closet` longtext NULL,
  `flooring` varchar(255) NULL,
  `common_room_access` bool NULL,
  `common_room_desc` longtext NULL,
  `thermostat_access` bool NULL,
  `key_or_card` varchar(255) NULL,
  `noise` longtext NULL,
  `bed_adjustable` bool NULL,
  `private_bathroom` bool NULL,
  `bathroom_desc` longtext NULL,
  `loudness` int NULL,
  `wifi` int NULL,
  `location` int NULL,
  `satisfaction` int NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_dormtrak_reviews_deleted_at` (`deleted_at`),
  INDEX `index_dormtrak_reviews_on_dorm_room_id` (`dorm_room_id`),
  INDEX `index_dormtrak_reviews_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "ephcatches" table
CREATE TABLE `ephcatches` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `seen` bool NOT NULL DEFAULT 0,
  `user_id` int unsigned NULL,
  `other_id` int unsigned NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_ephcatches_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "ephmatch_likes" table
CREATE TABLE `ephmatch_likes` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `liked_id` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_ephmatch_likes_deleted_at` (`deleted_at`),
  INDEX `index_ephmatch_likes_on_liked_id` (`liked_id`),
  INDEX `index_ephmatch_likes_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "ephmatch_matches" table
CREATE TABLE `ephmatch_matches` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_a_id` int unsigned NOT NULL,
  `user_b_id` int unsigned NOT NULL,
  `user_a_seen` bool NOT NULL DEFAULT 0,
  `user_b_seen` bool NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  INDEX `idx_ephmatch_matches_deleted_at` (`deleted_at`),
  INDEX `index_ephmatch_matches_on_user_a_id` (`user_a_id`),
  INDEX `index_ephmatch_matches_on_user_b_id` (`user_b_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "ephmatch_profiles" table
CREATE TABLE `ephmatch_profiles` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `description` varchar(255) NULL,
  `match_message` varchar(255) NULL,
  `location_visible` bool NOT NULL DEFAULT 1,
  `location_town` varchar(255) NULL,
  `location_state` varchar(255) NULL,
  `location_country` varchar(255) NULL,
  `messaging_platform` varchar(255) NULL,
  `messaging_username` varchar(255) NULL,
  `looking_for` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_ephmatch_profiles_deleted_at` (`deleted_at`),
  INDEX `index_ephmatch_profiles_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "ephmatch_relations" table
CREATE TABLE `ephmatch_relations` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `other_id` int unsigned NOT NULL,
  `relation` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_ephmatch_relations_deleted_at` (`deleted_at`),
  INDEX `index_ephmatch_relations_on_other_id` (`other_id`),
  INDEX `index_ephmatch_relations_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "factrak_agreements" table
CREATE TABLE `factrak_agreements` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `agrees` bool NULL,
  `factrak_survey_id` int unsigned NULL,
  `user_id` int unsigned NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_factrak_agreements_deleted_at` (`deleted_at`),
  INDEX `index_factrak_agreements_on_factrak_survey_id` (`factrak_survey_id`),
  INDEX `index_factrak_agreements_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "factrak_surveys" table
CREATE TABLE `factrak_surveys` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `professor_id` int unsigned NOT NULL,
  `course_id` int unsigned NOT NULL,
  `would_recommend_course` bool NULL,
  `course_workload` int NULL,
  `course_stimulating` int NULL,
  `would_take_another` bool NULL,
  `approachability` int NULL,
  `lead_lecture` int NULL,
  `promote_discussion` int NULL,
  `outside_helpfulness` int NULL,
  `mental_health_support` int NULL,
  `comment` longtext NULL,
  `flagged` bool NULL,
  `grade_received` varchar(255) NULL,
  `semester_season` varchar(255) NULL,
  `semester_year` int NULL,
  `course_format` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_factrak_surveys_deleted_at` (`deleted_at`),
  INDEX `index_factrak_surveys_on_course_id` (`course_id`),
  INDEX `index_factrak_surveys_on_professor_id` (`professor_id`),
  INDEX `index_factrak_surveys_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "goodrich_menu_items" table
CREATE TABLE `goodrich_menu_items` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `title` varchar(255) NULL,
  `description` varchar(255) NULL,
  `price` double NULL,
  `type` varchar(255) NULL,
  `category` varchar(255) NULL,
  `available` bool NULL,
  `quantity_limit` bool NOT NULL DEFAULT 0,
  `quantity` int NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_goodrich_menu_items_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "goodrich_orders" table
CREATE TABLE `goodrich_orders` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `status` int NULL,
  `phone_number` varchar(255) NULL,
  `date` varchar(255) NULL,
  `time_slot` varchar(255) NULL,
  `notes` varchar(255) NULL,
  `total_price` double NULL,
  `combo_deal` bool NOT NULL DEFAULT 0,
  `admin_notes` varchar(255) NULL,
  `payment_method` int NULL,
  `id_number` varchar(255) NULL,
  `item_list` varchar(255) NULL,
  `user_id` int unsigned NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_goodrich_orders_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "neighborhoods" table
CREATE TABLE `neighborhoods` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `name` varchar(255) NULL,
  `trakked` bool NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  INDEX `idx_neighborhoods_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "notification_settings" table
CREATE TABLE `notification_settings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `enable_notifications` bool NULL,
  `salmon_notify` bool NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_notification_settings_deleted_at` (`deleted_at`),
  INDEX `index_notification_settings_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "notification_tokens" table
CREATE TABLE `notification_tokens` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `type` varchar(255) NULL,
  `token` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_notification_tokens_deleted_at` (`deleted_at`),
  INDEX `index_notification_tokens_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "offices" table
CREATE TABLE `offices` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `number` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_offices_deleted_at` (`deleted_at`),
  UNIQUE INDEX `number` (`number`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "posts" table
CREATE TABLE `posts` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `discussion_id` int unsigned NOT NULL,
  `content` longtext NOT NULL,
  `ex_user_name` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_posts_deleted_at` (`deleted_at`),
  INDEX `index_posts_on_discussion_id` (`discussion_id`),
  INDEX `index_posts_on_user_id` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "tags" table
CREATE TABLE `tags` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `name` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_tags_deleted_at` (`deleted_at`),
  UNIQUE INDEX `name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "tags_users" table
CREATE TABLE `tags_users` (
  `user_id` int unsigned NOT NULL,
  `tag_id` int unsigned NOT NULL,
  PRIMARY KEY (`user_id`, `tag_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "user_areaOfStudy" table
CREATE TABLE `user_areaOfStudy` (
  `user_id` int unsigned NOT NULL,
  `area_of_study_id` int unsigned NOT NULL,
  PRIMARY KEY (`user_id`, `area_of_study_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "users" table
CREATE TABLE `users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `type` varchar(255) NULL,
  `name` varchar(255) NULL,
  `cell_phone` varchar(255) NULL,
  `campus_phone_ext` varchar(255) NULL,
  `unix_id` varchar(100) NOT NULL,
  `williams_email` varchar(255) NULL,
  `title` varchar(255) NULL,
  `visible` bool NOT NULL DEFAULT 1,
  `class_year` int NULL,
  `nickname` varchar(255) NULL,
  `dorm_visible` bool NOT NULL DEFAULT 1,
  `home_town` varchar(255) NULL,
  `home_zip` varchar(255) NULL,
  `home_phone` varchar(255) NULL,
  `home_state` varchar(255) NULL,
  `home_country` varchar(255) NULL,
  `home_visible` bool NOT NULL DEFAULT 1,
  `major` varchar(255) NULL,
  `su_box` varchar(255) NULL,
  `entry` varchar(255) NULL,
  `admin` bool NOT NULL DEFAULT 0,
  `factrak_admin` bool NOT NULL DEFAULT 0,
  `has_accepted_factrak_policy` bool NOT NULL DEFAULT 0,
  `has_accepted_dormtrak_policy` bool NOT NULL DEFAULT 0,
  `department_id` int unsigned NULL,
  `office_id` int unsigned NULL,
  `dorm_room_id` int unsigned NULL,
  `pronoun` varchar(255) NULL,
  `at_williams` bool NOT NULL DEFAULT 1,
  `off_cycle` bool NOT NULL DEFAULT 0,
  `factrak_survey_deficit` int NULL,
  `on_campus_semesters` int NOT NULL DEFAULT 0,
  `opt_out_ephcatch` bool NOT NULL DEFAULT 0,
  `ephcatch_eligibility` bool NOT NULL DEFAULT 0,
  `off_campus` bool NOT NULL DEFAULT 0,
  `search_fields` varchar(255) NULL DEFAULT "",
  `campus_status` varchar(255) NULL,
  `williams_id` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_users_deleted_at` (`deleted_at`),
  INDEX `index_rooms_on_dorm_room_id` (`dorm_room_id`),
  UNIQUE INDEX `unix_id` (`unix_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
