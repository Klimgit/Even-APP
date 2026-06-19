CREATE TABLE user_favorite_block_types (
    user_id UUID NOT NULL,
    block_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, block_type)
);
