//go:build windows

package platform

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCookiesPathPrefersNetwork(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "Network"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Network", "Cookies"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if got := cookiesPath(root); got != filepath.Join(root, "Network", "Cookies") {
		t.Fatalf("cookiesPath = %q", got)
	}
}

func TestCookiesPathFallsBackToRoot(t *testing.T) {
	root := t.TempDir()
	if got := cookiesPath(root); got != filepath.Join(root, "Cookies") {
		t.Fatalf("cookiesPath = %q", got)
	}
}

func TestSquirrelOwnsVersionedDesktopProcessesOnly(t *testing.T) {
	root := filepath.Join(t.TempDir(), "AnthropicClaude")
	platform := &windowsPlatform{
		executable:  filepath.Join(root, "claude.exe"),
		processRoot: root,
		kind:        installSquirrel,
	}
	if !platform.ownsProcess(filepath.Join(root, "app-1.20186.1", "claude.exe")) {
		t.Fatal("versioned Squirrel process was not recognized")
	}
	if !platform.ownsProcess(filepath.Join(root, "claude.exe")) {
		t.Fatal("Squirrel launcher was not recognized")
	}
	if platform.ownsProcess(filepath.Join(root, "resources", "claude.exe")) {
		t.Fatal("non-app Squirrel executable was recognized")
	}
	if platform.ownsProcess(filepath.Join(t.TempDir(), "claude.exe")) {
		t.Fatal("unrelated Claude executable was recognized")
	}
}

