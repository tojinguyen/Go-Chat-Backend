CREATE TABLE IF NOT EXISTS friend_requests (
    user_id_requester VARCHAR(36) NOT NULL,
    user_id_receiver VARCHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL,
    PRIMARY KEY (user_id_requester, user_id_receiver)
);

CREATE TABLE IF NOT EXISTS friendships (
    user_id_a VARCHAR(36) NOT NULL,
    user_id_b VARCHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id_a, user_id_b)
);