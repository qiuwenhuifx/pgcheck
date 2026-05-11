WITH data AS (
    SELECT
        format('%I.%I', n.nspname, c.relname) AS relation,
        greatest(age(c.relfrozenxid), age(t.relfrozenxid)) AS xid_age,
        greatest(mxid_age(c.relminmxid), mxid_age(t.relminmxid)) AS mxid_age,
        pg_table_size(c.oid) AS table_size,
        pg_table_size(t.oid) AS toast_size
    FROM
        pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        LEFT JOIN pg_class t ON c.reltoastrelid = t.oid
    WHERE
        c.relkind IN ('r', 'm', 'p')
        AND n.nspname NOT IN ('pg_catalog', 'information_schema')
)
SELECT
    relation,
    xid_age,
    mxid_age,
    pg_size_pretty(table_size) AS table_size,
    pg_size_pretty(toast_size) AS toast_size
FROM
    data
ORDER BY
    greatest(xid_age, mxid_age) DESC
LIMIT 25;
