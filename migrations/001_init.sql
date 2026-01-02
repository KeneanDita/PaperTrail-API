CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS
    users (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        public_id UUID NOT NULL DEFAULT gen_random_uuid () UNIQUE,
        email TEXT NOT NULL UNIQUE,
        role TEXT NOT NULL DEFAULT 'user',
        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS
    papers (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        public_id UUID NOT NULL DEFAULT gen_random_uuid () UNIQUE,
        title TEXT NOT NULL,
        abstract TEXT,
        author_id BIGINT NOT NULL,
        pdf_url TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_papers_author FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS
    reviews (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        public_id UUID NOT NULL DEFAULT gen_random_uuid () UNIQUE,
        paper_id BIGINT NOT NULL,
        reviewer_id BIGINT NOT NULL,
        rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
        comments TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_reviews_paper FOREIGN KEY (paper_id) REFERENCES papers (id) ON DELETE CASCADE,
        CONSTRAINT fk_reviews_reviewer FOREIGN KEY (reviewer_id) REFERENCES users (id) ON DELETE CASCADE,
        CONSTRAINT unique_reviewer_per_paper UNIQUE (paper_id, reviewer_id)
    );

CREATE TABLE IF NOT EXISTS
    comments (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        public_id UUID NOT NULL DEFAULT gen_random_uuid () UNIQUE,
        paper_id BIGINT NOT NULL,
        user_id BIGINT NOT NULL,
        body TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_comments_paper FOREIGN KEY (paper_id) REFERENCES papers (id) ON DELETE CASCADE,
        CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS idx_users_public_id ON users (public_id);

CREATE INDEX IF NOT EXISTS idx_papers_public_id ON papers (public_id);

CREATE INDEX IF NOT EXISTS idx_reviews_public_id ON reviews (public_id);

CREATE INDEX IF NOT EXISTS idx_comments_public_id ON comments (public_id);

CREATE INDEX IF NOT EXISTS idx_papers_author_id ON papers (author_id);

CREATE INDEX IF NOT EXISTS idx_reviews_paper_id ON reviews (paper_id);

CREATE INDEX IF NOT EXISTS idx_comments_paper_id ON comments (paper_id);
