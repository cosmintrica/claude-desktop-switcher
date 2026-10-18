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
	tests := []struct {
		name        string
		commandLine string
		want        bool
	}{
		{"main window", quoted(executable), false},
		{"separate window", quoted(executable) + " --user-data-dir=" + quoted(filepath.Join(home, "Claude-Accounts", "Main-account")), true},
		{"main window helper", quoted(executable) + " --type=renderer --user-data-dir=" + quoted(virtualFolder) + " --lang=en-US", false},
		{"main folder in capitals", quoted(executable) + " --user-data-dir=" + quoted(strings.ToUpper(mainFolder)), false},
		{"deep link", quoted(executable) + " claude://code/continue?session=local_1", false},
		{"flag inside the program path", quoted(filepath.Join(home, "--user-data-dir=x", "Claude.exe")), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := separateDataFolder(test.commandLine, mainFolder, virtualFolder); got != test.want {
				t.Fatalf("separateDataFolder(%q) = %v, want %v", test.commandLine, got, test.want)
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
