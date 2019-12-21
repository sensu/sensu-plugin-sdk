# Changelog

All notable changes to this project will be documented in this file. This
changelog format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Added CHANGELOG.md

- Added support for Check plugins

### Changed
- Removed redundant `basePlugin.eventMandatory` struct field which duplicates
  the purpose of the `basePlugin.readEvent` field (i.e. used to indicate whether
  the plugin should wait to read from stdin; this `bool` should always be `true`
  for Handler and Mutator plugins, but it is optional for Check plugins)

- Refactored plugin library to remove "Go" function and type name prefixes (i.e.
  "GoHandler" is now just "Handler")

- Renamed "New" functions to "Init" functions (e.g. "NewGoHandler" is now more
  appropriately named "InitHandler")

  - `NewGoHandler` is now `InitHandler`
  - `NewGoMutator` is now `InitMutator`

  _NOTE: this is a backwards-incompatible change. Please update your plugins to
  initialize using the "sensu.InitHandler" method instead of the
  "sensu.NewGoHandler" method._
