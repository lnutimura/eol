<!--
Title: a Conventional Commit header, e.g. `feat(check): add --warn-within`.
Lowercase, imperative, 72 characters at most (aim for 50). Mark changes
to exit codes or to JSON/YAML/CSV output with `!`, e.g. `feat(output)!:`.

Keep every heading. Tick a box only for a command you ran on this branch;
otherwise leave it unticked and say why below it.
-->

## What and why

<!-- What changes and the problem it solves. Link the issue: `Closes #123`. -->

## Checks

- [ ] `go vet ./...` and `go test -race ./...`
- [ ] `golangci-lint run`
- [ ] Tried the change against the live API: <!-- commands you ran -->

## Compatibility

- [ ] Exit codes are unchanged (0 supported, 1 error, 2 end-of-life)
- [ ] JSON and YAML still emit the API `result` payload; default columns are unchanged
- [ ] New commands, flags and columns are in the README and in `--help`
- [ ] Argument completions still work (`eol __complete ...`)

## Output

<!-- Before and after for any change to what eol prints. Retake the
README screenshots in `docs/screenshots/` if they no longer match.
Delete if the output is unchanged. -->
