SELECT
    slot_name,
    plugin,
    slot_type,
    database,
    active,
    xmin,
    age(xmin) AS xmin_age,
    catalog_xmin,
    age(catalog_xmin) AS catalog_xmin_age,
    restart_lsn,
    confirmed_flush_lsn,
    pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)) AS retained_wal
FROM
    pg_replication_slots
ORDER BY
    greatest(age(xmin), age(catalog_xmin)) DESC NULLS LAST,
    pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) DESC NULLS LAST;
