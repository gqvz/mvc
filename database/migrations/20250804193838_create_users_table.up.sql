CREATE TABLE IF NOT EXISTS `Users`
(
    `id`            INTEGER PRIMARY KEY AUTO_INCREMENT,
    `name`          VARCHAR(255) NOT NULL,
    `email`         VARCHAR(255) NOT NULL,
    `role`          TINYINT      NOT NULL,
    `password_hash` CHAR(60)     NOT NULL
);