-- +goose Up
-- +goose StatementBegin
CREATE TABLE chat_rooms (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255),
    type ENUM('GROUP', 'PRIVATE') NOT NULL,
    created_at DATETIME NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE messages (
    id VARCHAR(36) PRIMARY KEY,
    sender_id VARCHAR(36) NOT NULL,
    receiver_id VARCHAR(36) NOT NULL,
    type ENUM('TEXT', 'IMAGE', 'VIDEO', 'AUDIO', 'FILE') NOT NULL,
    mime_type VARCHAR(255),
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    chat_room_id VARCHAR(36),
    FOREIGN KEY (chat_room_id) REFERENCES chat_rooms(id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE chat_room_members (
    chat_room_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (chat_room_id, user_id),
    FOREIGN KEY (chat_room_id) REFERENCES chat_rooms(id),
    FOREIGN KEY (user_id) REFERENCES users(id) -- 👈 Kết nối với bảng users luôn
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chat_room_members;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS messages;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS chat_rooms;
-- +goose StatementEnd