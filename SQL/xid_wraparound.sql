WITH data AS (
    SELECT
        oid,
        datname,
        age(datfrozenxid) AS xid_age,
        mxid_age(datminmxid) AS mxid_age,
        pg_database_size(datname) AS db_size
    FROM
        pg_database
)
SELECT
    datname,
    xid_age,
    mxid_age,
    round(100::numeric * xid_age::numeric / 2000000000::numeric, 2) AS xid_used_pct,
    round(100::numeric * mxid_age::numeric / 2000000000::numeric, 2) AS mxid_used_pct,
    pg_size_pretty(db_size) AS db_size
FROM
    data
ORDER BY
    greatest(xid_age, mxid_age) DESC;
