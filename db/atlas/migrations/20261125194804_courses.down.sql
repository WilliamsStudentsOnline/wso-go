-- reverse: create "user_course_offerings" table
DROP TABLE `user_course_offerings`;
-- reverse: create "user_courses" table
DROP TABLE `user_courses`;
-- reverse: create "factrak_deficit_exemptions" table
DROP TABLE `factrak_deficit_exemptions`;
-- reverse: create "course_offering_meetings" table
DROP TABLE `course_offering_meetings`;
-- reverse: create "course_offering_instructors" table
DROP TABLE `course_offering_instructors`;
-- reverse: create "course_offerings" table
DROP TABLE `course_offerings`;
-- reverse: create "course_listings" table
DROP TABLE `course_listings`;
-- reverse: create "course_in_semesters" table
DROP TABLE `course_in_semesters`;
-- reverse: create "course_canonicals" table
DROP TABLE `course_canonicals`;
-- reverse: create "course_attributes" table
DROP TABLE `course_attributes`;
-- reverse: modify "factrak_surveys" table
ALTER TABLE `factrak_surveys` DROP INDEX `index_factrak_surveys_on_user_course_id`, DROP INDEX `index_factrak_surveys_on_course_offering_id`, DROP INDEX `index_factrak_surveys_on_course_link_verification_source`, DROP INDEX `index_factrak_surveys_on_course_canonical_id`, DROP COLUMN `course_link_verification_source`, DROP COLUMN `user_course_id`, DROP COLUMN `course_offering_id`, DROP COLUMN `course_canonical_id`;
-- reverse: modify "areas_of_study" table
ALTER TABLE `areas_of_study` DROP COLUMN `description`, DROP COLUMN `website_url`, DROP COLUMN `curriculum_pdf_url`, DROP COLUMN `catalog_url`, DROP COLUMN `catalog_kind`;
