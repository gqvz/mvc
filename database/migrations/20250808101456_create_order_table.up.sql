CREATE TABLE `Orders`
(
    `id`           INTEGER PRIMARY KEY AUTO_INCREMENT,
    `ordered_at`   DATETIME                NOT NULL,
    `customer_id`  INTEGER                 NOT NULL,
    `table_number` INTEGER                 NOT NULL,
    `status`       ENUM ('open', 'closed') NOT NULL,
    FOREIGN KEY (`customer_id`) REFERENCES `Users` (`id`)
);
