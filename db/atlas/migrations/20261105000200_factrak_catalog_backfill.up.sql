-- modify "factrak_surveys" table
ALTER TABLE `factrak_surveys` ADD COLUMN `canonical_course_id` int unsigned NULL AFTER `course_id`, ADD COLUMN `offering_id` int unsigned NULL AFTER `canonical_course_id`, ADD INDEX `index_factrak_surveys_on_canonical_course_id` (`canonical_course_id`), ADD INDEX `index_factrak_surveys_on_offering_id` (`offering_id`);
-- create "course_subject_aliases" table
CREATE TABLE `course_subject_aliases` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `factrak_abbrev` varchar(16) NOT NULL,
  `catalog_subject` varchar(16) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `factrak_abbrev` (`factrak_abbrev`),
  INDEX `idx_course_subject_aliases_deleted_at` (`deleted_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- seed high-confidence Factrak → catalog subject renames
INSERT INTO `course_subject_aliases` (`created_at`, `updated_at`, `factrak_abbrev`, `catalog_subject`)
VALUES
  (UTC_TIMESTAMP(), UTC_TIMESTAMP(), 'WGST', 'WGSS'),
  (UTC_TIMESTAMP(), UTC_TIMESTAMP(), 'ASST', 'ASIA');
