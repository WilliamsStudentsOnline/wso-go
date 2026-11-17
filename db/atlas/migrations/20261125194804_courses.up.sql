-- modify "areas_of_study" table
ALTER TABLE `areas_of_study` ADD COLUMN `catalog_kind` varchar(32) NULL AFTER `abbrev`, ADD COLUMN `catalog_url` varchar(255) NULL AFTER `catalog_kind`, ADD COLUMN `curriculum_pdf_url` varchar(255) NULL AFTER `catalog_url`, ADD COLUMN `website_url` varchar(255) NULL AFTER `curriculum_pdf_url`, ADD COLUMN `description` text NULL AFTER `website_url`;
-- modify "factrak_surveys" table
ALTER TABLE `factrak_surveys` ADD COLUMN `course_canonical_id` int unsigned NULL AFTER `course_id`, ADD COLUMN `course_offering_id` int unsigned NULL AFTER `course_canonical_id`, ADD COLUMN `user_course_id` int unsigned NULL AFTER `course_offering_id`, ADD COLUMN `course_link_verification_source` varchar(32) NULL AFTER `user_course_id`, ADD INDEX `index_factrak_surveys_on_course_canonical_id` (`course_canonical_id`), ADD INDEX `index_factrak_surveys_on_course_link_verification_source` (`course_link_verification_source`), ADD INDEX `index_factrak_surveys_on_course_offering_id` (`course_offering_id`), ADD INDEX `index_factrak_surveys_on_user_course_id` (`user_course_id`);
-- create "course_attributes" table
CREATE TABLE `course_attributes` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `code` varchar(64) NOT NULL,
  `name` varchar(255) NOT NULL,
  `kind` varchar(32) NOT NULL,
  `area_of_study_id` int unsigned NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `code` (`code`),
  INDEX `idx_course_attributes_area` (`area_of_study_id`),
  INDEX `idx_course_attributes_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "course_canonicals" table
CREATE TABLE `course_canonicals` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `crse_id` varchar(32) NOT NULL,
  `title` varchar(255) NOT NULL,
  `description` longtext NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `crse_id` (`crse_id`),
  INDEX `idx_course_canonicals_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "course_in_semesters" table
