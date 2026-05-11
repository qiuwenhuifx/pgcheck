# pgcheck

English | [中文](#中文)

`pgcheck` is a lightweight PostgreSQL health-check CLI for DBAs, SREs, and database engineers. It collects operational signals from PostgreSQL system catalogs and statistics views, including locks, wait events, replication, vacuum, transaction ID age, relation bloat, index health, WAL archiving, partitions, TOAST tables, and object ownership.

The project started as a Bash-based one-click inspection script. It is now being refactored into a structured Go project with embedded SQL assets, explicit command registration, server-version detection, and a cleaner compatibility model.

## Highlights

- Simple single-binary CLI written in Go.
- Uses PostgreSQL's standard `psql` connection behavior and environment variables.
- Embeds SQL checks into the binary with Go `embed`.
- Detects PostgreSQL server version instead of relying on client version.
- Keeps each check as a registered command, making the project easier to extend and test.
- Preserves the original SQL assets under `SQL/` for review and reuse.

## Compatibility

The original tool was tested up to PostgreSQL 15. The Go refactor keeps PostgreSQL 11-15 as the primary compatibility target and adds version-aware behavior for newer PostgreSQL releases where possible.

PostgreSQL 17+ changed some statistics views, including checkpoint and VACUUM progress views. `pgcheck` now ships version-specific SQL for those checks. Most read-only checks have been smoke-tested against a PostgreSQL 17+ public test instance; `wal_generate` still requires a valid server-side `pg_wal` path and should be verified per deployment.

## Requirements

- Go 1.23+ for building from source.
- PostgreSQL client tools, especially `psql`, available in `PATH` when using the default `psql` backend.
- A PostgreSQL role with enough privileges to read the relevant catalog and statistics views.

`pgcheck` can read connection and display settings from a JSON config file, command-line options, or libpq-compatible environment variables. Command-line options have the highest priority.

Use a config file when you do not want to export environment variables:

```bash
cp pgcheck.example.json pgcheck.json
bin/pgcheck --config pgcheck.json dbstatus
bin/pgcheck --config pgcheck.json --display expanded dbstatus
```

Config shape:

```json
{
  "_comment": "Keys starting with _comment are documentation only and ignored by pgcheck.",
  "backend": "psql",
  "connection": {
    "_comment": "Equivalent to PGHOST, PGPORT, PGUSER, PGPASSWORD, PGDATABASE and PGSSLMODE.",
    "host": "127.0.0.1",
    "port": "5432",
    "user": "postgres",
    "password": "",
    "database": "postgres",
    "sslmode": "disable"
  },
  "psql": {
    "_comment": "These options mirror psql command-line flags and mainly affect the psql backend.",
    "path": "psql",
    "_quiet": "quiet=true is the same as psql -q: reduce non-data messages.",
    "quiet": true,
    "_tuples_only": "tuples_only=true is the same as psql -t: show rows without headers/footers.",
    "tuples_only": false,
    "_no_align": "no_align=true is the same as psql -A: use unaligned output.",
    "no_align": false,
    "_no_psqlrc": "no_psqlrc=true is the same as psql -X: ignore ~/.psqlrc for stable output.",
    "no_psqlrc": true,
    "_extra_args": "extra_args are appended directly to psql, for example [\"--csv\"] or [\"--set\", \"ON_ERROR_STOP=1\"].",
    "extra_args": []
  },
  "output": {
    "_expanded": "auto keeps command defaults; table forces normal table output; expanded forces psql -x style output.",
    "expanded": "auto"
  }
}
```

You can also use environment variables:

```bash
export PGHOST=127.0.0.1
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=secret
```

If `psql` is not in `PATH`, set it in the config file or pass `--psql`:

```bash
bin/pgcheck --psql /usr/local/pgsql/bin/psql dbstatus
```

`pgcheck` also supports a native Go driver backend, which is useful on hosts without `psql`:

```bash
bin/pgcheck --backend native --host 127.0.0.1 --port 5432 --user postgres --password secret dbstatus
```

Common psql-style display options are configurable:

```bash
bin/pgcheck --display expanded dbstatus
bin/pgcheck --display table dbstatus
bin/pgcheck --tuples-only --no-align connections postgres
bin/pgcheck --psql-arg --single-transaction relation postgres public
```

Option notes for users familiar with `psql`:

| Config key | CLI option | psql flag | Meaning |
| --- | --- | --- | --- |
| `quiet` | `--quiet` / `--no-quiet` | `-q` | Reduce non-data messages printed by `psql`. |
| `tuples_only` | `--tuples-only` | `-t` | Print rows without column headers and footers. |
| `no_align` | `--no-align` | `-A` | Use unaligned output instead of padded table output. |
| `no_psqlrc` | `--no-psqlrc` | `-X` | Do not read `~/.psqlrc`, making output more predictable. |
| `extra_args` | `--psql-arg` | passthrough | Append raw arguments to `psql`, such as `--csv` or `--set name=value`. |

## Build

```bash
go build -o bin/pgcheck .
```

or:

```bash
make build
```

Show help:

```bash
bin/pgcheck help
```

Show version:

```bash
bin/pgcheck version
```

## Commands

```text
pgcheck relation <database> <schema>         List table and index size in a schema
pgcheck relconstraint <database> <relation>  List constraints and multi-column indexes for a relation
pgcheck alltoast <database> <schema>         List TOAST tables in a schema
pgcheck reltoast <database> <relation>       Show TOAST-related columns for a relation
pgcheck dbstatus                             Show database-level statistics
pgcheck index_bloat <database>               Estimate btree index bloat
pgcheck index_duplicate <database>           Find duplicate indexes
pgcheck index_low <database>                 Find low-efficiency indexes
pgcheck index_state <database>               Show index details and invalid indexes
pgcheck lock <database>                      Show lock waits and blocking queue
pgcheck checkpoint                           Show background writer and checkpointer statistics
pgcheck freeze <database>                    Show transaction ID consumption and freeze risk
pgcheck replication                          Show physical streaming replication status
pgcheck connections <database>               Show connection summary and active queries
pgcheck long_transaction <database>          Show long-running transactions
pgcheck relation_bloat <database>            Estimate table bloat and vacuum blockers
pgcheck vacuum_state <database>              Show running VACUUM progress
pgcheck vacuum_need <database>               Show tables likely to need vacuum
pgcheck index_create <database>              Show CREATE INDEX progress
pgcheck wal_archive                          Show WAL archiver statistics
pgcheck wal_generate <wal_path>              Show WAL generation speed by scanning pg_wal
pgcheck wait_event <database>                Show wait events and wait event types
pgcheck partition <database>                 Show partition information
pgcheck object <database> <user>             Show objects owned by a user and role membership
```

## Examples

```bash
bin/pgcheck dbstatus
bin/pgcheck --config pgcheck.json dbstatus
bin/pgcheck --backend native --host 127.0.0.1 --user postgres dbstatus
bin/pgcheck connections postgres
bin/pgcheck lock postgres
bin/pgcheck freeze postgres
bin/pgcheck relation postgres public
bin/pgcheck index_bloat postgres
bin/pgcheck wal_generate /var/lib/postgresql/data/pg_wal
```

## Project Layout

```text
.
├── main.go                 Go entrypoint and SQL embedding
├── internal/
│   ├── app/                CLI commands, version-aware checks, command execution flow
│   ├── pgexec/             psql runner and PostgreSQL server version detection
│   └── queries/            embedded SQL loader and small templating helpers
├── SQL/                    original SQL check assets
├── pgcheck.example.json    example configuration file
├── pgcheck.sh              legacy Bash implementation
└── README.md
```

## Design Notes

The current Go implementation keeps `psql` as the default execution backend. This preserves standard PostgreSQL behavior, including `.pgpass`, service files, SSL options, psql formatting, and existing environment variables.

The project also includes an optional native Go backend based on `database/sql` and `github.com/lib/pq`. Use `--backend native` when `psql` is unavailable or when you prefer not to shell out.

## Development

Run tests:

```bash
go test ./...
```

Format code:

```bash
gofmt -w main.go internal/**/*.go
```

Build:

```bash
go build -o bin/pgcheck .
```

## Roadmap

- Add automated PostgreSQL 11-18 compatibility tests with containers.
- Add structured output formats such as JSON and Markdown.
- Add severity classification for health-check results.
- Add native Go database driver execution mode.
- Add release artifacts for Linux/macOS on amd64 and arm64.

## License

Apache License 2.0. See [LICENSE](LICENSE).

## 中文

`pgcheck` 是一款轻量级 PostgreSQL 巡检 CLI，面向 DBA、SRE 和数据库工程师。它通过 PostgreSQL 系统表、系统视图和统计视图采集运行状态，覆盖锁等待、等待事件、复制、VACUUM、事务 ID 年龄、表膨胀、索引健康、WAL 归档、分区表、TOAST 表和对象归属等常见运维场景。

这个项目最早是一个 Bash 编写的一键巡检脚本。当前版本正在重构为结构化的 Go 项目：SQL 资源会被嵌入二进制，命令通过注册表管理，版本判断基于 PostgreSQL 服务端版本，并提供更清晰的兼容策略。

## 亮点

- 使用 Go 编写，构建后是一个简单的单文件 CLI。
- 复用 PostgreSQL 标准 `psql` 连接行为和环境变量。
- 使用 Go `embed` 将 SQL 巡检资源嵌入二进制。
- 检测 PostgreSQL 服务端版本，而不是依赖本地客户端版本。
- 每个巡检项都是独立注册的命令，后续扩展和测试更容易。
- 保留原始 `SQL/` 目录，方便审阅、复用和继续沉淀 SQL 资产。

## 兼容性

原始工具主要测试到 PostgreSQL 15。Go 重构版本继续将 PostgreSQL 11-15 作为主要兼容目标，同时尽量为更高版本保留版本感知能力。

PostgreSQL 17+ 对部分统计视图做了调整，例如 checkpoint 和 VACUUM progress 相关视图。`pgcheck` 现在已经为这些检查提供了版本专用 SQL。大部分只读巡检命令已经在 PostgreSQL 17+ 公网测试实例上完成冒烟验证；`wal_generate` 仍然依赖服务端 `pg_wal` 物理路径，需要按具体部署单独验证。

## 环境要求

- 从源码构建需要 Go 1.23+。
- 使用默认 `psql` 后端时，本机需要安装 PostgreSQL 客户端工具，并确保 `psql` 在 `PATH` 中。
- 巡检用户需要有读取相关系统视图和统计视图的权限。

`pgcheck` 支持从 JSON 配置文件、命令行参数或 libpq 兼容环境变量读取连接和展示设置。命令行参数优先级最高。

如果不想 export 一堆环境变量，可以使用配置文件：

```bash
cp pgcheck.example.json pgcheck.json
bin/pgcheck --config pgcheck.json dbstatus
bin/pgcheck --config pgcheck.json --display expanded dbstatus
```

配置结构：

```json
{
  "_comment": "以 _comment 开头的字段仅用于说明，pgcheck 会忽略它们。",
  "backend": "psql",
  "connection": {
    "_comment": "等价于 PGHOST、PGPORT、PGUSER、PGPASSWORD、PGDATABASE 和 PGSSLMODE。",
    "host": "127.0.0.1",
    "port": "5432",
    "user": "postgres",
    "password": "",
    "database": "postgres",
    "sslmode": "disable"
  },
  "psql": {
    "_comment": "这些选项和 psql 命令行参数保持一致，主要影响 psql 后端。",
    "path": "psql",
    "_quiet": "quiet=true 等价于 psql -q：减少非数据类输出。",
    "quiet": true,
    "_tuples_only": "tuples_only=true 等价于 psql -t：只输出行数据，不输出表头和页脚。",
    "tuples_only": false,
    "_no_align": "no_align=true 等价于 psql -A：使用非对齐输出。",
    "no_align": false,
    "_no_psqlrc": "no_psqlrc=true 等价于 psql -X：不读取 ~/.psqlrc，让输出更稳定。",
    "no_psqlrc": true,
    "_extra_args": "extra_args 会直接透传给 psql，例如 [\"--csv\"] 或 [\"--set\", \"ON_ERROR_STOP=1\"]。",
    "extra_args": []
  },
  "output": {
    "_expanded": "auto 保留命令默认展示；table 强制普通表格；expanded 强制 psql -x 风格展示。",
    "expanded": "auto"
  }
}
```

也可以继续使用环境变量：

```bash
export PGHOST=127.0.0.1
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=secret
```

如果 `psql` 不在 `PATH` 中，可以在配置文件里设置，或通过 `--psql` 指定：

```bash
bin/pgcheck --psql /usr/local/pgsql/bin/psql dbstatus
```

`pgcheck` 也支持原生 Go driver 后端，适合没有安装 `psql` 的环境：

```bash
bin/pgcheck --backend native --host 127.0.0.1 --port 5432 --user postgres --password secret dbstatus
```

常见 psql 展示选项也可以配置：

```bash
bin/pgcheck --display expanded dbstatus
bin/pgcheck --display table dbstatus
bin/pgcheck --tuples-only --no-align connections postgres
bin/pgcheck --psql-arg --single-transaction relation postgres public
```

给熟悉 `psql` 的用户看的对应关系：

| 配置项 | CLI 参数 | psql 参数 | 含义 |
| --- | --- | --- | --- |
| `quiet` | `--quiet` / `--no-quiet` | `-q` | 减少 `psql` 的非数据类输出。 |
| `tuples_only` | `--tuples-only` | `-t` | 只输出行数据，不输出列名和页脚。 |
| `no_align` | `--no-align` | `-A` | 使用非对齐输出，而不是补齐宽度的表格输出。 |
| `no_psqlrc` | `--no-psqlrc` | `-X` | 不读取 `~/.psqlrc`，避免本地 psql 配置影响巡检输出。 |
| `extra_args` | `--psql-arg` | 透传 | 直接追加原始 psql 参数，例如 `--csv` 或 `--set name=value`。 |

## 构建

```bash
go build -o bin/pgcheck .
```

也可以使用：

```bash
make build
```

查看帮助：

```bash
bin/pgcheck help
```

查看版本：

```bash
bin/pgcheck version
```

## 命令列表

```text
pgcheck relation <database> <schema>         查看指定 schema 下表和索引大小
pgcheck relconstraint <database> <relation>  查看指定表的约束和多列索引
pgcheck alltoast <database> <schema>         查看指定 schema 下的 TOAST 表
pgcheck reltoast <database> <relation>       查看指定表的 TOAST 相关列和 TOAST 表信息
pgcheck dbstatus                             查看数据库整体状态
pgcheck index_bloat <database>               估算 btree 索引膨胀
pgcheck index_duplicate <database>           查找重复索引
pgcheck index_low <database>                 查找低效索引
pgcheck index_state <database>               查看索引详情和异常索引
pgcheck lock <database>                      查看锁等待和阻塞队列
pgcheck checkpoint                           查看后台写进程和检查点统计
pgcheck freeze <database>                    查看事务 ID 消耗和 freeze 风险
pgcheck replication                          查看物理流复制状态
pgcheck connections <database>               查看连接汇总和当前查询
pgcheck long_transaction <database>          查看长事务
pgcheck relation_bloat <database>            估算表膨胀并查看 VACUUM 阻塞信息
pgcheck vacuum_state <database>              查看正在运行的 VACUUM 进度
pgcheck vacuum_need <database>               查看可能需要 VACUUM 的表
pgcheck index_create <database>              查看 CREATE INDEX 进度
pgcheck wal_archive                          查看 WAL 归档状态
pgcheck wal_generate <wal_path>              基于 pg_wal 目录估算 WAL 生成速度
pgcheck wait_event <database>                查看等待事件和等待类型
pgcheck partition <database>                 查看分区表信息
pgcheck object <database> <user>             查看用户拥有的对象和角色成员关系
```

## 示例

```bash
bin/pgcheck dbstatus
bin/pgcheck --config pgcheck.json dbstatus
bin/pgcheck --backend native --host 127.0.0.1 --user postgres dbstatus
bin/pgcheck connections postgres
bin/pgcheck lock postgres
bin/pgcheck freeze postgres
bin/pgcheck relation postgres public
bin/pgcheck index_bloat postgres
bin/pgcheck wal_generate /var/lib/postgresql/data/pg_wal
```

## 项目结构

```text
.
├── main.go                 Go 入口和 SQL 嵌入
├── internal/
│   ├── app/                CLI 命令、版本兼容逻辑和执行流程
│   ├── pgexec/             psql 执行器和服务端版本检测
│   └── queries/            嵌入式 SQL 加载和轻量模板处理
├── SQL/                    原始 SQL 巡检资产
├── pgcheck.example.json    配置文件示例
├── pgcheck.sh              旧版 Bash 实现
└── README.md
```

## 设计说明

当前 Go 版本保留 `psql` 作为默认 SQL 执行后端。这样可以继承 `.pgpass`、service file、SSL 参数、psql 展示格式和环境变量等 PostgreSQL 标准能力。

项目现在也包含基于 `database/sql` 和 `github.com/lib/pq` 的原生 Go 后端。没有安装 `psql`，或者不希望通过 shell 调用外部命令时，可以使用 `--backend native`。

## 开发

运行测试：

```bash
go test ./...
```

格式化：

```bash
gofmt -w main.go internal/**/*.go
```

构建：

```bash
go build -o bin/pgcheck .
```

## 后续计划

- 使用容器补齐 PostgreSQL 11-18 的自动化兼容测试。
- 增加 JSON、Markdown 等结构化输出格式。
- 为巡检结果增加风险等级和诊断建议。
- 增加原生 Go database driver 执行模式。
- 发布 Linux/macOS amd64/arm64 构建产物。

## License

Apache License 2.0. See [LICENSE](LICENSE).
