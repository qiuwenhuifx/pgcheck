SELECT
    c.num_timed + c.num_requested AS total_checkpoints,
    pg_size_pretty(bs.block_size * b.buffers_clean) AS bgworker_total_written,
    pg_size_pretty(bs.block_size * c.buffers_written) AS checkpointer_total_written,
    NULL::text AS backend_total_written,
    pg_size_pretty(bs.block_size * c.buffers_written / NULLIF(c.num_timed + c.num_requested, 0)) AS checkpoint_write_avg,
    EXTRACT(EPOCH FROM (now() - pg_postmaster_start_time())) / NULLIF(c.num_timed + c.num_requested, 0) / 60 AS minutes_between_checkpoints,
    NULL::bigint AS buffers_backend_fsync,
    c.write_time,
    c.sync_time,
    b.maxwritten_clean,
    b.buffers_alloc
FROM
    pg_stat_checkpointer c
    CROSS JOIN pg_stat_bgwriter b
    CROSS JOIN (
        SELECT current_setting('block_size')::int AS block_size) AS bs;
