# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.0] - 2022-11-10

### Added

- The `lg.Log` interface now has `With` methods for structured logging.

### Removed

- `loglg` (stdlib `log`) support has been removed. It's not worth maintaining this
   given the development of the [slog](https://pkg.go.dev/golang.org/x/exp@v0.0.0-20221110155412-d0897a79cd37/slog) experiment.


## [v1.0.0] - 2022-11-09

### Changed

- `v1.0.0` release.


[v2.0.0]: https://github.com/neilotoole/lg/compare/v1.0.0...v2.0.0
[v1.0.0]: https://github.com/neilotoole/lg/releases/tag/v0.0.3
