# `eol` — a cli for endoflife.date

Ever wondered **when your favorite software will stop receiving updates** or reach
its _end of life?_ Thanks to the folks at endoflife.date, checking that
information has never been easier!

That’s where `eol` comes in — a simple command-line tool that uses the
endoflife.date API to give you quick and easy access to this information, right
from your terminal. Whether you’re checking the support status of a language,
framework, or operating system, `eol` makes it effortless to stay informed.

## Commands

### `eol product list`

Lists all products referenced on endoflife.date.

**Usage:**

```bash
eol product list [flags]
```

**Flags:**

- `-a`, `--all`  
  Display all available columns.

- `-c`, `--columns`  
  Comma-separated list of columns to display (e.g., `Name,Category,Tags`).

- `--category <category>`  
  Filter the product list by category.

**Example:**

```bash
eol product list
eol product list -a
eol product list -c Name,Category
eol product list --category os
```

---

### `eol product get <product>`

Retrieve detailed information about a specific product.

**Usage:**

```bash
eol product get <product> [flags]
```

**Arguments:**

- `<product>`: The exact name of the product to query.

**Flags:**

- `-a`, `--all`  
  Display all available columns for product releases.

- `-c`, `--columns`  
  Comma-separated list of columns to display for product releases.

**Example:**

```bash
eol product get ubuntu
eol product get nodejs -a
eol product get python -c Name,ReleaseDate,EolFrom
```

---

### `eol category list`

Lists all categories available on endoflife.date.

**Usage:**

```bash
eol category list [flags]
```

**Flags:**

- `-a`, `--all`  
  Display all available columns.

- `-c`, `--columns`  
  Comma-separated list of columns to display.

**Example:**

```bash
eol category list
eol category list -a
eol category list -c Name,URI
```
