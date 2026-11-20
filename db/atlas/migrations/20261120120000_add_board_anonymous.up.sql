-- modify "board_posts" table
ALTER TABLE `board_posts` ADD COLUMN `anonymous` bool NOT NULL DEFAULT 0;
-- modify "board_threads" table
ALTER TABLE `board_threads` ADD COLUMN `anonymous` bool NOT NULL DEFAULT 0;
