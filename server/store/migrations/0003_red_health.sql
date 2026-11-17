-- Red's health % per battle (0 = full), set from the run's match number
-- (OpenSpec change: run-balance). Stored so old battles keep verifying with
-- the value they were issued with, even if the curve changes.
ALTER TABLE battles ADD COLUMN red_health_pct INTEGER NOT NULL DEFAULT 0;
