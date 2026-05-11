WITH io AS (
    SELECT
        backend_type,
        object,
        context,
        reads,
        read_time,
        writes,
        write_time,
        writebacks,
        writeback_time,
        extends,
        extend_time,
        hits,
        evictions,
        reuses,
        fsyncs,
        fsync_time,
        op_bytes,
        (reads + writes + extends + evictions + fsyncs) AS total_ops,
        (read_time + write_time + writeback_time + extend_time + fsync_time) AS total_time
    FROM
        pg_stat_io
)
SELECT
    backend_type,
    object,
    context,
    total_ops,
    total_time,
    reads,
    read_time,
    writes,
    write_time,
    extends,
    extend_time,
    evictions,
    reuses,
    fsyncs,
    fsync_time,
    pg_size_pretty(op_bytes * greatest(reads + writes + extends, 0)) AS io_volume
FROM
    io
ORDER BY
    total_time DESC NULLS LAST,
    total_ops DESC NULLS LAST
LIMIT 30;
