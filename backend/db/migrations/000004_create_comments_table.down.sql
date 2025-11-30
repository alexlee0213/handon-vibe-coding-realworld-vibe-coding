-- Rollback: Drop comments table and indexes
DROP INDEX IF EXISTS idx_comments_created_at;
DROP INDEX IF EXISTS idx_comments_author_id;
DROP INDEX IF EXISTS idx_comments_article_id;
DROP TABLE IF EXISTS comments;
