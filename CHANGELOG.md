# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

## [0.6.0] - 2020-02-06

### Added
- Added helpers for TLS configuration.

### Fixed
- Logs an error if the plugin fails to initialize.
- Prevent duplicated error messages fix the formatting.
- Fixed a bug that could result in a panic when CA certificate is specified.

## [0.5.0] - 2020-02-05

### Added
- Added package httpclient.
- Added documentation to a few packages.
- Added package version.

### Fixed
- Fixed a panic that could occur on nil checks or entities.
- Cleaned up argument parsing.

### Removed
- Removed package http.