func TestSquirrelExecutableUsesStableLauncher(t *testing.T) {
	root := t.TempDir()
	launcher := filepath.Join(root, processName)
	if err := os.WriteFile(launcher, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if got := squirrelExecutable(root); got != launcher {
		t.Fatalf("squirrelExecutable = %q", got)
	}
}

func TestSquirrelExecutableFallsBackToNewestVersion(t *testing.T) {
	root := t.TempDir()
	oldExecutable := filepath.Join(root, "app-1.9.0", processName)
	newExecutable := filepath.Join(root, "app-1.10.0", processName)
	for _, executable := range []string{oldExecutable, newExecutable} {
		if err := os.MkdirAll(filepath.Dir(executable), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(executable, nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	if err := os.Chtimes(oldExecutable, now.Add(-time.Hour), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newExecutable, now, now); err != nil {
		t.Fatal(err)
	}
	if got := squirrelExecutable(root); got != newExecutable {
		t.Fatalf("squirrelExecutable = %q", got)
	}
}

func TestMSIXFallbackOwnsOnlyClaudePackageProcesses(t *testing.T) {
	root := filepath.Join(`C:\Program Files`, "WindowsApps")
	install := windowsInstall{kind: installMSIX, msix: true, processRoot: root}
	if !install.ownsProcess(filepath.Join(root, "Claude_1.0.0.0_x64__pzs8sxrjxfjjc", "app", "Claude.exe")) {
		t.Fatal("Claude MSIX process was not recognized")
	}
	if install.ownsProcess(filepath.Join(root, "Other_1.0.0.0_x64__example", "app", "Claude.exe")) {
		t.Fatal("unrelated MSIX process was recognized")
	}
	if !install.ownsProcess(filepath.Join(`D:\WindowsApps`, "Claude_1.0.0.0_x64__pzs8sxrjxfjjc", "app", "Claude.exe")) {
		t.Fatal("Claude MSIX process on another drive was not recognized")
	}
}

func TestPathWithinRejectsSiblingDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "AnthropicClaude")
	if !pathWithin(filepath.Join(root, "app-1", "claude.exe"), root) {
		t.Fatal("child path was rejected")
	}
	if pathWithin(root+"-old", root) {
		t.Fatal("sibling path was accepted")
	}
}

func TestStopWindowsProcessesAcceptsLateExit(t *testing.T) {
	var events []string
	waits := []bool{false, false, true}
	wait := func(timeout time.Duration) (bool, error) {
		events = append(events, "wait:"+timeout.String())
		exited := waits[0]
		waits = waits[1:]
		return exited, nil
	}
	taskkill := func(force bool) {
		events = append(events, "taskkill:"+strconv.FormatBool(force))
	}
	force := func(timeout time.Duration) (bool, error) {
		events = append(events, "force:"+timeout.String())
		return false, nil
	}

	if err := stopWindowsProcesses(wait, taskkill, force); err != nil {
		t.Fatal(err)
	}
	want := []string{"wait:2s", "taskkill:false", "wait:2s", "force:3s", "wait:5s"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestStopWindowsProcessesStillRejectsStuckProcess(t *testing.T) {
	wait := func(time.Duration) (bool, error) { return false, nil }
	err := stopWindowsProcesses(wait, func(bool) {}, func(time.Duration) (bool, error) {
		return false, nil
	})
	if err == nil || err.Error() != "claude desktop processes did not exit before timeout" {
		t.Fatalf("error = %v", err)
	}
}

func TestStopWindowsProcessesReturnsInspectionError(t *testing.T) {
	want := errors.New("process inspection failed")
	err := stopWindowsProcesses(func(time.Duration) (bool, error) {
		return false, want
	}, func(bool) {}, func(time.Duration) (bool, error) {
		return false, nil
	})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}

func TestShowCommandKeepsAMaximizedWindowMaximized(t *testing.T) {
	if got := showCommand(false); got != windowsSWShow {
		t.Fatalf("showCommand(visible) = %d, want SW_SHOW", got)
	}
	if got := showCommand(true); got != windowsSWRestore {
		t.Fatalf("showCommand(minimized) = %d, want SW_RESTORE", got)
	}
}

func TestEnumTopLevelWindowsCanRunMoreTimesThanGoAllowsCallbacks(t *testing.T) {
	for range 2500 {
		enumTopLevelWindows(func(uintptr) bool { return false })
	}
}

func TestMainInstanceIgnoresClaudeWindowsWithTheirOwnDataFolder(t *testing.T) {
	processes := []windowsProcess{
		{pid: 100, parent: 4},
		{pid: 101, parent: 100},
		{pid: 102, parent: 100},
		{pid: 103, parent: 102},
		{pid: 200, parent: 50},
		{pid: 201, parent: 200},
		{pid: 202, parent: 201},
		{pid: 300, parent: 777},
	}
	var asked []int
	got := mainInstanceProcesses(processes, func(pid int) bool {
		asked = append(asked, pid)
		return pid == 200
	})
	var pids []int
	for _, process := range got {
		pids = append(pids, process.pid)
	}
	if want := []int{100, 101, 102, 103, 300}; !reflect.DeepEqual(pids, want) {
		t.Fatalf("main window processes = %v, want %v", pids, want)
	}
	if want := []int{100, 200, 300}; !reflect.DeepEqual(asked, want) {
		t.Fatalf("command lines read for %v, want only the roots %v", asked, want)
	}
}

func TestMainInstanceSurvivesAParentLoop(t *testing.T) {
	processes := []windowsProcess{{pid: 1, parent: 2}, {pid: 2, parent: 1}}
	if got := mainInstanceProcesses(processes, func(int) bool { return false }); len(got) != 2 {
		t.Fatalf("processes = %v", got)
	}
}

func TestSeparateDataFolderLooksOnlyAtArguments(t *testing.T) {
	home := t.TempDir()
	mainFolder := filepath.Join(home, "Packages", "Claude_pzs8sxrjxfjjc", "LocalCache", "Roaming", "Claude")
	virtualFolder := filepath.Join(home, "Roaming", "Claude")
	executable := filepath.Join(home, "WindowsApps", "Claude_2.42304.0.0_x64__pzs8sxrjxfjjc", "app", "Claude.exe")
	quoted := func(value string) string { return `"` + value + `"` }
	separate := filepath.Join(home, "Claude-Accounts", "Main-account")
	tests := []struct {
		name        string
		commandLine string
		want        string
	}{
		{"main window", quoted(executable), ""},
		{"separate window", quoted(executable) + " --user-data-dir=" + quoted(separate), separate},
		{"main window helper", quoted(executable) + " --type=renderer --user-data-dir=" + quoted(virtualFolder) + " --lang=en-US", ""},
		{"main folder in capitals", quoted(executable) + " --user-data-dir=" + quoted(strings.ToUpper(mainFolder)), ""},
		{"deep link", quoted(executable) + " claude://code/continue?session=local_1", ""},
		{"flag inside the program path", quoted(filepath.Join(home, "--user-data-dir=x", "Claude.exe")), ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := separateDataFolder(test.commandLine, mainFolder, virtualFolder); got != test.want {
				t.Fatalf("separateDataFolder(%q) = %q, want %q", test.commandLine, got, test.want)
			}
		})
	}
}

func TestProcessCommandLineReadsARunningProcess(t *testing.T) {
	commandLine, err := processCommandLine(uint32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(commandLine), strings.ToLower(filepath.Base(executable))) {
		t.Fatalf("command line %q does not name %s", commandLine, filepath.Base(executable))
	}
}

func TestManifestExecutableStaysInsideThePackage(t *testing.T) {
	manifest := []byte(`<?xml version="1.0" encoding="utf-8"?>
<Package xmlns="http://schemas.microsoft.com/appx/manifest/foundation/windows10">
  <Applications>
    <Application Id="SshProxy" Executable="app/resources/claude-ssh-proxy.exe" EntryPoint="Windows.FullTrustApplication"/>
    <Application Id="Claude" Executable="app/Claude.exe" EntryPoint="Windows.FullTrustApplication"/>
  </Applications>
</Package>`)
	got, err := manifestExecutable(manifest, "Claude")
	if err != nil || got != filepath.Join("app", "Claude.exe") {
		t.Fatalf("manifestExecutable = %q, %v", got, err)
	}
	escaping := []byte(`<Package><Applications><Application Id="Claude" Executable="../../evil.exe"/></Applications></Package>`)
	if got, err := manifestExecutable(escaping, "Claude"); err == nil {
		t.Fatalf("an executable outside the package was accepted: %q", got)
	}
}

func TestClaudeExecutableFromTheInstalledPackage(t *testing.T) {
	if _, err := packageInstallPath(windowsPackageFamily); err != nil {
		t.Skip("Claude Desktop from the Store is not installed here")
	}
	executable, err := (&windowsPlatform{msix: true}).claudeExecutable()
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(executable); err != nil || info.IsDir() || !strings.EqualFold(filepath.Base(executable), processName) {
		t.Fatalf("executable %q: %v", executable, err)
	}
}

func TestMsixDataFolderFollowsClaudesMoveOutOfThePackage(t *testing.T) {
	family := strings.ToLower(windowsPackageFamily)
	tests := []struct {
		name  string
		files []string
		moved bool
	}{
		{"Claude 2 keeps its data in the package", []string{"packaged/Local State"}, false},
		{"Claude 3 moved it and left a marker", []string{"packaged/.user-data-left." + family, "moved/Local State"}, true},
		{"only the moved folder holds data", []string{"moved/config.json"}, true},
		{"moved back into the package", []string{"packaged/.user-data-back." + family, "packaged/Local State", "moved/Local State"}, false},
		{"both hold data without a marker", []string{"packaged/config.json", "moved/config.json"}, false},
		{"nothing yet", nil, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			local := t.TempDir()
			packaged := filepath.Join(local, "Packages", windowsPackageFamily, "LocalCache", "Roaming", "Claude")
			moved := filepath.Join(local, "Claude-Data", family)
			for _, file := range test.files {
				folder, name, _ := strings.Cut(file, "/")
				path := filepath.Join(map[string]string{"packaged": packaged, "moved": moved}[folder], name)
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			want := packaged
			if test.moved {
				want = moved
			}
			if got := msixDataFolder(local); got != want {
				t.Fatalf("msixDataFolder() = %q, want %q", got, want)
			}
		})
	}
}

func TestShareSessionRecordsLinksANewWindowToTheMainRecords(t *testing.T) {
	main, window := t.TempDir(), t.TempDir()
	writeTestFiles(t, main, map[string]string{"claude-code-sessions/account/org/local_a.json": "main"})
	if err := shareSessionRecords(window, main); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(window, "claude-code-sessions")
	if target, err := os.Readlink(link); err != nil || normalizeWindowsPath(target) != normalizeWindowsPath(filepath.Join(main, "claude-code-sessions")) {
		t.Fatalf("link target = %q, %v", target, err)
	}
	if data, err := os.ReadFile(filepath.Join(link, "account", "org", "local_a.json")); err != nil || string(data) != "main" {
		t.Fatalf("record read through the link = %q, %v", data, err)
	}
	if err := os.WriteFile(filepath.Join(link, "account", "org", "local_b.json"), []byte("window"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := testFiles(t, filepath.Join(main, "claude-code-sessions")); got["account/org/local_b.json"] != "window" {
		t.Fatalf("a record the window wrote is missing from the main folder: %v", got)
	}
	if err := shareSessionRecords(window, main); err != nil {
		t.Fatalf("sharing again: %v", err)
	}
	if entries, err := os.ReadDir(window); err != nil || len(entries) != 1 {
		t.Fatalf("the window folder holds %d entries after sharing twice, want only the link (%v)", len(entries), err)
	}
}

func TestShareSessionRecordsMovesTheWindowsOwnRecordsToTheMainFolder(t *testing.T) {
	main, window := t.TempDir(), t.TempDir()
	writeTestFiles(t, main, map[string]string{
		"claude-code-sessions/account/org/local_a.json":         "main",
		"claude-code-sessions/account/org/scheduled-tasks.json": "main tasks",
	})
	writeTestFiles(t, window, map[string]string{
		"claude-code-sessions/account/org/local_a.json":               "window copy",
		"claude-code-sessions/account/org/local_b.json":               "window",
		"claude-code-sessions/account/org/deleted_c":                  "deleted",
		"claude-code-sessions/account/org/scheduled-tasks.json":       "window tasks",
		"claude-code-sessions/account/org/waiting-input/local_b.json": "queued",
		"claude-code-sessions/other/org/local_d.json":                 "other account",
	})
	if err := shareSessionRecords(window, main); err != nil {
		t.Fatal(err)
	}
	wantMain := map[string]string{
		"account/org/local_a.json":         "main",
		"account/org/local_b.json":         "window",
		"account/org/deleted_c":            "deleted",
		"account/org/scheduled-tasks.json": "main tasks",
		"other/org/local_d.json":           "other account",
	}
	if got := testFiles(t, filepath.Join(main, "claude-code-sessions")); !reflect.DeepEqual(got, wantMain) {
		t.Fatalf("main records = %v\nwant %v", got, wantMain)
	}
	backups, err := filepath.Glob(filepath.Join(window, "claude-code-sessions.before-sharing-*"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("set-aside folders = %v, %v", backups, err)
	}
	wantLeft := map[string]string{
		"account/org/local_a.json":               "window copy",
		"account/org/scheduled-tasks.json":       "window tasks",
		"account/org/waiting-input/local_b.json": "queued",
	}
	if got := testFiles(t, backups[0]); !reflect.DeepEqual(got, wantLeft) {
		t.Fatalf("set-aside files = %v\nwant %v", got, wantLeft)
	}
	if got := testFiles(t, filepath.Join(window, "claude-code-sessions")); !reflect.DeepEqual(got, wantMain) {
		t.Fatalf("the window sees %v\nwant %v", got, wantMain)
	}
}

func TestShareSessionRecordsFollowsTheMainFolderWhenItMoves(t *testing.T) {
	old, moved, window := t.TempDir(), t.TempDir(), t.TempDir()
	writeTestFiles(t, old, map[string]string{"claude-code-sessions/account/org/local_a.json": "old"})
	if err := shareSessionRecords(window, old); err != nil {
		t.Fatal(err)
	}
	if err := shareSessionRecords(window, moved); err != nil {
		t.Fatal(err)
	}
	if target, err := os.Readlink(filepath.Join(window, "claude-code-sessions")); err != nil || normalizeWindowsPath(target) != normalizeWindowsPath(filepath.Join(moved, "claude-code-sessions")) {
		t.Fatalf("link target = %q, %v", target, err)
	}
	if got := testFiles(t, filepath.Join(old, "claude-code-sessions")); got["account/org/local_a.json"] != "old" {
		t.Fatalf("replacing the link touched the old folder's records: %v", got)
	}
}

func TestShareSessionRecordsRefusesTheMainFolderItself(t *testing.T) {
	main := t.TempDir()
	writeTestFiles(t, main, map[string]string{"claude-code-sessions/account/org/local_a.json": "main"})
	if err := shareSessionRecords(main, main); err == nil {
		t.Fatal("sharing the main folder with itself was accepted")
	}
	if got := testFiles(t, filepath.Join(main, "claude-code-sessions")); got["account/org/local_a.json"] != "main" {
		t.Fatalf("the main records changed: %v", got)
	}
}

func writeTestFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func testFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	files := make(map[string]string)
	// WalkDir does not enter a link given as its root, so the root is resolved.
	if resolved, err := os.Readlink(root); err == nil {
		root = resolved
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || !entry.Type().IsRegular() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		files[filepath.ToSlash(relative)] = string(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}
