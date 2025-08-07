CREATE TABLE `Requests`
(
    `id`          INTEGER PRIMARY KEY AUTO_INCREMENT,
    `user_id`     INTEGER                                 NOT NULL,
    `role`        TINYINT                                 NOT NULL,
    `status`      ENUM ('pending', 'granted', 'rejected') NOT NULL,
    `user_status` ENUM ('seen', 'unseen')                 NOT NULL,
    `granted_by`  INTEGER
);