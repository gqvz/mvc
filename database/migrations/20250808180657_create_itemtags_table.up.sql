CREATE TABLE `ItemTags`
(
    `item_id` INTEGER NOT NULL,
    `tag_id`  INTEGER NOT NULL,
    PRIMARY KEY (`item_id`, `tag_id`),
    FOREIGN KEY (`item_id`) REFERENCES `Items` (`id`),
    FOREIGN KEY (`tag_id`) REFERENCES `Tags` (`id`)
);
