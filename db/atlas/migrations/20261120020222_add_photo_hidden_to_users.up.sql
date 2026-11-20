-- modify "users" table
ALTER TABLE `users` ADD COLUMN `photo_hidden` bool NOT NULL DEFAULT 0 AFTER `home_visible`;
