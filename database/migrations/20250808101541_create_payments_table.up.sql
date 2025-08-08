CREATE TABLE `Payments`
(
    `id`             INTEGER PRIMARY KEY AUTO_INCREMENT,
    `user_id`        INTEGER                         NOT NULL,
    `cashier_id`     INTEGER                         NOT NULL,
    `order_id`       INTEGER                         NOT NULL,
    `order_subtotal` DECIMAL(10, 2)                  NOT NULL,
    `tip`            DECIMAL(6, 2)                   NOT NULL,
    `status`         ENUM ('processing','accepted') NOT NULL,
    `total`          DECIMAL(10, 2) GENERATED ALWAYS AS (order_subtotal + tip) STORED,
    FOREIGN KEY (`user_id`) REFERENCES `Users` (`id`),
    FOREIGN KEY (`cashier_id`) REFERENCES `Users` (`id`),
    FOREIGN KEY (`order_id`) REFERENCES `Orders` (`id`)
);
