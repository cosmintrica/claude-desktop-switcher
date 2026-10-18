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

## Running a local build

1. Build with `.\scripts\project.ps1 -Task Build -Arch amd64 -Version <version>`,
   or with `go build` and the same `-ldflags` (`-H=windowsgui` and the version).
2. Keep each build next to the previous ones under its own file name, and point
   a desktop shortcut at it with the argument `tray`.
3. Only one tray runs at a time. To change builds, choose **Exit** in the tray
   menu, then start the other one. The menu's last line shows the running version.
