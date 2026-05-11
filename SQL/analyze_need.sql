WITH table_opts AS (
    SELECT
        c.oid,
        c.relname,
        n.nspname,
        array_to_string(c.reloptions, '') AS relopts
    FROM
        pg_class c
        JOIN pg_namespace n ON c.relnamespace = n.oid
    WHERE
        c.relkind IN ('r', 'p', 'm')
), analyze_settings AS (
    SELECT
        oid,
        relname,
        nspname,
        CASE
            WHEN relopts LIKE '%autovacuum_analyze_threshold%' THEN
                regexp_replace(relopts, '.*autovacuum_analyze_threshold=([0-9.]+).*', e'\\1')::int8
            ELSE current_setting('autovacuum_analyze_threshold')::int8
        END AS analyze_threshold,
        CASE
            WHEN relopts LIKE '%autovacuum_analyze_scale_factor%' THEN
                regexp_replace(relopts, '.*autovacuum_analyze_scale_factor=([0-9.]+).*', e'\\1')::numeric
            ELSE current_setting('autovacuum_analyze_scale_factor')::numeric
        END AS analyze_scale_factor,
        CASE
            WHEN relopts ~ 'autovacuum_enabled=(false|off)' THEN false
            ELSE true
        END AS autovacuum_enabled
    FROM
        table_opts
)
SELECT
    format('%I.%I', s.schemaname, s.relname) AS relation,
    s.n_live_tup,
    s.n_mod_since_analyze,
    round((a.analyze_threshold + a.analyze_scale_factor * greatest(c.reltuples, 0))::numeric) AS analyze_trigger,
    round((100 * s.n_mod_since_analyze::numeric / nullif(greatest(c.reltuples, 0), 0))::numeric, 2) AS modified_pct,
    s.last_analyze,
    s.last_autoanalyze,
    a.autovacuum_enabled,
    format('at: %s, asf: %s', a.analyze_threshold, a.analyze_scale_factor) AS effective_settings
FROM
    pg_stat_user_tables s
    JOIN pg_class c ON c.oid = s.relid
    JOIN analyze_settings a ON a.oid = s.relid
WHERE
    s.n_mod_since_analyze > a.analyze_threshold + a.analyze_scale_factor * greatest(c.reltuples, 0)
    OR s.last_analyze IS NULL
       AND s.last_autoanalyze IS NULL
ORDER BY
    s.n_mod_since_analyze DESC,
    relation;
