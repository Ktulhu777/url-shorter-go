ALTER TABLE url ADD COLUMN clicks INTEGER DEFAULT 3;

CREATE TABLE click_details (
    id INTEGER PRIMARY KEY,
    url_id INTEGER REFERENCES url(id) ON DELETE CASCADE,
    ip INET,
    user_agent TEXT,
    country    VARCHAR(100),
    device     VARCHAR(50),
    browser    VARCHAR(50),
    referrer   TEXT,
    created_at TIMESTAMP
);