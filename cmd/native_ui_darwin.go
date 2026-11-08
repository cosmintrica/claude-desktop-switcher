//go:build darwin

package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Dialogs run through osascript: its standard dialogs need no extra
// permission, and argv carries the text so nothing has to be escaped.

var errDialogCancelled = errors.New("dialog cancelled")

// scriptLines wraps a dialog in a handler that reads its text from argv and
// brings the dialog in front of other windows.
func scriptLines(script string) []string {
	command := []string{}
	for _, line := range append(append([]string{"on run argv", "activate"}, strings.Split(script, "\n")...), "end run") {
		command = append(command, "-e", line)
	}
	return command
}

func runDialog(script string, args ...string) (string, error) {
	command := append(scriptLines(script), "--")
	command = append(command, args...)
	out, err := exec.Command("osascript", command...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && strings.Contains(string(exitErr.Stderr), "-128") {
			return "", errDialogCancelled
		}
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// dialogIcon is the app's own icon inside its bundle, or a standard icon when
// the switcher runs outside one.
func dialogIcon(fallback string) string {
	if exe, err := os.Executable(); err == nil {
		icon := filepath.Join(filepath.Dir(exe), "..", "Resources", "AppIcon.icns")
		if _, err := os.Stat(icon); err == nil {
			return "(POSIX file \"" + strings.ReplaceAll(filepath.Clean(icon), `"`, `\"`) + "\")"
		}
	}
	return fallback
}

func buttonDialog(title, message string, buttons []string, defaultButton, cancelButton, icon string) (string, error) {
	return runDialog(buttonDialogScript(buttons, defaultButton, cancelButton, icon), message, title)
}

func buttonDialogScript(buttons []string, defaultButton, cancelButton, icon string) string {
	quoted := make([]string, len(buttons))
	for i, button := range buttons {
		quoted[i] = `"` + button + `"`
	}
	script := `set reply to display dialog (item 1 of argv) with title (item 2 of argv) buttons {` + strings.Join(quoted, ", ") + `} default button "` + defaultButton + `"`
	if cancelButton != "" {
		script += ` cancel button "` + cancelButton + `"`
	}
	return script + " with icon " + dialogIcon(icon) + "\nreturn button returned of reply"
}

func nativeTrayChoice(title, message string) (trayChoiceValue, error) {
	button, err := buttonDialog(title, message, []string{"Cancel", "No", "Yes"}, "Yes", "Cancel", "note")
	switch {
	case errors.Is(err, errDialogCancelled):
		return trayCancel, nil
	case err != nil:
		return trayCancel, err
	case button == "Yes":
		return trayYes, nil
	case button == "No":
		return trayNo, nil
	}
	return trayCancel, nil
}

func nativeTrayConfirm(title, message string) (bool, error) {
	button, err := buttonDialog(title, message, []string{"No", "Yes"}, "Yes", "No", "note")
	if errors.Is(err, errDialogCancelled) {
		return false, nil
	}
	return button == "Yes", err
}

func trayWarning(title, message string) {
	_, _ = buttonDialog(title, message, []string{"OK"}, "OK", "", "caution")
}

func trayDeleteConfirm(name string) (bool, error) {
	button, err := buttonDialog(
		"Delete account?",
		name+"\n\nRemove its saved sign-in data from "+ProductName+"?\nYour Claude account will not be deleted.",
		[]string{"Cancel", "Delete"}, "Cancel", "Cancel", "caution",
	)
	if errors.Is(err, errDialogCancelled) {
		return false, nil
	}
	return button == "Delete", err
}

func nativeTraySecretPrompt(title, message string) (string, error) {
	return runDialog(secretPromptScript(), message, title)
}

func secretPromptScript() string {
	return `set reply to display dialog (item 1 of argv) with title (item 2 of argv) default answer "" with hidden answer buttons {"Cancel", "OK"} default button "OK" cancel button "Cancel" with icon ` + dialogIcon("note") + "\nreturn text returned of reply"
}

const (
	openBackupScript = `return POSIX path of (choose file with prompt "Choose a ` + ProductName + ` backup" of type {"csb"})`
	saveBackupScript = `return POSIX path of (choose file name with prompt "Save the backup as" default name (item 1 of argv) default location (path to documents folder))`
)

func nativeTrayFileDialog(open bool, defaultName string) (string, error) {
	var path string
	var err error
	if open {
		path, err = runDialog(openBackupScript)
	} else {
		path, err = runDialog(saveBackupScript, defaultName)
	}
	if errors.Is(err, errDialogCancelled) {
		return "", nil
	}
	return path, err
}
