# Changes in this fork

This fork of [OpenScz/claude-desktop-switcher](https://github.com/OpenScz/claude-desktop-switcher),
made by [@cosmintrica](https://github.com/cosmintrica),
collects fixes found while using Claude Desktop Switcher every day on
Windows 11 with Claude Desktop from the Microsoft Store (MSIX). It has no
installer releases; the builds were made locally from source.

| Branch | Contents |
|--------|----------|
| `main` | Everything below, plus this page |
| `fix/stable-switching` | The code and documentation changes for the original project |
| `codex/harden-profile-import-and-rollback` | The first round, proposed in [OpenScz/claude-desktop-switcher#16](https://github.com/OpenScz/claude-desktop-switcher/pull/16) |
| `feat/account-sessions` | The third round: sessions per account, separate windows, the aquarium |

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

Each account's row in the tray has a **Continue a session** submenu. It
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

### Groups as in Claude's sidebar

The menu groups each account's sessions the way Claude's sidebar does. Groups
the user made come first, in sidebar order: `claude_desktop_config.json`,
`preferences.epitaxyPrefs`, `dframe-group-scopes` and `dframe-code-sections`,
keyed by account and organization. Nothing else is read from that file. The
other sessions are grouped by repository and named after the origin Claude
pinned for it, its origin URL, or its folder (`git-worktrees.json`,
`originPins` and `originUrls`), alphabetically. A side session sits under the
session that started it, except in groups the user made. Guarding tests:
`TestRecentGroupsSessionsLikeClaudesSidebar`,
`TestSessionMenuEntriesFollowTheSidebar`.

### A terminal progress card

The progress card is now a short terminal session that floats over the screen
with no background. The command types itself (`claude switch <account>`,
`claude continue <account>`, `claude open`, `claude quit`, `claude backup`,
`claude window <account>`, `claude account add`). A bar fills with the real
steps of the operation: closing Claude, saving the outgoing account, restoring
the incoming one, opening Claude, opening the session. The current step is
centered below the bar, and all three lines share one width. The window is
layered with per-pixel alpha (`UpdateLayeredWindow`), and a dark outline keeps
the text readable over light and dark windows. The font is Cascadia Mono where
Windows has it, else Consolas. Guarding tests: `TestTerminalCard*`,
`TestPartialCell`, `TestCardSteps*`.

### Open an account in a separate window

**Open in a separate window** starts a second Claude Desktop beside the main
one, signed in to the chosen account. It runs the Store package's `Claude.exe`,
found from the installed package each time (`GetPackagesByPackageFamily` and
`AppxManifest.xml`), with `--user-data-dir=%LOCALAPPDATA%\Claude-Accounts\<profile>`.
That folder sits outside the package's redirected Roaming folder, and Claude
keeps its package identity. The first time, the window opens signed out and
the user signs in with **Sign in with code**. Sign-ins and cookies are never
copied between folders. The account's row says when it is open in a separate
window, and when another account is signed in to its window.
Limits: `claude://` links (connector sign-ins, "Open in Claude") always reach
the main window, there is no Cowork VM in a separate window, each window keeps
its own copy of Claude Code, and each uses about 1.5-2 GB of memory. Guarding
tests: `TestAccountRowTitleSaysWhereTheAccountIsOpen`, `TestSeparateWindowFolderSitsOutsideTheStorePackage`,
`TestManifestExecutableStaysInsideThePackage`,
`TestClaudeExecutableFromTheInstalledPackage`.

### Claude 3 moved its data folder

Claude Desktop 3 (seen with 3.1623 on 2026-10-21; the move itself happened on
2026-10-20) moves its data out of the Store package's redirected Roaming folder
into `%LOCALAPPDATA%\Claude-Data\<package family>` (here
`claude_pzs8sxrjxfjjc`), and leaves a `.user-data-left.<family>` marker in the
old folder; `.user-data-back.<family>` marks a move back. The old folder kept
only that marker, so a switch would have restored the account into a folder
Claude no longer reads and reported success while Claude stayed on the same
account (nothing was lost: with no live cookies there, the outgoing account
was not saved over). The switcher now follows the marker, or the folder that
holds Claude's data, the same way Claude does. The sign-in keys in
`config.json` are unchanged. Guarding test:
`TestMsixDataFolderFollowsClaudesMoveOutOfThePackage`.

### A separate window shows the account's sessions

Claude keeps the Code tab's session records in its data folder
(`claude-code-sessions\<account>\<org>\local_*.json`), so a separate window
opened with an empty list even for an account with many sessions (seen on
2026-10-24 with Claude 3.6556).

The first fix (2026-10-24, 27a87d0) made the window's `claude-code-sessions` a
directory junction to the main window's folder. Claude reads through such a
link but refuses to write through it: its private-file writer opens the
temporary file with `O_EXCL` and checks the path, and every save failed with
`EEXIST`. From 2026-10-25 to 2026-11-01 nothing the separate windows did was
saved: 12 new sessions of one account and 2 of another lived only in memory
and left the list when Claude restarted after its update to 4.4212. Their
conversations were safe in `~/.claude/projects`, and their records were
rebuilt from the windows' logs ("Mapping internal session ... to CLI session
...") and the conversations.

Now each separate window keeps a records folder of its own, and the switcher
copies one account's files between it and the main folder: the records, the
deletion markers, `scheduled-tasks.json` and `archived-sessions.idx`. The copy
changed last wins and keeps its source's time, a deletion marker removes the
record on the other side, and nothing is copied into a folder whose Claude
holds that account. Before a window that is not running starts, a link left by
the first fix is removed (the main folder is untouched), what the window saved
before goes to the main folder, then the account's sessions come back from it.
While a window runs, what it saves reaches the main folder on every tray
refresh (10 s) and right before a switch, so **Continue a session** lists it
and the main window has it after the account moves there.

Two windows must never hold the same account. The row of an account signed in
to a separate window offers **Show its window** instead of **Switch to this
account**, and a switch that reaches such an account anyway offers to bring
that window forward instead of switching the main window to it. **Continue a
session** for such an account opens the session in its separate window, without
closing the main one: the link is passed to `Claude.exe` started with that
window's `--user-data-dir`, which hands it to the running window, since links
opened through Windows always reach the main window. The menu marks that
account "(separate window)".

Claude reserves each worktree folder for one session in `git-worktrees.json`,
one file per data folder, and gives a session whose reservation is missing a
fresh worktree when its folder has moved to another branch. Before a separate
window starts, it therefore receives the main window's reservations for its
account's sessions (only those not yet in its own file, never the spare pool
or folders still being created). Reservations a window makes for new sessions
stay in its own file; when that account opens in the main window again, Claude
leases such a folder back in place if its branch is unchanged. The question
about busy sessions before a switch counts only the sessions of the account
signed in to the main window, because separate windows keep running. Sidebar
groups and starred sessions stay per window. Guarding tests:
`TestMirrorAccount*`, `TestUnlinkRecordsReplacesALinkWithoutTouchingTheMainFolder`,
`TestCopyWorktreeLeases*`, `TestWorkingLeavesOutSessionsOfAccountsInSeparateWindows`,
`TestSeparateWindowInUsePrompt`, `TestAccountRowTitleSaysWhereTheAccountIsOpen`.

### A switch always asks first

A switch and **Add account...** closed Claude Desktop at once
when no Claude Code session was working. On 2026-10-24 a build from before the
Claude 3 fix looked for working sessions in the old, empty data folder, found
none, and closed Claude without asking while sessions were working. Both
actions now always ask first while Claude Desktop is open, and the question
lists the sessions that are working or says that none is. **Close Claude
Desktop** still asks only while a session is working. Guarding tests:
`TestSwitchPromptAsksEvenWhenNoSessionIsWorking`,
`TestAddAccountPromptAsksBeforeClaudeCloses`.

### The aquarium

A small aquarium sits above the clock, drawn as pixel art. Each open Claude
window is a fish in its account's colour, the crowned fish is the main window,
and bubbles rise while a Claude Code session works in that window. The fish
are voxel models drawn from eight angles (`internal/aquarium`), so they turn
through them as they swim. Hovering a fish shows its account; a click shows
or closes its window, asking first when a session is working. The plus on the
rim lists the accounts without a window, to open one in a separate window or
switch the main window to it. Closing a window makes its fish jump out, and
the progress shows in the aquarium instead of the terminal card.

The window never takes focus and its transparent pixels let clicks through.
It hides while the window in front fills its screen without a title bar: a
game, a video or a presentation. Maximized windows keep their title bar, and
the desktop, Alt+Tab and Task View belong to Explorer, so none of them hide it.
Windows' own notification state is no longer asked: it does not say which
screen is full or whether that window is in front. Every two seconds the
aquarium also checks the windows above its bowl and comes back in front when
Windows has dropped its always-on-top flag, when the desktop has come over it,
or when another program's always-on-top window covers it. Menus, tooltips,
click-through overlays and Explorer's pop-ups stay in front. Each hide, return
and comeback writes one line to the tray log, with the program that caused it.
**Show aquarium** under **Settings and backup** turns it off and on; the choice is kept in
`~/.claude-swap/aquarium-off`.

### Account limits

Each account's submenu starts with one row of how much of its claude.ai limits
is used, `5h 40% · week 12%`, so an account with room can be picked before a
switch. A full window adds the time it starts over, `week 100% (Tue 00:00)`,
and a click on that row reads the limits again. Two earlier versions put the
limits on the account's own row, as bars made of characters and then as
numbers in a column; both made the menu too wide. The tray log keeps the
words, for example `5h 92% used, resets 15:20`. In the aquarium, a fish's panel and the question
before a switch show the two limits as bars, and the accounts without a window
show their percents in colour: teal below 70 %, amber below 90 %, red from 90 %.

The limits are read when the tray starts, after each switch for the account it
left (its usage moved while it was in use), and on **Refresh
limits**. For each saved account the switcher opens the claude.ai sign-in kept
in its Cookies copy with Claude's own key from `Local State` (unwrapped with
DPAPI for this Windows user, then AES-GCM) and asks claude.ai's usage endpoint
for the organization the account last used. This is the account-info code in
`internal/account`, the one place that opens a cookie. The sign-in stays in
memory for the request; only the percents and reset times are kept, in memory
too. Each read writes one line per account to the tray log. Guarding tests:
`TestWindowsDecryptorOpensTheSavedSignIn`, `TestWindowsDecryptorRefusesAnotherKey`,
`TestUsageReadsTheFiveHourAndWeeklyWindows`, `TestLimitStartsOverAtItsReset`,
`TestFetchUsage*`, `TestLimitSummaryFitsOnOneRow`. Checked live on 2026-11-07: all three saved accounts
answered with their 5-hour and weekly limits and reset times.

The accounts are asked one at a time. On 2026-11-08 the tray asked for all
five at once and claude.ai's Cloudflare guard answered some of them with
`HTTP 403`, `cf-mitigated: challenge` and its "Just a moment..." page, a check
that only a browser passes: two or three accounts showed no limits after each
start, though their sign-ins were fine. Five requests sent together reproduced
it; the same requests one after another all answered. Such an answer is now
reported as a browser check, and the accounts that got one are asked once more
after 5 seconds. Guarding tests: `TestFetchUsageRecognizesTheBrowserCheck`,
`TestFetchEachAsksOneAtATimeAndOnceMoreAfterABrowserCheck`; checked live with
the five saved accounts, which all answered.

### Start at login

**Start at login** under **Settings and backup** starts the switcher when you sign in to
Windows. It writes a `Claude Desktop Switcher` value with this copy's path and
`tray` under `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`. The tick
follows what Windows will do: the installer's Startup shortcut counts as on,
and turning it off removes that shortcut too; a switch-off in Settings > Apps >
Startup shows as off, and turning it on in the menu clears that switch-off.
While it is on, each start points the value at the copy that runs, so a new
build under another name or in another folder is the one Windows starts.
Guarding tests: `TestStartAtLogin*`.

### A fishbowl icon

The tray icon and the program's own icon are a pixel-art fishbowl with an
orange fish, in the aquarium's colours. It is drawn at 16 pixels and scaled by
whole pixels, so it stays sharp in the tray: the tray library loads the
32-pixel frame and Windows halves it there. Before, the program had no icon of
its own, so a shortcut to it showed Windows' blank program icon.
`rsrc_windows_amd64.syso` and `rsrc_windows_arm64.syso` carry the icon into
the program; `go run scripts/icon-resource.go` writes them again from
`cmd/assets/windows-claude-swap-icon-v2.ico`.

### Close one Claude window at a time

Claude's tray icons look the same for every window, so a white icon does not
tell which account it belongs to. Each open account's row has **Close its
window**, and while a separate window is open the main row reads **Close the
main window**. A separate window closes the way the main one does: its window
is asked to close, then its own process tree is ended, and the other windows
stay open. Claude Code sessions working in that window are listed first, as for
the main window.

systray puts a hidden row back whenever its title or state changes, and adding
or removing a row of the main menu closes that menu if it is open, so the
refresh every 10 seconds once closed an open tray menu. A row that stays hidden
is only hidden again, and rows of the main menu are added or removed only when
an account comes or goes.

### A simpler tray menu

The menu had grown to 14 rows, and the same accounts were listed in four
submenus (**Accounts**, **Continue a session**, **Open in a separate window**,
**Delete account...**). Now each saved account is one row at the top, ticked
while it is open, with where it is open, `(main window)` or
`(separate window)`. Its submenu starts with how much of its limits is used
and holds what can be done with it: **Switch to this account** and **Open in a separate window** while it
is closed, **Show its window** and **Close its window** while it is open,
**Continue a session**, and **Remove from Switcher...**, which removes the saved
sign-in and leaves the Claude account alone. Then come **Refresh limits**,
**Add account...**, **Open Claude Desktop** or **Close Claude Desktop**,
**Settings and backup** (backups, the aquarium, start at login, the logs and
the version) and **Exit**. Both mouse buttons open the same menu.

systray keeps the rows of a menu in the order they were made, so the account
rows are made first when the tray starts, one per saved account and three
spare ones that stay hidden until an account is added. Checked on 2026-11-08
by building the same menu with the same Win32 calls on a hidden desktop and
saving what Windows drew there (`PrintWindow`): the ticks show on rows that
have a submenu, and the top menu stays as narrow as the longest account name.
Guarding tests: `TestAccountRowTitleSaysWhereTheAccountIsOpen`,
`TestAccountActionsFollowWhereTheAccountIsOpen`, `TestClaudeMenuState`.

### The session bridge in the tray

claude-bridge is a separate local MCP server through which the Claude Code
sessions of every account on one computer talk. It keeps one SQLite file,
`~/.claude-bridge/bridge.db`, whose schema only ever grows. When that file
exists, the tray shows a **Session bridge** row, for example
`Session bridge: 5 sessions, 2 working`, and the icon's tooltip says the same.
Its submenu lists each active session with its account and `working` or
`idle`, then what waits for them: messages not read yet, files or tasks
claimed, and handoffs nobody has taken. The rows are only read. A session is
active while it checked in within the last 2 minutes, as the bridge counts it,
and it is working while Claude Code's own record of its process
(`~/.claude/sessions/<pid>.json`) says `busy`. The tray opens the database
read-only every 10 seconds and never writes to it; without the bridge, the row
and the tooltip stay as before. Guarding tests:
`TestReadShowsTheActiveSessionsAndWhatWaitsForThem`,
`TestReadSaysWhenTheBridgeIsNotInstalled`,
`TestBusyReadsTheProcessRecordOfThatSession`,
`TestBridgeSummaryCountsTheWorkingSessions`,
`TestBridgeEntriesListTheSessionsThenWhatWaits`,
`TestBridgeEntriesFitTheMenuSlots`. Checked on 2026-11-09 against the
installed bridge (five active sessions, two working) and in the real menu
drawn on a hidden desktop.

### Verification of the third round

- gofmt, `go mod tidy -diff`, `go vet ./...` and all package tests pass;
  golangci-lint 2.15.0 reports 0 issues; govulncheck 1.6.0 finds no reachable
  vulnerability.
- Live with Claude Desktop 2.42304 on 2026-10-19, with a separate test window
  open (started with its own `--user-data-dir`): **Close Claude Desktop** listed
  the busy sessions and asked first; the separate window stayed open through
  that close and through a switch; **Continue a session** switched to the other
  account and opened the chosen session, then switched back and opened the
  first one.
- Live with Claude Desktop 4.4212 on 2026-11-01: the account's separate window
  started with a records folder of its own, the old junction gone. A new session
  survived closing and reopening the window, a deleted session stayed deleted,
  and both showed up in the main folder within one refresh.

## Based on upstream v1.2.2 (2026-11-01)

The original project released v1.2.1, with this fork's switching fixes from
[#17](https://github.com/OpenScz/claude-desktop-switcher/pull/17), and v1.2.2
on 2026-10-26. `feat/account-sessions` merges v1.2.2. Two places were resolved
by hand:

- Claude's data folder: the Store install keeps following Claude 3's move to
  `Claude-Data` and, like upstream #14, uses an existing profile in the real
  Roaming folder when the package has none.
- The progress card: the fork keeps its terminal card. Upstream's card drawn
  over the Claude window (6dce87c) and the `FocusDesktopWindow` it needed are
  left out.

Fork builds name their version `v<upstream version>-fork-<commit>`. The tray's
update check reads the original project's releases and ignores the `-fork`
part, so **New version available** appears only when upstream releases
something newer than the version merged here. That release lacks the fork's
changes: merge it here instead of installing it.

Claude from the Store writes the main window's log to
`%LOCALAPPDATA%\Claude\Logs\main.log`, like the classic install. Upstream #14
reads that file only for the classic install; the fork reads it for the Store
install too, with its data in the package or in `Claude-Data`, so the checks
after a switch or a new sign-in see Claude's "logged in" and "logged out"
lines again. A separate window keeps reading the log in its own data folder.

## Running a local build

1. Build with `.\scripts\project.ps1 -Task Build -Arch amd64 -Version <version>`,
   or with `go build` and the same `-ldflags` (`-H=windowsgui` and the version).
2. Keep each build next to the previous ones under its own file name, and point
   a desktop shortcut at it with the argument `tray`.
3. Only one tray runs at a time. To change builds, choose **Exit** in the tray
   menu, then start the other one. **Settings and backup** shows the running
   version on its last line.
