# Changelog

## [v6.0.0] - Unreleased

### Breaking Changes

* **Module path updated** from `github.com/apolloconfig/agollo/v5` to `github.com/apolloconfig/agollo/v6`:

  ```
  go get github.com/apolloconfig/agollo/v6@latest
  ```

* New integrations use the instance-scoped `NewClient` / `ApolloClient` API.
  Configuration is read through explicit namespace views (`Config` and
  `ConfigFile`); `Load` replaces implicit `MustStart` startup behavior.
* Process-global extension setters are replaced by `ClientOptions` injection:
  `RequestSigner`, `ConfigServiceSelector`, `HTTPClient`, `CacheDir`, and
  `ConfigMapStore`. See `docs/migration-to-apollo-client.md` for the complete
  v5-to-v6 mapping.

### What's Changed

* Add multi-AppId isolation, ConfigFile views, full/incremental sync, source
  inspection, long-poll lifecycle ownership, and bounded subscriptions.
* Add typed int64, duration, string-slice, and int-slice configuration reads.
* Preserve the last successful in-memory snapshot on refresh failures; use
  local cache and ConfigMap only for initial fallback.
* Add v6 migration documentation and external public API compile coverage.

## [v5.0.0](https://github.com/apolloconfig/agollo/releases/tag/v5.0.0) - 2026-03-14

### Breaking Changes

* **Module path updated** from `github.com/apolloconfig/agollo/v4` to `github.com/apolloconfig/agollo/v5`. Update your import paths accordingly:

  ```
  go get -u github.com/apolloconfig/agollo/v5@latest
  ```

### What's Changed

* [Fix]: Avoid stop race on component stop channel ([#354](https://github.com/apolloconfig/agollo/pull/354))
* [Fix]: Fix serverIPListComponent not closing when invoking Client close ([#350](https://github.com/apolloconfig/agollo/pull/350))
* [Fix]: Add `Stoppable` interface and invoke `time.Stop()` in component `Start` method
* [Fix]: Use `sync.Once` to avoid panic on repeated Stop calls; add recover protection in `StartRefreshConfig`
* [Fix]: `NewConfigComponent()` returns `component.AbsComponent` type
* [Fix]: Compatibility handling — retain `InitSyncServerIPList` and add `NewSyncServerIPListComponent`; improve panic log format
* [Fix]: Delete `SetAppConfig` method and `SetCache` method
* [Test]: Add nil-kind coverage for `IsNilObject` ([#352](https://github.com/apolloconfig/agollo/pull/352))
* [Chore]: Upgrade Go toolchain to 1.18 ([#337](https://github.com/apolloconfig/agollo/pull/337))
* [Chore]: Update Go version to 1.20
* [Dep]: Bump `github.com/agiledragon/gomonkey/v2` from 2.11.0 to 2.13.0 ([#329](https://github.com/apolloconfig/agollo/pull/329))

**Full Changelog**: https://github.com/apolloconfig/agollo/compare/v4.4.0...v5.0.0
