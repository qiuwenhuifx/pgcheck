package app

var checkpointNotes = []string{
	"Background writer and checkpointer should usually be the main writers of dirty pages.",
	"If backend_total_written is much larger than the other values, review bgwriter and checkpoint tuning.",
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

var indexCreateNotes = []string{
	"Progress is sampled five times, one second apart.",
	"CREATE INDEX CONCURRENTLY can wait for writers, validation, old snapshots, and readers during different phases.",
}

var indexLowNotes = []string{
	"Low scan counts can indicate unused indexes, but always check whether the index backs a constraint or rare critical query.",
}

var lockNotes = []string{
	"PostgreSQL lock capacity is affected by max_connections and max_locks_per_transaction.",
	"Pay attention to out of shared memory errors when lock usage is high.",
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

var relConstraintNotes = []string{
	"Dropping a column can automatically drop related composite indexes and constraints; review dependencies before DDL.",
}

var vacuumStateNotes = []string{
	"phase shows the current internal VACUUM stage.",
	"heap block counters help estimate scan and cleanup progress.",
	"dead tuple counters are bounded by autovacuum_work_mem or maintenance_work_mem.",
}
