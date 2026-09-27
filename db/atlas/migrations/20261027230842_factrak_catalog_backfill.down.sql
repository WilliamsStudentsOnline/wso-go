-- reverse: create "course_subject_aliases" table
DROP TABLE `course_subject_aliases`;
-- reverse: modify "factrak_surveys" table
ALTER TABLE `factrak_surveys` DROP INDEX `index_factrak_surveys_on_offering_id`, DROP INDEX `index_factrak_surveys_on_canonical_course_id`, DROP COLUMN `offering_id`, DROP COLUMN `canonical_course_id`;
