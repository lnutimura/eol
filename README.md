# eol

A command-line client for [endoflife.date](https://endoflife.date) support lifecycles.

[![CI](https://img.shields.io/github/actions/workflow/status/lnutimura/eol/ci.yml?branch=main)](https://github.com/lnutimura/eol/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/lnutimura/eol)](https://github.com/lnutimura/eol/releases/latest)
[![License](https://img.shields.io/github/license/lnutimura/eol)](LICENSE)

`eol` looks up when a product, OS, or runtime stops receiving updates, right from
the terminal. Data comes from the public endoflife.date API.

## Quick start

```bash
eol product get ubuntu
eol check ubuntu@18.04
eol product list --category os -o csv
```

```
Name:         ubuntu
Label:        Ubuntu
Category:     os
Tags:         linux-distribution, os
Version:      cat /etc/os-release
EOL label:    Maintenance & Security Support
HTML:         https://endoflife.date/ubuntu

NAME   LABEL                           RELEASE DATE  EOL FROM    EOES FROM
26.04  26.04 'Resolute Raccoon' (LTS)  2026-04-23    2031-05-29  2036-04-23
24.04  24.04 'Noble Numbat' (LTS)      2024-04-25    2029-05-31  2034-04-25
22.04  22.04 'Jammy Jellyfish' (LTS)   2022-04-21    2027-06-01  2032-04-21
```

## Install

Download a Linux amd64 build (macOS, Windows, and arm64 archives are on the
[Releases](https://github.com/lnutimura/eol/releases/latest) page):

```bash
curl -sSL https://github.com/lnutimura/eol/releases/download/v0.1.0/eol_0.1.0_linux_amd64.tar.gz \
  | tar -xz
sudo mv eol /usr/local/bin/
eol --version
```

With Go 1.25 or later:

```bash
go install github.com/lnutimura/eol@latest
```

## Usage

| Command | Purpose |
| --- | --- |
| `eol product` | List products, or show one product and its releases |
| `eol release` | List, get, or fetch the latest release cycle |
| `eol category` | List categories, or products in a category |
| `eol tag` | List tags, or products with a tag |
| `eol identifier` | List identifier types, or identifiers of a type |
| `eol index` | List the main API endpoints |
| `eol check` | Exit 0 if supported, 2 if end-of-life |

### Check support (CI)

Bare `check` uses the latest cycle. `ubuntu@22.04.1` resolves to cycle `22.04`
by longest prefix. Exit `1` on request errors.

```bash
eol check ubuntu
eol check ubuntu@18.04   # exit 2 — end-of-life (eol from 2023-05-31)
```

### Inspect a product

`product get` prints labels, links, and identifiers, then the releases table.
`--full` on `product list` adds `latest`, `latestDate`, and `releases` columns.

```bash
eol product get python -c name,releaseDate,eolFrom,isEoas,eoasFrom
eol release latest nodejs -o yaml
eol category get os
```

### Script with JSON or CSV

JSON and YAML emit the API `result` payload, so `jq` works without unwrapping.
`--columns` / `--all` apply only to `table`, `csv`, `tsv`, and `markdown`.

```bash
eol product list --category os -o csv
eol product list -o json | jq '.[0].name'
eol release latest nodejs --meta -o json
```

## Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-o`, `--output` | `table` | `table`, `json`, `yaml`, `csv`, `tsv`, `markdown` |
| `--color` | `auto` | `auto`, `always`, `never` (honors `NO_COLOR`) |
| `-c`, `--columns` | command default | Columns for tabular formats |
| `-a`, `--all` | off | All columns in tabular formats |
| `--meta` | off | Include `schema_version`, `generated_at`, and `last_modified` in JSON/YAML |
| `--timeout` | `10s` | HTTP timeout |
| `--cache-ttl` | `1h` | Serve cached responses newer than this |
| `--no-cache` | off | Bypass the on-disk cache |

On a TTY, `table` is colored. Through a pipe it prints plain aligned columns
for `awk`/`cut`. `--color always` keeps the pretty table.

## Completions

```bash
# bash
eol completion bash > /etc/bash_completion.d/eol

# zsh
eol completion zsh > "${fpath[1]}/_eol"

# fish
eol completion fish > ~/.config/fish/completions/eol.fish
```

Product, category, tag, and identifier-type arguments complete from cache.

## Cache

Responses are stored under `$XDG_CACHE_HOME/eol` (or `~/.cache/eol`) for one
hour by default. Pass `--no-cache` to always hit the network.

## License

[MIT](LICENSE). Product data is provided by
[endoflife.date](https://endoflife.date).