CREATE TABLE `course_in_semesters` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `course_canonical_id` int unsigned NOT NULL,
  `strm` int NOT NULL,
  `year` int NOT NULL,
  `term` varchar(16) NOT NULL,
  `status` varchar(32) NOT NULL,
  `source` varchar(16) NOT NULL DEFAULT 'catalog',
  `title` varchar(255) NOT NULL,
  `description` longtext NULL,
  `short_description` longtext NULL,
  `components` varchar(64) NULL,
  `class_format` longtext NULL,
  `grading_basis` varchar(16) NULL,
  `units` varchar(8) NULL,
  `campus` varchar(16) NULL,
  `attributes` varchar(1024) NULL,
  `div1` bool NOT NULL DEFAULT 0,
  `div2` bool NOT NULL DEFAULT 0,
  `div3` bool NOT NULL DEFAULT 0,
  `dpe` bool NOT NULL DEFAULT 0,
  `qfr` bool NOT NULL DEFAULT 0,
  `wac` bool NOT NULL DEFAULT 0,
  `prereqs` longtext NULL,
  `enrollment_preferences` longtext NULL,
  `requirements_evaluation` longtext NULL,
  `department_notes` longtext NULL,
  `wac_notes` longtext NULL,
  `dpe_notes` longtext NULL,
  `qfr_notes` longtext NULL,
  `extra_info` longtext NULL,
  `material_fee` longtext NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_course_in_semesters_course` (`course_canonical_id`),
  INDEX `idx_course_in_semesters_deleted_at` (`deleted_at`),
  INDEX `idx_course_in_semesters_strm` (`strm`),
  UNIQUE INDEX `idx_course_in_semesters_term` (`course_canonical_id`, `strm`)
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
-- create "course_offerings" table
CREATE TABLE `course_offerings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `course_canonical_id` int unsigned NOT NULL,
  `course_in_semester_id` int unsigned NULL,
  `source` varchar(16) NOT NULL DEFAULT 'catalog',
  `strm` int NOT NULL,
  `year` int NOT NULL,
  `term` varchar(16) NOT NULL,
  `section` varchar(16) NOT NULL,
  `class_nbr` int NOT NULL,
  `title` varchar(255) NOT NULL,
  `component` varchar(32) NULL,
  `status` varchar(32) NOT NULL,
  `subject` varchar(16) NULL,
  `catalog_number` int NULL,
  `catalog_letter` varchar(8) NULL,
  `attributes` varchar(1024) NULL,
  `div1` bool NOT NULL DEFAULT 0,
  `div2` bool NOT NULL DEFAULT 0,
  `div3` bool NOT NULL DEFAULT 0,
  `dpe` bool NOT NULL DEFAULT 0,
  `qfr` bool NOT NULL DEFAULT 0,
  `wac` bool NOT NULL DEFAULT 0,
  `consent` varchar(8) NULL,
  `enrollment_limit` int NULL,
  `expected_enrollment` int NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_course_offerings_course` (`course_canonical_id`),
  INDEX `idx_course_offerings_course_in_semester` (`course_in_semester_id`),
  INDEX `idx_course_offerings_deleted_at` (`deleted_at`),
  UNIQUE INDEX `idx_course_offerings_strm_class` (`strm`, `class_nbr`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "course_offering_instructors" table
CREATE TABLE `course_offering_instructors` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `course_offering_id` int unsigned NOT NULL,
  `unix_id` varchar(100) NULL,
  `user_id` int unsigned NULL,
  `name_as_listed` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_course_offering_instructors_course_offering` (`course_offering_id`),
  INDEX `idx_course_offering_instructors_deleted_at` (`deleted_at`),
  INDEX `idx_course_offering_instructors_unix` (`unix_id`),
  INDEX `idx_course_offering_instructors_user` (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "course_offering_meetings" table
CREATE TABLE `course_offering_meetings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `course_offering_id` int unsigned NOT NULL,
  `days` varchar(16) NULL,
  `start` varchar(8) NULL,
  `end` varchar(8) NULL,
  `facility` varchar(255) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_course_offering_meetings_course_offering` (`course_offering_id`),
  INDEX `idx_course_offering_meetings_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "factrak_deficit_exemptions" table
CREATE TABLE `factrak_deficit_exemptions` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `semester_season` varchar(32) NOT NULL,
  `semester_year` int NOT NULL,
  `reason` longtext NOT NULL,
  `status` varchar(16) NOT NULL DEFAULT 'pending',
  PRIMARY KEY (`id`),
  INDEX `idx_factrak_deficit_exemptions_deleted_at` (`deleted_at`),
  INDEX `index_factrak_deficit_exemptions_on_status` (`status`),
  INDEX `index_factrak_deficit_exemptions_on_user_id_semester` (`user_id`, `semester_season`, `semester_year`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "user_courses" table
CREATE TABLE `user_courses` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_id` int unsigned NOT NULL,
  `course_canonical_id` int unsigned NOT NULL,
  `course_in_semester_id` int unsigned NULL,
  `hidden` bool NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  INDEX `idx_user_courses_course` (`course_canonical_id`),
  INDEX `idx_user_courses_deleted_at` (`deleted_at`),
  UNIQUE INDEX `idx_user_courses_term` (`user_id`, `course_in_semester_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- create "user_course_offerings" table
CREATE TABLE `user_course_offerings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `user_course_id` int unsigned NOT NULL,
  `course_offering_id` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_user_course_offerings_course_offering` (`course_offering_id`),
  INDEX `idx_user_course_offerings_deleted_at` (`deleted_at`),
  UNIQUE INDEX `idx_user_course_offerings_pair` (`user_course_id`, `course_offering_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
