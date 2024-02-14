CREATE TABLE `members`
(
    `id`              varchar(64) NOT NULL,
    `group_chat_id`   varchar(64) NOT NULL,
    `user_account_id` varchar(64) NOT NULL,
    `role`            varchar(64) NOT NULL,
    `created_at`      datetime    NOT NULL,
    `updated_at`      datetime    NOT NULL,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`group_chat_id`) REFERENCES group_chats (`id`),
    UNIQUE KEY `group_chat_id_user_account_id` (`group_chat_id`, `user_account_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;
