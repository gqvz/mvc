CREATE TABLE `Items`
(
    `id`           INTEGER PRIMARY KEY AUTO_INCREMENT,
    `name`         VARCHAR(32)   NOT NULL,
    `description`  VARCHAR(255)  NOT NULL,
    `price`        DECIMAL(6, 2) NOT NULL,
    `is_available` BOOLEAN       NOT NULL,
    `image_url`    VARCHAR(255)  NOT NULL
);

CREATE TABLE `Tags`
(
    `id`   INTEGER PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(32) NOT NULL
);

CREATE TABLE `ItemTags`
(
    `item_id` INTEGER NOT NULL,
    `tag_id`  INTEGER NOT NULL,
    PRIMARY KEY (`item_id`, `tag_id`)
);

ALTER TABLE `ItemTags`
    ADD FOREIGN KEY (`item_id`) REFERENCES `Items` (`id`);

ALTER TABLE `ItemTags`
    ADD FOREIGN KEY (`tag_id`) REFERENCES `Tags` (`id`);