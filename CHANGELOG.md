# Changelog

## [1.0.1]

- Add `--json` / `-j` flag to write results to a JSON file (#45)
- Update Go dependencies

## [1.0.0]

- Optimize key generation by reusing buffers to reduce heap allocations
- Add comprehensive unit tests for key generation and utility functions
- Add GitHub Actions workflow for automated testing on push and pull requests
- Update Go to 1.25.0
- Add `-v` / `--version` flag to display current version and check for updates
- Add `-u` / `--update` flag to self-update to the latest release

## [0.1.3]

- Switch to `github.com/oasisprotocol/curve25519-voi` for a ~18% speed improvement
- Refactor key management to use AtomicCounter for thread-safe operations in Cruncher
- Add benchmark test
- Update Go dependencies
- Code cleanup
- Update Dependabot schedule to semiannually

## [0.1.2]

- Update Go dependencies

## [0.1.1]

- Update Go dependencies

## [0.1.0]

- Add timeout option (#16)

## [0.0.9]

- Add regex support (#13)
- Update Go dependencies

## [0.0.8]

- Update Go dependencies
- Update Github workflow versions

## [0.0.7]

- Add Github actions for releases & updates
- Update library function naming
- Update Go dependencies
- Remove Makefile (#3)

## [0.0.6]

- Update Go dependencies

## [0.0.5]

- Allow for inclusion as a module (#2)

## [0.0.4]

- Update Go dependencies

## [0.0.3]

- Update Go dependencies
- Set default cores to all available minus 1

## [0.0.2]

- Show duration estimate for weeks, months and years

## [0.0.1]

- First release
