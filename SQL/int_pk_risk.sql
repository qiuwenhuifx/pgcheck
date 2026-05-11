WITH pk_cols AS (
    SELECT
        c.oid AS table_oid,
        n.nspname AS schema_name,
        c.relname AS table_name,
        a.attname AS column_name,
        t.typname AS type_name,
        seq.oid AS sequence_oid,
        seq_ns.nspname AS sequence_schema,
        seq.relname AS sequence_name
    FROM
        pg_index i
        JOIN pg_class c ON c.oid = i.indrelid
        JOIN pg_namespace n ON n.oid = c.relnamespace
        JOIN pg_attribute a ON a.attrelid = i.indrelid
            AND a.attnum = ANY (i.indkey)
        JOIN pg_type t ON t.oid = a.atttypid
        LEFT JOIN pg_depend d ON d.refobjid = c.oid
            AND d.refobjsubid = a.attnum
            AND d.deptype IN ('a', 'i')
        LEFT JOIN pg_class seq ON seq.oid = d.objid
            AND seq.relkind = 'S'
        LEFT JOIN pg_namespace seq_ns ON seq_ns.oid = seq.relnamespace
    WHERE
        i.indisprimary
        AND c.relkind IN ('r', 'p')
        AND t.typname IN ('int2', 'int4')
        AND n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
    GROUP BY
        c.oid,
        n.nspname,
        c.relname,
        a.attname,
        t.typname,
        seq.oid,
        seq_ns.nspname,
        seq.relname
    HAVING
        count(*) = 1
), data AS (
    SELECT
        format('%I.%I', p.schema_name, p.table_name) AS relation,
        p.column_name,
        p.type_name,
        CASE
            WHEN p.sequence_oid IS NULL THEN NULL
            ELSE format('%I.%I', p.sequence_schema, p.sequence_name)
        END AS sequence_name,
        ps.last_value,
        CASE p.type_name
            WHEN 'int2' THEN 32767::numeric
            WHEN 'int4' THEN 2147483647::numeric
        END AS max_value,
        pg_total_relation_size(p.table_oid) AS total_size
    FROM
        pk_cols p
        LEFT JOIN pg_sequences ps ON ps.schemaname = p.sequence_schema
            AND ps.sequencename = p.sequence_name
)
SELECT
    relation,
    column_name,
    type_name,
    sequence_name,
    last_value,
    max_value::bigint,
    round(100 * last_value::numeric / nullif(max_value, 0), 2) AS capacity_used_pct,
    pg_size_pretty(total_size) AS total_size
FROM
    data
ORDER BY
    capacity_used_pct DESC NULLS LAST,
    total_size DESC;
