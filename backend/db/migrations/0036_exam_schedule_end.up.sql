-- 0036_exam_schedule_end.up.sql
-- Optional exam availability window: when set, students may check in/start
-- any time in [scheduled_at, scheduled_end_at] instead of only at the single
-- scheduled_at instant. NULL preserves today's fixed-instant behavior.

ALTER TABLE exam ADD COLUMN scheduled_end_at TIMESTAMPTZ;
