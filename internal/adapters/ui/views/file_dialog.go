package views

import (
	"os/exec"
	"runtime"
	"strings"
)

// pickMultipleFiles opens a native OS file dialog allowing multiple PDF/EPUB selection.
// Returns the list of absolute paths chosen by the user (empty if cancelled).
func pickMultipleFiles() []string {
	var rawOut []byte
	var err error

	switch runtime.GOOS {
	case "darwin":
		// No "of type" restriction — UTI codes are unreliable across macOS versions.
		// We accept any file and let the extractor reject unsupported formats.
		script := `set output to ""
set theFiles to choose file with prompt "Importer des livres (PDF, EPUB)" with multiple selections allowed
repeat with f in theFiles
	set output to output & POSIX path of f & linefeed
end repeat
return output`
		rawOut, err = exec.Command("osascript", "-e", script).Output()

	case "linux":
		rawOut, err = exec.Command("zenity",
			"--file-selection", "--multiple", "--separator=\n",
			"--file-filter=Livres (*.pdf *.epub)|*.pdf *.epub",
			"--title=Importer des livres").Output()
		if err != nil {
			rawOut, err = exec.Command("kdialog",
				"--getopenfilename", ".", "*.pdf *.epub",
				"--title", "Importer des livres", "--multiple").Output()
		}

	case "windows":
		ps := `Add-Type -AssemblyName System.Windows.Forms; ` +
			`$d = New-Object System.Windows.Forms.OpenFileDialog; ` +
			`$d.Filter="Livres (*.pdf;*.epub)|*.pdf;*.epub"; ` +
			`$d.Multiselect=$true; $d.ShowDialog()|Out-Null; ` +
			`$d.FileNames -join "\n"`
		rawOut, err = exec.Command("powershell", "-NoProfile", "-Command", ps).Output()
	}

	if err != nil || len(rawOut) == 0 {
		return nil
	}
	var paths []string
	for _, line := range strings.Split(string(rawOut), "\n") {
		if p := strings.TrimSpace(line); p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

// openFolderDialog opens a native folder picker dialog.
func openFolderDialog(defaultDir string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		script := `POSIX path of (choose folder with prompt "Choisir le dossier de destination")`
		out, err := exec.Command("osascript", "-e", script).Output()
		if err == nil {
			if r := strings.TrimSpace(string(out)); r != "" {
				return r, nil
			}
		}
	case "linux":
		for _, args := range [][]string{
			{"zenity", "--file-selection", "--directory", "--title=Choisir le dossier"},
			{"kdialog", "--getexistingdirectory", defaultDir},
		} {
			out, err := exec.Command(args[0], args[1:]...).Output()
			if err == nil {
				if r := strings.TrimSpace(string(out)); r != "" {
					return r, nil
				}
			}
		}
	default:
		ps := `(New-Object -ComObject Shell.Application).BrowseForFolder(0,'Destination',0).Self.Path`
		out, err := exec.Command("powershell", "-NoProfile", "-Command", ps).Output()
		if err == nil {
			if r := strings.TrimSpace(string(out)); r != "" {
				return r, nil
			}
		}
	}
	return defaultDir, nil
}
