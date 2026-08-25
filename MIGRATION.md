# Migrating to v0.4.0

## Compatibility change

`goio` v0.4.0 raises the module's minimum Go version from Go 1.18 to Go 1.26. This is an intentional compatibility-breaking change. It does not change exported Go API signatures or effect-runtime semantics.

The maintained release line supports the two most recent stable Go release families. At the time of the v0.4.0 decision, those families are Go 1.26 and Go 1.27.

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

The v0.3.7 tag remains available, but the maintained release line moves to the rolling support policy described above.

## Future minimum-version changes

When a new stable Go family causes the older supported family to leave the two-release window, the next planned `goio` release may raise its minimum version. Each increase must be explicit in `go.mod`, CI, release notes, and migration guidance; it is not performed automatically by an unpinned `stable` CI alias.

Dependency upgrades and source simplifications enabled by Go 1.26 are intentionally separate follow-up changes so they can be reviewed and rolled back independently.
