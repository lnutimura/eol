# `eol` — a CLI for endoflife.date

Ever wondered **when your favorite software will stop receiving updates** or
reach its _end of life?_ `eol` queries the [endoflife.date](https://endoflife.date)
API so you can check support lifecycles from the terminal.

## Install

```bash
go install github.com/lnutimura/eol@latest
```

## Global flags

| Flag | Default | Description |
| --- | --- | --- |
| `-o`, `--output` | `table` | `table`, `json`, `yaml`, `csv`, `tsv`, `markdown` |
| `--color` | `auto` | `auto`, `always`, `never` (honors `NO_COLOR`) |
| `-c`, `--columns` | command default | Columns for tabular formats |
| `-a`, `--all` | off | All columns in tabular formats |
| `--meta` | off | Wrap JSON/YAML with `schema_version`, `generated_at`, `last_modified` |
| `--timeout` | `10s` | HTTP timeout |
| `--cache-ttl` | `1h` | Serve fresh on-disk cache without hitting the network |
| `--no-cache` | off | Bypass the ETag cache |

`table` uses a colored lipgloss table on a TTY. Through a pipe it degrades to
plain aligned columns so `awk`/`cut` still work. `--color always` forces the
pretty table.

JSON and YAML emit the API `result` payload so `| jq '.[0].name'` works.
`--columns` / `--all` apply only to tabular formats.

## Commands

### `eol product list`

List products. `--full` uses `/products/full` and unlocks `latest`,
`latestDate`, and `releases` columns.

```bash
eol product list
eol product list --category os -o csv
eol product list --tag canonical
eol product list --full -c name,latest,latestDate,releases
```

### `eol product get <product>`

Print a detail panel (labels, links, identifiers, version command) followed by
the releases table. JSON/YAML include the full product object.

```bash
eol product get ubuntu
eol product get nodejs -a
eol product get python -c name,releaseDate,eolFrom,isEoas,eoasFrom
```

### `eol release list|get|latest`

```bash
eol release list ubuntu
eol release get ubuntu 22.04
eol release latest nodejs -o yaml
```

### `eol category list|get`

```bash
eol category list
eol category get os
```

### `eol tag list|get`

```bash
eol tag list
eol tag get canonical
```

### `eol identifier list|get`

```bash
eol identifier list
eol identifier get purl
```

### `eol index`

List the main API endpoints (`/`).

### `eol check <product>[@<release>]`

Exits `0` if the cycle is still in standard support, `2` if it is
end-of-life (`isEol` or unmaintained), and `1` on error. Bare `<product>`
checks the latest cycle. An unknown `@release` falls back to the longest
matching cycle prefix (`ubuntu@22.04.1` → `22.04`).

```bash
eol check ubuntu
eol check ubuntu@18.04   # exit 2
```

## Cache

Responses are stored under `$XDG_CACHE_HOME/eol` (or `~/.cache/eol`) keyed by
the request URL. Within `--cache-ttl` the CLI serves the file with no network.
After that it revalidates with `If-None-Match` and treats `304` as a hit.

Shell completions for product, category, tag, and identifier-type arguments
are served from this cache.
