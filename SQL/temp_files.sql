WITH tablespaces AS (
    SELECT
        spcname AS tablespace_name,
        coalesce(nullif(pg_tablespace_location(oid), ''), current_setting('data_directory') || '/base') AS tablespace_location
    FROM
        pg_tablespace
), temp_dirs AS (
    SELECT
        tablespace_name,
        tablespace_location || '/pgsql_tmp' AS path
    FROM
        tablespaces
    WHERE
        tablespace_name = 'pg_default'
    UNION ALL
    SELECT
        tablespace_name,
        tablespace_location || '/' || version_dir || '/pgsql_tmp' AS path
    FROM
        tablespaces,
        LATERAL pg_ls_dir(tablespace_location, true, false) AS version_dir
    WHERE
        tablespace_name <> 'pg_default'
        AND version_dir ~ ('^PG_' || split_part(current_setting('server_version'), '.', 1))
), temp_files AS (
    SELECT
        substring(file_name FROM '[0-9]+') AS pid,
        tablespace_name AS temp_tablespace,
        sum((pg_stat_file(path || '/' || file_name, true)).size) AS temp_bytes
    FROM
        temp_dirs,
        LATERAL pg_ls_dir(path, true, false) AS file_name
    GROUP BY
        pid,
        temp_tablespace
)
SELECT
    a.datname,
    a.pid,
    pg_size_pretty(coalesce(t.temp_bytes, 0)) AS temp_size_written,
    coalesce(t.temp_tablespace, 'not using temp files') AS temp_tablespace,
    a.application_name,
    a.client_addr,
    a.usename,
    (clock_timestamp() - a.query_start)::interval(0) AS duration,
    (clock_timestamp() - a.state_change)::interval(0) AS duration_since_state_change,
    trim(trailing ';' FROM left(a.query, 1000)) AS query,
    a.state,
    a.wait_event_type || ':' || a.wait_event AS wait
FROM
    pg_stat_activity AS a
    LEFT JOIN temp_files AS t ON a.pid = t.pid::int
WHERE
    a.pid <> pg_backend_pid()
ORDER BY
    coalesce(t.temp_bytes, 0) DESC,
    duration DESC NULLS LAST;
