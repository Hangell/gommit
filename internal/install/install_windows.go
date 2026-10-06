//go:build windows
// +build windows

package install

import (
	"os"
	"os/exec"
	"path/filepath"
)

func targetBinDir() (string, error) {
	// Preferência: %USERPROFILE%\go\bin se existir; senão, %LOCALAPPDATA%\Programs\gommit\bin
	home, _ := os.UserHomeDir()
	goBin := filepath.Join(home, "go", "bin")
	if st, err := os.Stat(goBin); err == nil && st.IsDir() {
		return goBin, nil
	}
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		local = filepath.Join(home, "AppData", "Local")
	}
	return filepath.Join(local, "Programs", "gommit", "bin"), nil
}

// Adiciona dir ao PATH do usuário via PowerShell (persistente)
func addDirToPath(dir string) error {
	// Usa PowerShell para concatenar no PATH do usuário (sem mexer no PATH do sistema)
	ps := `[Environment]::SetEnvironmentVariable('Path', ($env:Path + ';` + dir + `'), 'User')`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", ps)
	return cmd.Run()
}
