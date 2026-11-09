//go:build darwin

package cmd

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// The dialogs only run when someone uses the tray, so their AppleScript is
// compiled here to catch a syntax error early.
func TestDialogScriptsCompile(t *testing.T) {
	scripts := map[string]string{
		"choice":  buttonDialogScript([]string{"Cancel", "No", "Yes"}, "Yes", "Cancel", "note"),
		"warning": buttonDialogScript([]string{"OK"}, "OK", "", "caution"),
		"secret":  secretPromptScript(),
		"open":    openBackupScript,
		"save":    saveBackupScript,
	}
	for name, script := range scripts {
		t.Run(name, func(t *testing.T) {
			args := append(scriptLines(script), "-o", filepath.Join(t.TempDir(), name+".scpt"))
			if out, err := exec.Command("osacompile", args...).CombinedOutput(); err != nil {
				t.Fatalf("osacompile: %v\n%s", err, out)
			}
		})
	}
}
