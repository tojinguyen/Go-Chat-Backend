-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS friend_requests (
    id VARCHAR(36) PRIMARY KEY,
    user_id_requester VARCHAR(36) NOT NULL,
    user_id_receiver VARCHAR(36) NOT NULL,
    status ENUM('pending', 'accepted', 'rejected', 'cancelled') NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id_requester) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id_receiver) REFERENCES users(id) ON DELETE CASCADE,
    
    UNIQUE KEY idx_unique_request (user_id_requester, user_id_receiver),
    INDEX idx_receiver (user_id_receiver)
);

CREATE TABLE IF NOT EXISTS friendships (
    id VARCHAR(36) PRIMARY KEY,
    user_id_a VARCHAR(36) NOT NULL,
    user_id_b VARCHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id_a) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id_b) REFERENCES users(id) ON DELETE CASCADE,
    
    UNIQUE KEY idx_unique_friendship (user_id_a, user_id_b),
    INDEX idx_user_b (user_id_b)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS friendships;
DROP TABLE IF EXISTS friend_requests;
-- +goose StatementEnd
