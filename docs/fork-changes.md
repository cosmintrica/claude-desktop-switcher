# Changes in this fork

This fork of [OpenScz/claude-desktop-switcher](https://github.com/OpenScz/claude-desktop-switcher)
collects fixes found while using Claude Desktop Switcher every day on
Windows 11 with Claude Desktop from the Microsoft Store (MSIX). It has no
installer releases; the builds were made locally from source.

| Branch | Contents |
|--------|----------|
| `main` | Everything below, plus this page |
| `fix/stable-switching` | The code and documentation changes for the original project |
| `codex/harden-profile-import-and-rollback` | The first round, proposed in [OpenScz/claude-desktop-switcher#16](https://github.com/OpenScz/claude-desktop-switcher/pull/16) |
| `feat/account-sessions` | The third round, the **Continue a session** menu |

## First round: profile import and rollback (2026-10-18)

Reported in [issue #1](https://github.com/cosmintrica/claude-desktop-switcher/issues/1).

- An imported backup could carry a profile name that pointed outside its own
  folder, and metadata backfill would then rewrite an existing `meta.json`
  elsewhere. Metadata reads and writes are now bound to the validated profile
  folder, and an explicit name mismatch rejects the import before anything is
  committed.
- When a restore or account-file replacement failed and could not roll back
  (for example because Windows kept a file locked), the rollback folder holding
  the originals was deleted anyway. The originals are now kept, the error says
  where they are, and Claude is not reopened on a half-restored state.
- Go was updated from 1.25.12 to 1.27.2 and golangci-lint to 2.15.0.
  govulncheck reported 13 reachable standard-library advisories with the old
  toolchain and none with the new one.

## Second round: switching stability (2026-10-18)

Three symptoms were reported: the screen flickered while switching and the
whole thing did not feel stable; connecting a connector in Claude only made
Claude flicker; and Claude could not be pinned to the taskbar, while clicking
its taskbar icon minimized and reopened it.

### Fixed

| Problem | Cause | Change | Guarding test |
|---------|-------|--------|---------------|
| A switch failed with `rename ... Access is denied` after Claude had already closed; Claude reopened on the old account and the outgoing session was not saved | Windows refuses to rename a folder while a file inside it is open, and the tray refreshed its menu (opening every profile's Cookies database) at the same moment a checkpoint renamed the profile folder | Profile-store renames retry for up to 5 s on access, sharing and lock violations; background menu refreshes and email discovery pause during a switch, an account addition or any operation lock; a closed session is saved before the menu refresh | `TestCheckpointSucceedsWhileTheSavedProfileIsBrieflyRead` reproduces the exact error without the retry; `TestRenamePath*`, `TestOperationLockHeldReportsOnlyALiveOwner`, `TestBackgroundReadsPauseWhileProfilesChange` |
| The screen flickered during every switch | A translucent topmost window covered every monitor and the taskbar, was repainted 8 times a second, took focus from Claude and stayed up while Claude reopened | A small centered card that never takes focus or clicks, drawn in memory and copied in one step | Checked visually |
| The tray would abort after 2000 window checks | Each `EnumWindows` call created a new `windows.NewCallback`; Go never frees them and throws `too many callback functions` at 2000. The login-window check runs four times a second while adding an account | One shared callback serves every enumeration | `TestEnumTopLevelWindowsCanRunMoreTimesThanGoAllowsCallbacks` |
| Reopening Claude took a maximized window out of its maximized size | `SW_RESTORE` was used for every focus | `SW_RESTORE` only for a minimized window | `TestShowCommandKeepsAMaximizedWindowMaximized` |
| A double-click or desktop shortcut on the executable did nothing | With Explorer as the parent, cobra's mousetrap waited 5 s and exited with status 1 before the tray started (the installer's native launcher hid this) | `cobra.MousetrapHelpText` is cleared | `TestExplorerStartIsNotStoppedByTheCommandLineGuard` |

### Investigated, not caused by the switcher

- **Connectors.** Claude opens the connector login in the default browser.
  After the login, the browser hands back through a `claude://` link: Windows
  starts a second Claude process for a moment, which passes the link to the
  running Claude and exits, and Claude comes to the front. That is the flicker
  seen; the connector was added successfully.
- **Taskbar.** The switcher starts Claude through its package identity
  (`shell:AppsFolder\Claude_pzs8sxrjxfjjc!Claude`), and Claude stayed pinned and
  grouped with its window after every switch. A double-click on a taskbar button
  minimizes and immediately restores the window; one click is enough. During a
  switch, the old full-screen overlay also covered the taskbar.

### Known limitation

Claude Desktop hides in the system tray when its window closes (its default),
so the switcher's close sequence ends with `taskkill` and Claude's quit cleanup
does not run. Claude Code sessions running inside Claude Desktop are stopped by
a switch and can be continued after it reopens. Sending `WM_ENDSESSION` was
tried: Claude calls `app.quit()` for it, but Electron ends the process about
40 ms after that message (reproduced with a disposable Electron 44 window), so
the cleanup starts and is cut off. That attempt was reverted.

### Verification

- gofmt, `go mod tidy -diff`, `go vet ./...` and all package tests pass;
  golangci-lint 2.15.0 reports 0 issues; govulncheck 1.6.0 finds 0 reachable
  vulnerabilities.
- A verified local backup (every copied file re-hashed with SHA256) was made
  before any live test.
- Two live rounds of switching away and back with Claude Desktop 2.42304: both
  outgoing profiles were saved, no errors were logged, and Claude reopened with
  its package identity and stayed pinned.

### Related observation, not changed here

Claude Desktop keeps Claude Code session records per account. A session
record copied to another account can report that its branch is checked out in
another folder, because each worktree folder is leased to one record id.

## Third round: continue a session on its account (2026-10-18)

Claude Desktop keeps each account's Claude Code sessions apart
(`claude-code-sessions\<account>\<organization>\local_*.json`), and a session
opens only on the account that owns its record. Moving between accounts meant
remembering where each piece of work lived.

The tray has a new **Continue a session** menu. Under each saved account it
lists that account's 10 most recent sessions, grouped by project folder, with
the time of their last activity. Clicking one opens it on its account:

- On the account already open in Claude, Claude comes to the front with the
  session.
- On another account, the switcher asks first (Claude closes and reopens, and
  work running on the current account stops), switches, waits for Claude's
  window and opens `claude://code/continue?session=<id>`. Claude 2.42304
  accepts only ids made of `local_` and up to 64 letters, digits and dashes,
  and opens the session only if the account in use owns it and it is not
  archived.

Which account lists a session:

- A record belongs to a saved account when the SHA-256 of its account folder
  name equals the profile's `account_uuid_hash`.
- Records copied between accounts share one `cliSessionId`. Each conversation
  is listed once, under the account whose record was active most recently.
  Exact copies tie on activity time; the account open in Claude wins the tie,
  then the menu order.
- Claude reserves each worktree folder for one record (`leasedBy` in
  `git-worktrees.json`) and opens any other record of that folder in a fresh
  worktree on the record's old branch, so only the record holding the folder
  is listed.
- Archived sessions, scheduled-task runs, sessions without a transcript and
  sessions whose folder no longer exists are left out.

The menu refreshes with the rest of the tray every 10 seconds, and only records
whose size or modification time changed are read again. Records are opened with
`FILE_SHARE_DELETE`, so Claude can delete one while the switcher reads it. A
replacement that meets an open record fails for a moment; Claude's own writer
retries it after 50 and 100 ms and then writes in place.

Guarding tests: `internal/sessions` (`TestRecent*`,
`TestReadingARecordDoesNotStopClaudeFromDeletingIt`) and
`cmd/tray_sessions_test.go`.

### A warning before running work stops

Claude Code writes one file per running process in `~/.claude/sessions` and
marks it `busy` while a turn runs. Before switching accounts, continuing a
session on another account or closing Claude Desktop, the switcher lists the
busy sessions of the main window and asks first; **No** cancels. A file whose
process has ended, or whose process id now belongs to a process started at
another time, is ignored. Guarding tests: `TestWorking*`,
`TestProcessRunningComparesTheRecordedStartTime`,
`TestStopWorkNoticeListsAtMostFiveSessions`.

### Separate Claude windows are left alone

Claude Desktop can run a second window beside itself when started with its own
`--user-data-dir`. The switcher counted every `Claude.exe` of the package as
Claude, so a switch or **Close Claude Desktop** also stopped such windows. Each
process is now judged by the root of its Claude process tree, and a tree whose
root was started with another data folder is ignored. The main window's own
helper processes also carry `--user-data-dir` (pointing at its own folder), so
a helper's flag decides nothing, and a root whose command line cannot be read
still counts as the main window. Guarding tests: `TestMainInstance*`,
`TestSeparateDataFolderLooksOnlyAtArguments`,
`TestProcessCommandLineReadsARunningProcess`.

## Running a local build

1. Build with `.\scripts\project.ps1 -Task Build -Arch amd64 -Version <version>`,
   or with `go build` and the same `-ldflags` (`-H=windowsgui` and the version).
2. Keep each build next to the previous ones under its own file name, and point
   a desktop shortcut at it with the argument `tray`.
3. Only one tray runs at a time. To change builds, choose **Exit** in the tray
   menu, then start the other one. The menu's last line shows the running version.
