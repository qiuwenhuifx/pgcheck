WITH raw AS (
    SELECT
        s.*,
        to_jsonb(s) AS stat
    FROM
        pg_stat_io AS s
), io AS (
    SELECT
        backend_type,
        object,
        context,
        reads,
        coalesce((stat ->> 'read_bytes')::numeric, reads::numeric * coalesce((stat ->> 'op_bytes')::numeric, current_setting('block_size')::numeric)) AS read_bytes,
        read_time,
        writes,
        coalesce((stat ->> 'write_bytes')::numeric, writes::numeric * coalesce((stat ->> 'op_bytes')::numeric, current_setting('block_size')::numeric)) AS write_bytes,
        write_time,
        writebacks,
        writeback_time,
        extends,
        coalesce((stat ->> 'extend_bytes')::numeric, extends::numeric * coalesce((stat ->> 'op_bytes')::numeric, current_setting('block_size')::numeric)) AS extend_bytes,
        extend_time,
        hits,
        evictions,
        reuses,
        fsyncs,
        fsync_time,
        (reads + writes + extends + evictions + fsyncs) AS total_ops,
        (read_time + write_time + writeback_time + extend_time + fsync_time) AS total_time,
        (
            coalesce((stat ->> 'read_bytes')::numeric, reads::numeric * coalesce((stat ->> 'op_bytes')::numeric, current_setting('block_size')::numeric)) +
            coalesce((stat ->> 'write_bytes')::numeric, writes::numeric * coalesce((stat ->> 'op_bytes')::numeric, current_setting('block_size')::numeric)) +
            coalesce((stat ->> 'extend_bytes')::numeric, extends::numeric * coalesce((stat ->> 'op_bytes')::numeric, current_setting('block_size')::numeric))
        ) AS total_bytes
    FROM
        raw
)
SELECT
    backend_type,
    object,
    context,
    total_ops,
    total_time,
    pg_size_pretty(total_bytes) AS io_bytes,
    reads,
    pg_size_pretty(read_bytes) AS read_bytes,
    read_time,
    writes,
    pg_size_pretty(write_bytes) AS write_bytes,
    write_time,
    extends,
    pg_size_pretty(extend_bytes) AS extend_bytes,
    extend_time,
    evictions,
    reuses,
    fsyncs,
    fsync_time
FROM
    io
ORDER BY
    total_time DESC NULLS LAST,
    total_bytes DESC NULLS LAST,
    total_ops DESC NULLS LAST
LIMIT 30;
