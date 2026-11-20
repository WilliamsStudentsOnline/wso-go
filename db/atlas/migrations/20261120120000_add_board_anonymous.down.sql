-- reverse: modify "board_threads" table
ALTER TABLE `board_threads` DROP COLUMN `anonymous`;
-- reverse: modify "board_posts" table
ALTER TABLE `board_posts` DROP COLUMN `anonymous`;
