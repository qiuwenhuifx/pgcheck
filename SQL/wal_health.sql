WITH settings AS (
    SELECT
        current_setting('archive_mode', true) AS archive_mode,
        current_setting('archive_command', true) AS archive_command,
        current_setting('wal_keep_size', true) AS wal_keep_size,
        current_setting('max_wal_size', true) AS max_wal_size,
        current_setting('checkpoint_timeout', true) AS checkpoint_timeout,
        current_setting('max_slot_wal_keep_size', true) AS max_slot_wal_keep_size
), slots AS (
    SELECT
        count(*) AS slot_count,
        count(*) FILTER (WHERE NOT active) AS inactive_slot_count,
        pg_size_pretty(max(pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn))) AS max_retained_wal
    FROM
        pg_replication_slots
), archiver AS (
    SELECT
        archived_count,
        failed_count,
        last_archived_time,
        last_failed_time,
        stats_reset
    FROM
        pg_stat_archiver
)
SELECT
    settings.archive_mode,
    nullif(settings.archive_command, '') IS NOT NULL AS has_archive_command,
    archiver.archived_count,
    archiver.failed_count,
    archiver.last_archived_time,
    archiver.last_failed_time,
    settings.wal_keep_size,
    settings.max_wal_size,
    settings.checkpoint_timeout,
    settings.max_slot_wal_keep_size,
    slots.slot_count,
    slots.inactive_slot_count,
    slots.max_retained_wal,
    archiver.stats_reset
FROM
    settings,
    slots,
    archiver;
