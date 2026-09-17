# Migrating to v0.5.0

## Go 1.27 requirement

`goio` v0.5.0 raises the module's minimum Go version from Go 1.26 to Go 1.27. Projects that must remain on Go 1.26 can continue using the v0.4.x release line.

Generic methods are additive. Existing calls to package functions remain supported, so adopting the fluent method syntax can be incremental:

```go
// Existing form.
mapped := io.Map(source, transform)

// Equivalent Go 1.27 method form.
mapped := source.Map(transform)
```

The method implementations delegate to the established package combinators and do not introduce new runtime semantics.

---

# Migrating to v0.4.0

## Compatibility change

`goio` v0.4.0 raises the module's minimum Go version from Go 1.18 to Go 1.26. This is an intentional compatibility-breaking change. The release also renames `option.Fold` to `option.Match`; other source changes primarily stabilize existing effect-runtime behavior.

The v0.4.x release line supports Go 1.26 and Go 1.27. It is the final release line supporting Go 1.26; development after v0.4.0 targets Go 1.27 so that goio can adopt generic methods.

## Upgrade steps

1. Install the latest patch release of Go 1.26 or Go 1.27.
2. Update local, CI, container, and release toolchains that build modules depending on `goio`.
3. Upgrade the dependency:

   ```sh
   go get github.com/primetalk/goio@v0.4.0
   go mod tidy
   ```

4. Run the downstream project's normal tests and race-enabled tests where supported:

   ```sh
   go test ./...
   go test -race ./...
   ```

No source migration is expected solely because of this version-floor change.

## Remaining on the former floor

Projects that cannot yet move from Go 1.18 through Go 1.25 can remain on the last release with the former module floor:

```sh
go get github.com/primetalk/goio@v0.3.7
```

The v0.3.7 tag remains available for projects that cannot yet adopt Go 1.26.

## Development after v0.4.0

The next planned goio release raises its minimum version to Go 1.27. The increase will be explicit in `go.mod`, CI, release notes, and migration guidance; it is not performed automatically by an unpinned `stable` CI alias.

Dependency upgrades and source simplifications enabled by Go 1.26 are intentionally separate follow-up changes so they can be reviewed and rolled back independently.
