CREATE TABLE `messages`
(
    `id`            varchar(64) NOT NULL,
    `group_chat_id` varchar(64) NOT NULL,
    `account_id`    varchar(64) NOT NULL,
    `text`          TEXT        NOT NULL,
    `created_at`    datetime    NOT NULL,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`group_chat_id`) REFERENCES group_chats (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

