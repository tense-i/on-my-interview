CREATE TABLE IF NOT EXISTS crawl_jobs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    trigger_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    platforms_json JSON NOT NULL,
    keywords_json JSON NOT NULL,
    pages INT NOT NULL DEFAULT 1,
    force_reparse BOOLEAN NOT NULL DEFAULT FALSE,
    started_at DATETIME NULL,
    finished_at DATETIME NULL,
    stats_json JSON NULL,
    error_message TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS raw_posts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    platform VARCHAR(64) NOT NULL,
    source_post_id VARCHAR(128) NOT NULL,
    title TEXT NOT NULL,
    content MEDIUMTEXT NOT NULL,
    content_hash VARCHAR(128) NOT NULL,
    author_name VARCHAR(255) NOT NULL,
    post_url TEXT NOT NULL,
    company_name_raw VARCHAR(255) NULL,
    company_name_norm VARCHAR(255) NULL,
    source_created_at DATETIME NOT NULL,
    source_edited_at DATETIME NOT NULL,
    last_crawled_at DATETIME NOT NULL,
    parse_status VARCHAR(32) NOT NULL,
    parse_attempts INT NOT NULL DEFAULT 0,
    raw_payload_json JSON NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_raw_posts_platform_source (platform, source_post_id)
);

CREATE TABLE IF NOT EXISTS post_parse_results (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    raw_post_id BIGINT NOT NULL,
    schema_version VARCHAR(32) NOT NULL,
    llm_provider VARCHAR(64) NOT NULL,
    llm_model VARCHAR(128) NOT NULL,
    company_name VARCHAR(255) NULL,
    sentiment VARCHAR(64) NULL,
    sentiment_reason TEXT NULL,
    key_events_json JSON NOT NULL,
    questions_json JSON NOT NULL,
    tags_json JSON NOT NULL,
    raw_json JSON NOT NULL,
    parsed_at DATETIME NOT NULL,
    UNIQUE KEY uniq_post_parse_results_raw_post (raw_post_id)
);

CREATE TABLE IF NOT EXISTS interview_questions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    raw_post_id BIGINT NOT NULL,
    platform VARCHAR(64) NOT NULL,
    company_name VARCHAR(255) NULL,
    question_text TEXT NOT NULL,
    question_order INT NOT NULL,
    category VARCHAR(128) NULL,
    tags_json JSON NOT NULL,
    source_excerpt TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS question_tags (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    question_id BIGINT NOT NULL,
    tag VARCHAR(128) NOT NULL
);

CREATE TABLE IF NOT EXISTS crawl_job_posts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    job_id BIGINT NOT NULL,
    raw_post_id BIGINT NOT NULL,
    disposition VARCHAR(32) NOT NULL,
    message TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
