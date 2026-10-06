//go:build !windows
// +build !windows

package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func targetBinDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "bin"), nil
}

// Tenta adicionar o ~/.local/bin ao PATH do usuário
func addDirToPath(dir string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	lines := []string{
		fmt.Sprintf(`export PATH="%s:$PATH"`, dir),
	}
	shells := []string{".zshrc", ".bashrc", ".profile"}
	for _, rc := range shells {
		rcPath := filepath.Join(home, rc)
		_ = ensureLineInFile(rcPath, lines[0])
	}
	return nil
}

func ensureLineInFile(path, line string) error {
	// se arquivo não existe, cria com a linha
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.WriteFile(path, []byte("\n# gommit\n"+line+"\n"), 0o644)
	}
	// se já contém a linha, não duplica
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(string(b), line) {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n# gommit\n" + line + "\n")
	return err
}
