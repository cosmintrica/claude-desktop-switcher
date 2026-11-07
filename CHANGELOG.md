# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0](https://github.com/cosmintrica/claude-desktop-switcher/compare/v1.2.2...v1.3.0) (2026-11-07)


### Features

* **account:** read each saved account's 5-hour and weekly limits on Windows ([ad1ff6a](https://github.com/cosmintrica/claude-desktop-switcher/commit/ad1ff6a1b4f74116589f17abb2e5eeb957a28ff4))
* add interactive profile picker and account info display ([3813d97](https://github.com/cosmintrica/claude-desktop-switcher/commit/3813d974f079f4a3a342b01b037b3879260915b9))
* add Windows Claude Swap tray and release automation ([1331cd8](https://github.com/cosmintrica/claude-desktop-switcher/commit/1331cd83e73435b7cfb3fc59384574017dcdff80))
* **cli:** initial implementation of claude-swap ([bff1ae5](https://github.com/cosmintrica/claude-desktop-switcher/commit/bff1ae57739f0f881e23fbc9f0a3498ae62a743f))
* improve Windows tray and documentation ([5a563fe](https://github.com/cosmintrica/claude-desktop-switcher/commit/5a563fe878213a82e1772a193472c6850597006e))
* make save automatic by stopping and relaunching Claude Desktop ([09ef9aa](https://github.com/cosmintrica/claude-desktop-switcher/commit/09ef9aab99b2aa2486abe2f74eab3352fe7e2a0b))
* prepare Windows Claude Swap release ([f980d2b](https://github.com/cosmintrica/claude-desktop-switcher/commit/f980d2bc8b0f40ced354af4d1cd42f01a9c2bccc))
* rename product to Claude Desktop Switcher ([7373970](https://github.com/cosmintrica/claude-desktop-switcher/commit/7373970f0ab4850acaae52003702bdecc1c8876d))
* rename to claude-desktop-swap, add interactive 'add' command, auto-prime on save ([8d68249](https://github.com/cosmintrica/claude-desktop-switcher/commit/8d68249940134bf7e975ec6e1d70e23f4e921d54))
* **sessions:** read each account's recent Claude Code sessions ([15227f4](https://github.com/cosmintrica/claude-desktop-switcher/commit/15227f42518ae8f34605cc9aa65e6adff21e2ae9))
* show live account email and plan in list and picker ([c928767](https://github.com/cosmintrica/claude-desktop-switcher/commit/c92876701bf0defbcd0d1c2ccff11eed5c275efd))
* **tray:** an aquarium above the clock, one fish per Claude window ([0b0ef83](https://github.com/cosmintrica/claude-desktop-switcher/commit/0b0ef83036a90ab49d5cd0f52ea7aa33552d9be0))
* **tray:** ask before closing Claude while sessions are working ([2ace23e](https://github.com/cosmintrica/claude-desktop-switcher/commit/2ace23ecb62227dd26e5cfe15618705ecc56577a))
* **tray:** close one Claude window at a time ([f40d9f9](https://github.com/cosmintrica/claude-desktop-switcher/commit/f40d9f9c863ff188b722fc3c3060f604d9610adc))
* **tray:** continue a recent session on its own account ([67ac73c](https://github.com/cosmintrica/claude-desktop-switcher/commit/67ac73c5de88b1c0b123b2006d214c9e4bb62b16))
* **tray:** group sessions the way Claude's sidebar does ([e6a9c64](https://github.com/cosmintrica/claude-desktop-switcher/commit/e6a9c64e1219a55414a5685c35d79574832ca7f9))
* **tray:** show each account's limits in the menu and the aquarium ([b0bf7fa](https://github.com/cosmintrica/claude-desktop-switcher/commit/b0bf7fa017db16e7e93727a380801196d6cc011d))
* **tray:** show progress as a terminal card with no background ([eb2db23](https://github.com/cosmintrica/claude-desktop-switcher/commit/eb2db239d881bb667490913645227d24840426b0))
* **tray:** start at login ([074954d](https://github.com/cosmintrica/claude-desktop-switcher/commit/074954d30e6546380fded3293f239bf1abf0b02b))
* **windows:** a fishbowl icon for the tray and the program ([5cf7f1d](https://github.com/cosmintrica/claude-desktop-switcher/commit/5cf7f1dc9a73616ef91c31d38fb742ac55f0f721))
* **windows:** open an account in a separate Claude window ([25bf7b1](https://github.com/cosmintrica/claude-desktop-switcher/commit/25bf7b1791dbe70b0fdb5a7cabfca42f1449522c))


### Bug Fixes

* **aquarium:** stay in front, and hide only for a full-screen window in front ([8cffaaf](https://github.com/cosmintrica/claude-desktop-switcher/commit/8cffaaf4c265cf38a3b2cc0eb2add0417a803c64))
* **cd:** remove unsupported changelog path field ([c2cb3db](https://github.com/cosmintrica/claude-desktop-switcher/commit/c2cb3db971363f75f54be1b781456038722277b8))
* check fork releases for updates ([c35362f](https://github.com/cosmintrica/claude-desktop-switcher/commit/c35362fbe16d55264de4a7c698572487b04bb9ff))
* check fork releases for updates ([d050c4b](https://github.com/cosmintrica/claude-desktop-switcher/commit/d050c4b5039980aa054c313c80e984b84afd4398))
* checkpointed profile switching and cookie-only v2 profiles ([#3](https://github.com/cosmintrica/claude-desktop-switcher/issues/3)) ([abda7b1](https://github.com/cosmintrica/claude-desktop-switcher/commit/abda7b1e73e17877753890e24d46ce460bebca32))
* constrain profile metadata and preserve failed rollback state ([4e180fc](https://github.com/cosmintrica/claude-desktop-switcher/commit/4e180fc034c1e69b2cfe1612cab11545cd451ef0))
* decrypt real sessionKey and read correct account API fields ([c52aaa7](https://github.com/cosmintrica/claude-desktop-switcher/commit/c52aaa7cc152e5ab92be9d4ca253c54eec5d1cf4))
* detect current Windows Claude profile and login log paths ([#14](https://github.com/cosmintrica/claude-desktop-switcher/issues/14)) ([543acf5](https://github.com/cosmintrica/claude-desktop-switcher/commit/543acf506c696bce28eb0a08e8b093817648cb20))
* make account switching on Windows more reliable ([#17](https://github.com/cosmintrica/claude-desktop-switcher/issues/17)) ([ed99588](https://github.com/cosmintrica/claude-desktop-switcher/commit/ed99588dd69efb967d8cc639813ec6e2c1a78837))
* preserve device trust across account swaps ([5cdd6c4](https://github.com/cosmintrica/claude-desktop-switcher/commit/5cdd6c4cbe16e0537e3ac679c3c1f18985e9cd23))
* **profile:** clear Cookies-journal on restore ([3a29f63](https://github.com/cosmintrica/claude-desktop-switcher/commit/3a29f6395895dd9a12ea0084443dd6f5754820d4)), closes [#1](https://github.com/cosmintrica/claude-desktop-switcher/issues/1)
* **profile:** read the Store install's log in the Local folder ([8fc751d](https://github.com/cosmintrica/claude-desktop-switcher/commit/8fc751d6a28b707b8a46e8114c57d6545eaba004))
* **profile:** retry renames that Windows refuses while a file is briefly open ([50c88dc](https://github.com/cosmintrica/claude-desktop-switcher/commit/50c88dc96f1ed8dc621392fd6f95dea8625c5d46))
* **profile:** wipe IndexedDB and Session Storage on restore ([1238355](https://github.com/cosmintrica/claude-desktop-switcher/commit/1238355292e8b95f7c3b9af206d88195b6b3fd10))
* publish Windows installers ([b3a4665](https://github.com/cosmintrica/claude-desktop-switcher/commit/b3a4665df62f270dccda42ced98f579a045723fa))
* save renewed Claude sessions automatically ([96f9783](https://github.com/cosmintrica/claude-desktop-switcher/commit/96f97831579b24a2cbe8397197bf11771b798e98))
* save renewed Claude sessions automatically ([5fda5b3](https://github.com/cosmintrica/claude-desktop-switcher/commit/5fda5b3b3e247f6508b79163305dac7be2e5104a))
* scope all progress overlays to the Claude window ([6dce87c](https://github.com/cosmintrica/claude-desktop-switcher/commit/6dce87c9a8009f62fca3570f06f76dbc94225abe))
* **tray:** ask before a switch or a new account closes Claude ([97cf4ce](https://github.com/cosmintrica/claude-desktop-switcher/commit/97cf4ceecc518e3e5b83d9922115948a66df0fe8))
* **tray:** fish from the start, the plus only under the pointer ([91d34fa](https://github.com/cosmintrica/claude-desktop-switcher/commit/91d34fa5f9b2b7afb37444b7280e02b6d70f380c))
* **tray:** keep the tray menu open while a separate window is open ([d9a5d90](https://github.com/cosmintrica/claude-desktop-switcher/commit/d9a5d905cba50bd38095017362f3aff33a66ba66))
* **tray:** pause background profile reads while profiles change ([5791ca5](https://github.com/cosmintrica/claude-desktop-switcher/commit/5791ca566ec7bf8158597e1def4a5d77ff6e2cf9))
* **tray:** say how much of each limit is used ([56865b7](https://github.com/cosmintrica/claude-desktop-switcher/commit/56865b70f385653588f2f912b1ae780a3ad82921))
* **tray:** show switching progress in a small card instead of dimming the screen ([14aa1b6](https://github.com/cosmintrica/claude-desktop-switcher/commit/14aa1b61a605b98550b68681af2bc47a5b3bafe0))
* update Go to 1.27.3 for GO-2026-6809 ([a8ed70a](https://github.com/cosmintrica/claude-desktop-switcher/commit/a8ed70a0a5bcb0a559c91f8c138bf0a95a9be2b5))
* update Go toolchain to address standard library advisories ([715a76e](https://github.com/cosmintrica/claude-desktop-switcher/commit/715a76e1df95e80c96f16c2993c274a1b1186512))
* **windows:** ask Claude to quit cleanly before stopping it ([b5e2234](https://github.com/cosmintrica/claude-desktop-switcher/commit/b5e2234d4b03b598e67175c6b36fa822b2e91f8e))
* **windows:** copy a separate window's sessions instead of linking them ([9279186](https://github.com/cosmintrica/claude-desktop-switcher/commit/92791864706f53a9c0f42b931e444aba0ec459aa))
* **windows:** follow Claude 3's data folder out of the Store package ([ea05558](https://github.com/cosmintrica/claude-desktop-switcher/commit/ea05558c3a1655b9fc62ac478daf3ffc9cd44f4e))
* **windows:** leave Claude windows with their own data folder alone ([2683a74](https://github.com/cosmintrica/claude-desktop-switcher/commit/2683a74d7c3cf409db080b3e8cd910081bd512a0))
* **windows:** open the tray when started from Explorer ([6274728](https://github.com/cosmintrica/claude-desktop-switcher/commit/62747289f2c970ac008fde174c833483a9f15f46))
* **windows:** show an account's sessions in its separate window ([e9ec14e](https://github.com/cosmintrica/claude-desktop-switcher/commit/e9ec14e8e0494e48d26bed7f467f0dd6b994f934))


### Reverts

* **windows:** stop Claude with the original close sequence ([d056fb9](https://github.com/cosmintrica/claude-desktop-switcher/commit/d056fb99734d2c1f9518b257bb4933887494d932))

## [1.2.2](https://github.com/OpenScz/claude-desktop-switcher/compare/v1.2.1...v1.2.2) (2026-10-26)


### Bug Fixes

* scope all progress overlays to the Claude window ([6dce87c](https://github.com/OpenScz/claude-desktop-switcher/commit/6dce87c9a8009f62fca3570f06f76dbc94225abe))

## [1.2.1](https://github.com/OpenScz/claude-desktop-switcher/compare/v1.2.0...v1.2.1) (2026-10-26)


### Bug Fixes

* detect current Windows Claude profile and login log paths ([#14](https://github.com/OpenScz/claude-desktop-switcher/issues/14)) ([543acf5](https://github.com/OpenScz/claude-desktop-switcher/commit/543acf506c696bce28eb0a08e8b093817648cb20))
* make account switching on Windows more reliable ([#17](https://github.com/OpenScz/claude-desktop-switcher/issues/17)) ([ed99588](https://github.com/OpenScz/claude-desktop-switcher/commit/ed99588dd69efb967d8cc639813ec6e2c1a78837))

## [1.2.0](https://github.com/TinoA/claude-desktop-switcher/compare/v1.1.0...v1.2.0) (2026-07-25)


### Features

* rename product to Claude Desktop Switcher ([7373970](https://github.com/TinoA/claude-desktop-switcher/commit/7373970f0ab4850acaae52003702bdecc1c8876d))

## [1.1.0](https://github.com/TinoA/claude-desktop-swap/compare/v1.0.3...v1.1.0) (2026-07-25)


### Features

* improve Windows tray and documentation ([5a563fe](https://github.com/TinoA/claude-desktop-swap/commit/5a563fe878213a82e1772a193472c6850597006e))
* prepare Windows Claude Swap release ([f980d2b](https://github.com/TinoA/claude-desktop-swap/commit/f980d2bc8b0f40ced354af4d1cd42f01a9c2bccc))

## [1.0.3](https://github.com/TinoA/claude-desktop-swap/compare/v1.0.2...v1.0.3) (2026-07-14)


### Bug Fixes

* save renewed Claude sessions automatically ([96f9783](https://github.com/TinoA/claude-desktop-swap/commit/96f97831579b24a2cbe8397197bf11771b798e98))

## [1.0.2](https://github.com/TinoA/claude-desktop-swap/compare/v1.0.1...v1.0.2) (2026-07-13)


### Bug Fixes

* check fork releases for updates ([c35362f](https://github.com/TinoA/claude-desktop-swap/commit/c35362fbe16d55264de4a7c698572487b04bb9ff))

## [1.0.1](https://github.com/TinoA/claude-desktop-swap/compare/v1.0.0...v1.0.1) (2026-07-13)


### Bug Fixes

* publish Windows installers ([b3a4665](https://github.com/TinoA/claude-desktop-swap/commit/b3a4665df62f270dccda42ced98f579a045723fa))

## 1.0.0 (2026-07-13)


### Features

* add interactive profile picker and account info display ([3813d97](https://github.com/TinoA/claude-desktop-swap/commit/3813d974f079f4a3a342b01b037b3879260915b9))
* add Windows Claude Swap tray and release automation ([1331cd8](https://github.com/TinoA/claude-desktop-swap/commit/1331cd83e73435b7cfb3fc59384574017dcdff80))
* **cli:** initial implementation of claude-swap ([bff1ae5](https://github.com/TinoA/claude-desktop-swap/commit/bff1ae57739f0f881e23fbc9f0a3498ae62a743f))
* make save automatic by stopping and relaunching Claude Desktop ([09ef9aa](https://github.com/TinoA/claude-desktop-swap/commit/09ef9aab99b2aa2486abe2f74eab3352fe7e2a0b))
* rename to claude-desktop-swap, add interactive 'add' command, auto-prime on save ([8d68249](https://github.com/TinoA/claude-desktop-swap/commit/8d68249940134bf7e975ec6e1d70e23f4e921d54))
* show live account email and plan in list and picker ([c928767](https://github.com/TinoA/claude-desktop-swap/commit/c92876701bf0defbcd0d1c2ccff11eed5c275efd))


### Bug Fixes

* **cd:** remove unsupported changelog path field ([c2cb3db](https://github.com/TinoA/claude-desktop-swap/commit/c2cb3db971363f75f54be1b781456038722277b8))
* checkpointed profile switching and cookie-only v2 profiles ([#3](https://github.com/TinoA/claude-desktop-swap/issues/3)) ([abda7b1](https://github.com/TinoA/claude-desktop-swap/commit/abda7b1e73e17877753890e24d46ce460bebca32))
* decrypt real sessionKey and read correct account API fields ([c52aaa7](https://github.com/TinoA/claude-desktop-swap/commit/c52aaa7cc152e5ab92be9d4ca253c54eec5d1cf4))
* preserve device trust across account swaps ([5cdd6c4](https://github.com/TinoA/claude-desktop-swap/commit/5cdd6c4cbe16e0537e3ac679c3c1f18985e9cd23))
* **profile:** clear Cookies-journal on restore ([3a29f63](https://github.com/TinoA/claude-desktop-swap/commit/3a29f6395895dd9a12ea0084443dd6f5754820d4)), closes [#1](https://github.com/TinoA/claude-desktop-swap/issues/1)
* **profile:** wipe IndexedDB and Session Storage on restore ([1238355](https://github.com/TinoA/claude-desktop-swap/commit/1238355292e8b95f7c3b9af206d88195b6b3fd10))

## [Unreleased]

### Fixed
- Preserve per-account Local Storage, IndexedDB, Session Storage, and Cloudflare security cookies during switches to avoid unnecessary device or human verification prompts.
- Roll back every replaced account-state directory when a restore fails, including paths that did not exist before the attempted switch.
- Refresh the active profile before export and refuse incomplete legacy-profile backups; imports identify accounts that still need a one-time refresh.

## [0.3.1] - 2026-06-19

### Added
- Local profile health states (`usable`, `expired`, `missing`, and `unknown`) based on non-secret SQLite evidence.
- Cookie-only v2 profiles with integrity metadata, secure permissions, atomic generation recovery, and compatible v1 reads.

### Changed
- Switching now fully stops Claude and its helpers, checkpoints the outgoing profile before restoration, replaces live Cookies atomically, and advances tracking only after commit.
- `save --force` now refuses to snapshot an open database; all saves require Claude to be stopped.
- Local Storage, IndexedDB, and Session Storage are cleared as volatile caches instead of stored in profiles; global and machine data remain untouched.

### Fixed
- Refreshed outgoing cookies are no longer discarded during repeated account switches.
- Interrupted profile writes and live-cookie replacement retain a recoverable prior generation.

## [0.3.0]

### Added
- `add <name>` command — snapshots your current session, gives Claude Desktop a clean slate to log into a new account, then snapshots that new session. Removes the manual logout/login dance for adding accounts.

### Changed
- Binary renamed from `claude-swap` to `claude-desktop-swap` to avoid collision with an unrelated tool
- `save` now primes the on-disk state right after capturing the profile (sets it as active and wipes stale per-account caches), so the next switch always lands cleanly without needing a manual `use` first

### Fixed
- Profile snapshots now include `IndexedDB`, `Session Storage`, and `ant-did` — stale per-account state from a previous profile was invalidating the restored session and forcing a re-login

### Added
- `save` command — snapshot the current Claude Desktop session as a named profile
- `use` command — switch to a saved profile (kills and restarts Claude Desktop)
- `list` command — list saved profiles with creation date and last used timestamp
- `delete` command — remove a saved profile
- `status` command — show the currently active profile
- Full session isolation: both cookies (SQLite) and local storage (LevelDB) are swapped on profile switch
- macOS platform implementation with graceful SIGTERM → SIGKILL shutdown sequence
- Platform abstraction layer for future Windows support
- CI pipeline (lint + tests on Ubuntu and macOS) via GitHub Actions
- CD pipeline (GoReleaser) triggered on SemVer tags — builds macOS binaries (amd64 + arm64)
- `CODEOWNERS` and branch protection documentation for maintainer-approval workflow
- `.gitmessage` commit template following Conventional Commits
