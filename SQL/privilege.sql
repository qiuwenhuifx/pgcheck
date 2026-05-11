SELECT
    current_user AS current_user,
    rolsuper AS superuser,
    pg_has_role(current_user, 'pg_monitor', 'member') AS pg_monitor,
    pg_has_role(current_user, 'pg_read_all_stats', 'member') AS pg_read_all_stats,
    pg_has_role(current_user, 'pg_read_all_settings', 'member') AS pg_read_all_settings,
    rolreplication AS replication_privilege,
    CASE
        WHEN to_regprocedure('pg_ls_waldir()') IS NULL THEN false
        ELSE has_function_privilege(current_user, 'pg_ls_waldir()', 'execute')
    END AS can_execute_pg_ls_waldir,
    CASE
        WHEN to_regprocedure('pg_ls_dir(text, boolean, boolean)') IS NULL THEN false
        ELSE has_function_privilege(current_user, 'pg_ls_dir(text, boolean, boolean)', 'execute')
    END AS can_execute_pg_ls_dir,
    CASE
        WHEN to_regprocedure('pg_stat_file(text, boolean)') IS NULL THEN false
        ELSE has_function_privilege(current_user, 'pg_stat_file(text, boolean)', 'execute')
    END AS can_execute_pg_stat_file
FROM
    pg_roles
WHERE
    rolname = current_user;
