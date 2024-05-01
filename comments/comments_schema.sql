DROP TABLE IF EXISTS comments;

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    id_news BIGINT NOT NULL,
    id_parent BIGINT,
    content TEXT NOT NULL,
    commented_at BIGINT NOT NULL DEFAULT 0
);

INSERT INTO comments (id, id_news, id_parent, content, commented_at) VALUES (0, 0, 0, 'hello', 0);