CREATE TABLE `OrderItems`
(
    `id`                  INTEGER PRIMARY KEY          NOT NULL AUTO_INCREMENT,
    `order_id`            INTEGER                      NOT NULL,
    `item_id`             INTEGER                      NOT NULL,
    `count`               INTEGER                      NOT NULL,
    `status`              ENUM ('preparing','completed') NOT NULL,
    `custom_instructions` VARCHAR(255)
);

ALTER TABLE `OrderItems`
    ADD FOREIGN KEY (`order_id`) REFERENCES `Orders` (`id`);

ALTER TABLE `OrderItems`
    ADD FOREIGN KEY (`item_id`) REFERENCES `Items` (`id`);