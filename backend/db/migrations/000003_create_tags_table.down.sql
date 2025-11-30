-- Rollback: Drop tags and article_tags tables
DROP INDEX IF EXISTS idx_article_tags_tag_id;
DROP INDEX IF EXISTS idx_article_tags_article_id;
DROP INDEX IF EXISTS idx_tags_name;
DROP TABLE IF EXISTS article_tags;
DROP TABLE IF EXISTS tags;
