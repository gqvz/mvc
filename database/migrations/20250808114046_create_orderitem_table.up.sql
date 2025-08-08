CREATE TABLE `OrderItems`
(
    `id`                  INTEGER PRIMARY KEY NOT NULL AUTO_INCREMENT,
    `order_id`            INTEGER             NOT NULL,
    `item_id`             INTEGER             NOT NULL,
    `count`               INTEGER             NOT NULL,
    `status`              ENUM ('preparing','completed') NOT NULL,
    `custom_instructions` VARCHAR(255),
    FOREIGN KEY (`order_id`) REFERENCES `Orders` (`id`),
    FOREIGN KEY (`item_id`) REFERENCES `Items` (`id`)
);
