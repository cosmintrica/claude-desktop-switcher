# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0](https://github.com/cosmintrica/claude-desktop-switcher/compare/v1.2.2...v1.3.0) (2026-11-06)


### Features

* add interactive profile picker and account info display ([3813d97](https://github.com/cosmintrica/claude-desktop-switcher/commit/3813d974f079f4a3a342b01b037b3879260915b9))
* add Windows Claude Swap tray and release automation ([1331cd8](https://github.com/cosmintrica/claude-desktop-switcher/commit/1331cd83e73435b7cfb3fc59384574017dcdff80))
* **cli:** initial implementation of claude-swap ([bff1ae5](https://github.com/cosmintrica/claude-desktop-switcher/commit/bff1ae57739f0f881e23fbc9f0a3498ae62a743f))
* improve Windows tray and documentation ([5a563fe](https://github.com/cosmintrica/claude-desktop-switcher/commit/5a563fe878213a82e1772a193472c6850597006e))
* make save automatic by stopping and relaunching Claude Desktop ([09ef9aa](https://github.com/cosmintrica/claude-desktop-switcher/commit/09ef9aab99b2aa2486abe2f74eab3352fe7e2a0b))
* prepare Windows Claude Swap release ([f980d2b](https://github.com/cosmintrica/claude-desktop-switcher/commit/f980d2bc8b0f40ced354af4d1cd42f01a9c2bccc))
* rename product to Claude Desktop Switcher ([7373970](https://github.com/cosmintrica/claude-desktop-switcher/commit/7373970f0ab4850acaae52003702bdecc1c8876d))
* rename to claude-desktop-swap, add interactive 'add' command, auto-prime on save ([8d68249](https://github.com/cosmintrica/claude-desktop-switcher/commit/8d68249940134bf7e975ec6e1d70e23f4e921d54))
* **sessions:** read each account's recent Claude Code sessions ([2860bbd](https://github.com/cosmintrica/claude-desktop-switcher/commit/2860bbd1692ff7a51513bc38f2f50a53811e1867))
* show live account email and plan in list and picker ([c928767](https://github.com/cosmintrica/claude-desktop-switcher/commit/c92876701bf0defbcd0d1c2ccff11eed5c275efd))
* **tray:** an aquarium above the clock, one fish per Claude window ([1956c87](https://github.com/cosmintrica/claude-desktop-switcher/commit/1956c8766fa3eb639e01f99e5961a5902e7a9629))
* **tray:** ask before closing Claude while sessions are working ([e4ad94a](https://github.com/cosmintrica/claude-desktop-switcher/commit/e4ad94a7a24f54a2bfa44356faae03b33adb104a))
* **tray:** close one Claude window at a time ([0846f2e](https://github.com/cosmintrica/claude-desktop-switcher/commit/0846f2e9967f31c1c9396f38ae9cf81715874ffa))
* **tray:** continue a recent session on its own account ([7047052](https://github.com/cosmintrica/claude-desktop-switcher/commit/70470525d0e7ef3020a0e9e0a096647c9846a946))
* **tray:** group sessions the way Claude's sidebar does ([2db994b](https://github.com/cosmintrica/claude-desktop-switcher/commit/2db994beb2869079734710bb056d95450652bf72))
* **tray:** show progress as a terminal card with no background ([7072862](https://github.com/cosmintrica/claude-desktop-switcher/commit/70728624280ee473dfbd026cd78ae756c24996fa))
* **tray:** start at login ([6ce92b7](https://github.com/cosmintrica/claude-desktop-switcher/commit/6ce92b7b0dfd05521de90fa5e9e31c077f56354f))
* **windows:** a fishbowl icon for the tray and the program ([070b472](https://github.com/cosmintrica/claude-desktop-switcher/commit/070b4724f0a69bfd95b5984446d3176cc3791777))
* **windows:** open an account in a separate Claude window ([b5dec23](https://github.com/cosmintrica/claude-desktop-switcher/commit/b5dec23209b406e1b01fd69056cdffbb88ed8e77))


### Bug Fixes

* **aquarium:** stay in front, and hide only for a full-screen window in front ([2d677ee](https://github.com/cosmintrica/claude-desktop-switcher/commit/2d677eea71a7cdae4fbc8f027504cf414f178519))
* **cd:** remove unsupported changelog path field ([c2cb3db](https://github.com/cosmintrica/claude-desktop-switcher/commit/c2cb3db971363f75f54be1b781456038722277b8))
* check fork releases for updates ([c35362f](https://github.com/cosmintrica/claude-desktop-switcher/commit/c35362fbe16d55264de4a7c698572487b04bb9ff))
* check fork releases for updates ([d050c4b](https://github.com/cosmintrica/claude-desktop-switcher/commit/d050c4b5039980aa054c313c80e984b84afd4398))
* checkpointed profile switching and cookie-only v2 profiles ([#3](https://github.com/cosmintrica/claude-desktop-switcher/issues/3)) ([abda7b1](https://github.com/cosmintrica/claude-desktop-switcher/commit/abda7b1e73e17877753890e24d46ce460bebca32))
* constrain profile metadata and preserve failed rollback state ([35adcfc](https://github.com/cosmintrica/claude-desktop-switcher/commit/35adcfca8a1d2956ab1f0dac92b50ad04657b236))
* decrypt real sessionKey and read correct account API fields ([c52aaa7](https://github.com/cosmintrica/claude-desktop-switcher/commit/c52aaa7cc152e5ab92be9d4ca253c54eec5d1cf4))
* detect current Windows Claude profile and login log paths ([#14](https://github.com/cosmintrica/claude-desktop-switcher/issues/14)) ([543acf5](https://github.com/cosmintrica/claude-desktop-switcher/commit/543acf506c696bce28eb0a08e8b093817648cb20))
* make account switching on Windows more reliable ([#17](https://github.com/cosmintrica/claude-desktop-switcher/issues/17)) ([ed99588](https://github.com/cosmintrica/claude-desktop-switcher/commit/ed99588dd69efb967d8cc639813ec6e2c1a78837))
* preserve device trust across account swaps ([5cdd6c4](https://github.com/cosmintrica/claude-desktop-switcher/commit/5cdd6c4cbe16e0537e3ac679c3c1f18985e9cd23))
* **profile:** clear Cookies-journal on restore ([3a29f63](https://github.com/cosmintrica/claude-desktop-switcher/commit/3a29f6395895dd9a12ea0084443dd6f5754820d4)), closes [#1](https://github.com/cosmintrica/claude-desktop-switcher/issues/1)
* **profile:** read the Store install's log in the Local folder ([60fb6ed](https://github.com/cosmintrica/claude-desktop-switcher/commit/60fb6edd0036b7dcb2b1a2bc6d8fb383dd7b50b3))
* **profile:** retry renames that Windows refuses while a file is briefly open ([3d855aa](https://github.com/cosmintrica/claude-desktop-switcher/commit/3d855aa96ff70ba2518e742237f8db0a328dac51))
* **profile:** wipe IndexedDB and Session Storage on restore ([1238355](https://github.com/cosmintrica/claude-desktop-switcher/commit/1238355292e8b95f7c3b9af206d88195b6b3fd10))
* publish Windows installers ([b3a4665](https://github.com/cosmintrica/claude-desktop-switcher/commit/b3a4665df62f270dccda42ced98f579a045723fa))
* save renewed Claude sessions automatically ([96f9783](https://github.com/cosmintrica/claude-desktop-switcher/commit/96f97831579b24a2cbe8397197bf11771b798e98))
* save renewed Claude sessions automatically ([5fda5b3](https://github.com/cosmintrica/claude-desktop-switcher/commit/5fda5b3b3e247f6508b79163305dac7be2e5104a))
* scope all progress overlays to the Claude window ([6dce87c](https://github.com/cosmintrica/claude-desktop-switcher/commit/6dce87c9a8009f62fca3570f06f76dbc94225abe))
* **tray:** ask before a switch or a new account closes Claude ([765a4a8](https://github.com/cosmintrica/claude-desktop-switcher/commit/765a4a86eb82323742d6f7dbd9e3a9a948e66ef6))
* **tray:** fish from the start, the plus only under the pointer ([e640bba](https://github.com/cosmintrica/claude-desktop-switcher/commit/e640bba01d25256047e2cf434279d9051b358d3f))
* **tray:** pause background profile reads while profiles change ([1c83bfc](https://github.com/cosmintrica/claude-desktop-switcher/commit/1c83bfc2112e77048992b74ba79830ede63c2cf9))
* **tray:** show switching progress in a small card instead of dimming the screen ([6f01a45](https://github.com/cosmintrica/claude-desktop-switcher/commit/6f01a45f391f7015a89945e424ede25172653c7a))
* update Go to 1.27.3 for GO-2026-6809 ([c7f6022](https://github.com/cosmintrica/claude-desktop-switcher/commit/c7f6022a99ea4b4e1124aaa0f8e844ef0fae4ae5))
* update Go toolchain to address standard library advisories ([f8d9a17](https://github.com/cosmintrica/claude-desktop-switcher/commit/f8d9a17aa86b3b0f640cc57feb1a4d9a982b23a5))
* **windows:** ask Claude to quit cleanly before stopping it ([3206ab4](https://github.com/cosmintrica/claude-desktop-switcher/commit/3206ab4ebe24af1350035e2177f8e40c95a8adfb))
* **windows:** copy a separate window's sessions instead of linking them ([5eeddb4](https://github.com/cosmintrica/claude-desktop-switcher/commit/5eeddb4e2136395a83535fa415683e6b30532629))
* **windows:** follow Claude 3's data folder out of the Store package ([2d29171](https://github.com/cosmintrica/claude-desktop-switcher/commit/2d29171f9bb991a631bebce3d27d6cdf674105b2))
* **windows:** leave Claude windows with their own data folder alone ([ea37882](https://github.com/cosmintrica/claude-desktop-switcher/commit/ea378825d75b7256e906ce20ca7d61caadfe9ee7))
* **windows:** open the tray when started from Explorer ([d938fea](https://github.com/cosmintrica/claude-desktop-switcher/commit/d938feac3e08fdd82d158c37bc451752571041ab))
* **windows:** show an account's sessions in its separate window ([27a87d0](https://github.com/cosmintrica/claude-desktop-switcher/commit/27a87d0e5c6cc81d90cf2cb4097f14719ec10db7))


### Reverts

* **windows:** stop Claude with the original close sequence ([b9a8dcd](https://github.com/cosmintrica/claude-desktop-switcher/commit/b9a8dcdd72006b1fdeffff5880f82f2b112645d0))

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
