CREATE TABLE `group_chats`
(
    `id`         varchar(64) NOT NULL,
    `name`       varchar(64) NOT NULL,
    `owner_id`   varchar(64) NOT NULL,
    `created_at` datetime    NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;
