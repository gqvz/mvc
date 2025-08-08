CREATE TABLE `Items`
(
    `id`           INTEGER PRIMARY KEY AUTO_INCREMENT,
    `name`         VARCHAR(32)   NOT NULL,
    `description`  VARCHAR(255)  NOT NULL,
    `price`        DECIMAL(6, 2) NOT NULL,
    `is_available` BOOLEAN       NOT NULL,
    `image_url`    VARCHAR(255)  NOT NULL
);