-- 002_mood_trash: seven-day recycle bin for mood records.
-- GORM AutoMigrate performs the same change on startup; this script is the
-- explicit migration copy for existing databases.
ALTER TABLE moods ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_moods_deleted_at ON moods (deleted_at);
