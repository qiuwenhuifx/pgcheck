package app

var checkpointNotes = []string{
	"Background writer and checkpointer should usually be the main writers of dirty pages.",
	"If backend_total_written is much larger than the other values, review bgwriter and checkpoint tuning.",
}

var analyzeNeedNotes = []string{
	"Fresh statistics are required for stable query plans.",
	"Run ANALYZE after bulk loads, major data changes, or major upgrades.",
}

var freezeNotes = []string{
	"frozen_xid_age shows database-level transaction ID age.",
	"consumed_txid_pct shows how much of the two-billion transaction ID safety window has been consumed.",
	"remaining_aggressive_vacuum shows distance to autovacuum_freeze_max_age; negative values deserve attention.",
}

var indexBloatNotes = []string{
	"idx_scans is the number of index scans; low values may indicate low value, but unique and constraint indexes still matter.",
	"bloat_pct is the estimated bloat percentage.",
	"bloat_mb is the estimated wasted index space.",
}

var ioNotes = []string{
	"pg_stat_io is available in PostgreSQL 16+.",
	"Use this to compare relation, temp, and WAL I/O across backend, bgwriter, checkpointer, and autovacuum.",
}

var ioHotspotNotes = []string{
	"Rows are ordered by accumulated I/O time and operation count.",
	"High checkpointer fsync or write time can indicate checkpoint spikes; high temp I/O can indicate work_mem pressure.",
}

var indexCreateNotes = []string{
	"Progress is sampled five times, one second apart.",
	"CREATE INDEX CONCURRENTLY can wait for writers, validation, old snapshots, and readers during different phases.",
}

var indexLowNotes = []string{
	"Low scan counts can indicate unused indexes, but always check whether the index backs a constraint or rare critical query.",
}

var indexNullFracNotes = []string{
	"High null_frac on a large single-column index may indicate a candidate for a partial index.",
	"Review query predicates before changing or dropping any index.",
}

var intPKRiskNotes = []string{
	"int2 and int4 primary keys can exhaust their value range on old or fast-growing tables.",
	"Rows with high capacity_used_pct should be planned for bigint migration before an incident window.",
}

var unusedIndexesNotes = []string{
	"Statistics must be old enough to represent real workload before removing any index.",
	"Check primary and read replicas before treating an index as truly unused.",
}

var lockNotes = []string{
	"PostgreSQL lock capacity is affected by max_connections and max_locks_per_transaction.",
	"Pay attention to out of shared memory errors when lock usage is high.",
}

var privilegeNotes = []string{
	"pg_monitor is the recommended role for broad monitoring visibility.",
	"Grant pg_read_all_stats or pg_read_all_settings when a narrower privilege is enough.",
}

var longTransactionNotes = []string{
	"Long transactions can block vacuum cleanup and cause table bloat.",
	"They can make new indexes confusing to validate or use, especially with concurrent index builds.",
	"They can also retain WAL for logical decoding, CDC, and standby replay.",
}

var relationBloatNotes = []string{
	"Table bloat is estimated from statistics; run ANALYZE first for better accuracy.",
	"Old xmin and old transactions can block vacuum from reclaiming dead tuples.",
}

var replicationSlotsNotes = []string{
	"xmin and catalog_xmin retained by replication slots can block vacuum cleanup.",
	"retained_wal shows WAL kept since restart_lsn; inactive slots with large retained WAL need attention.",
}

var tempFilesNotes = []string{
	"Current temporary files usually indicate large sorts, hashes, materialized CTEs, or insufficient work_mem for a query.",
	"This check requires enough privilege to execute pg_ls_dir and pg_stat_file, or a monitoring role such as pg_monitor.",
}

var walHealthNotes = []string{
	"Growing pg_wal is often caused by inactive or lagging replication slots, archiver failure, or large WAL retention settings.",
	"Use replication_slots and wal_archive for deeper follow-up when this summary looks suspicious.",
}

var xidWraparoundNotes = []string{
	"Monitor both XID and MultiXID age; MultiXID risk is often missed by simpler checks.",
	"Large relation-level ages usually mean freeze work is falling behind or is blocked by xmin horizon.",
}

var relConstraintNotes = []string{
	"Dropping a column can automatically drop related composite indexes and constraints; review dependencies before DDL.",
}

var vacuumStateNotes = []string{
	"phase shows the current internal VACUUM stage.",
	"heap block counters help estimate scan and cleanup progress.",
	"dead tuple counters are bounded by autovacuum_work_mem or maintenance_work_mem.",
}

var vacuumQueueNotes = []string{
	"Rows marked as in queue have crossed effective autovacuum thresholds but are not currently being vacuumed.",
	"Use this with xmin_blockers and long_transaction when dead tuples keep accumulating.",
}

var xminHorizonNotes = []string{
	"xmin_horizon_age highlights the oldest xmin retained by sessions, replication slots, standbys, or prepared transactions.",
	"A large value can explain vacuum not reclaiming dead tuples and table bloat growth.",
}
