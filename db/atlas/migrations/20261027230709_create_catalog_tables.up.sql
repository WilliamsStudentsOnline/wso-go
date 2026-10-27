-- create "courses_canonical" table
CREATE TABLE `courses_canonical` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `crse_id` varchar(32) NOT NULL,
  `title` varchar(255) NOT NULL,
  `description` longtext NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `crse_id` (`crse_id`),
  INDEX `idx_courses_canonical_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "course_listings" table
CREATE TABLE `course_listings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `course_canonical_id` int unsigned NOT NULL,
  `subject` varchar(16) NOT NULL,
  `number` int NOT NULL,
  `letter` varchar(8) NOT NULL,
  `first_valid_year` int NOT NULL,
  `last_valid_year` int NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_course_listings_course` (`course_canonical_id`),
  INDEX `idx_course_listings_deleted_at` (`deleted_at`),
  UNIQUE INDEX `idx_course_listings_identity` (`course_canonical_id`, `subject`, `number`, `letter`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "offerings" table
CREATE TABLE `offerings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `course_canonical_id` int unsigned NOT NULL,
  `strm` int NOT NULL,
  `year` int NOT NULL,
  `term` varchar(16) NOT NULL,
  `section` varchar(16) NOT NULL,
  `class_nbr` int NOT NULL,
  `title` varchar(255) NOT NULL,
  `component` varchar(32) NULL,
  `status` varchar(32) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_offerings_course` (`course_canonical_id`),
  INDEX `idx_offerings_deleted_at` (`deleted_at`),
  UNIQUE INDEX `idx_offerings_strm_class` (`strm`, `class_nbr`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "offering_instructors" table
CREATE TABLE `offering_instructors` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `offering_id` int unsigned NOT NULL,
  `unix_id` varchar(100) NULL,
  `user_id` int unsigned NULL,
  `name_as_listed` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_offering_instructors_deleted_at` (`deleted_at`),
  INDEX `idx_offering_instructors_offering` (`offering_id`),
  INDEX `idx_offering_instructors_unix` (`unix_id`),
  INDEX `idx_offering_instructors_user` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "offering_meetings" table
CREATE TABLE `offering_meetings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `offering_id` int unsigned NOT NULL,
  `days` varchar(16) NULL,
  `start` varchar(8) NULL,
  `end` varchar(8) NULL,
  `facility` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_offering_meetings_deleted_at` (`deleted_at`),
  INDEX `idx_offering_meetings_offering` (`offering_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
