package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Result struct {
	TargetDir   string
	BinPath     string
	WasPresent  bool   // já existia bin em TargetDir?
	OldVersion  string // versão anterior (se conseguir detectar)
	NewVersion  string // versão atual (passada por parâmetro)
	PathUpdated bool   // PATH do usuário foi atualizado
	Message     string // mensagem amigável para o usuário
}

var ErrUnsupported = errors.New("unsupported platform")

// InstallSelf copia o executável atual para o diretório de bin do usuário,
// garante que esse diretório esteja no PATH do usuário e retorna um resumo.
// newVersion é a string de versão reportada pelo gommit (ex.: "0.1.0").
func InstallSelf(newVersion string) (Result, error) {
	res := Result{NewVersion: newVersion}

	exe, err := os.Executable()
	if err != nil {
		return res, fmt.Errorf("cannot resolve current executable: %w", err)
	}
	exe, _ = filepath.EvalSymlinks(exe)

	targetDir, err := targetBinDir()
	if err != nil {
		return res, err
	}
	res.TargetDir = targetDir
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return res, fmt.Errorf("cannot create target dir: %w", err)
	}

	binName := "gommit"
	if isWindows() {
		binName += ".exe"
	}
	dest := filepath.Join(targetDir, binName)
	res.BinPath = dest

	// detecta versão anterior (se já existir)
	if _, err := os.Stat(dest); err == nil {
		res.WasPresent = true
		old, _ := exec.Command(dest, "--version").Output()
		res.OldVersion = strings.TrimSpace(string(old))
	}

	// copia/atualiza binário
	if err := copyFile(exe, dest); err != nil {
		return res, fmt.Errorf("failed to install binary: %w", err)
	}

	// garante execução (unix)
	_ = os.Chmod(dest, 0o755)

	// PATH
	onPath, err := dirOnPath(targetDir)
	if err != nil {
		return res, fmt.Errorf("failed to inspect PATH: %w", err)
	}
	if !onPath {
		if err := addDirToPath(targetDir); err == nil {
			res.PathUpdated = true
		}
	}

	// Mensagem amigável (i18n simples via LANG)
	lang := strings.ToLower(os.Getenv("LANG"))
	pt := strings.HasPrefix(lang, "pt")

	switch {
	case res.WasPresent && normalizeVer(res.OldVersion) == ("gommit version "+newVersion):
		if pt {
			res.Message = fmt.Sprintf("Gommit já estava instalado (%s) em %s.\nFeche e reabra o terminal se o comando ainda não for reconhecido.", res.OldVersion, targetDir)
		} else {
			res.Message = fmt.Sprintf("Gommit was already installed (%s) at %s.\nClose and reopen your terminal if the command isn't recognized yet.", res.OldVersion, targetDir)
		}
	case res.WasPresent:
		if pt {
			res.Message = fmt.Sprintf("Gommit atualizado de %s para %s em %s.\nFeche e reabra o terminal para recarregar o PATH.", res.OldVersion, "gommit version "+newVersion, targetDir)
		} else {
			res.Message = fmt.Sprintf("Gommit updated from %s to %s at %s.\nClose and reopen your terminal to reload PATH.", res.OldVersion, "gommit version "+newVersion, targetDir)
		}
	default:
		if pt {
			res.Message = fmt.Sprintf("Gommit instalado em %s.\nAgora é só fechar e abrir o terminal; depois, rode: gommit --version", targetDir)
		} else {
			res.Message = fmt.Sprintf("Gommit installed to %s.\nClose and reopen your terminal; then run: gommit --version", targetDir)
		}
	}

	if !onPath {
		if pt {
			res.Message += fmt.Sprintf("\n\nObservação: adicionei %s ao seu PATH de usuário. Se não surtir efeito imediato, reabra o terminal.", targetDir)
		} else {
			res.Message += fmt.Sprintf("\n\nNote: I added %s to your user PATH. If it doesn't take effect immediately, reopen the terminal.", targetDir)
		}
	}

	return res, nil
}

func normalizeVer(s string) string { return strings.TrimSpace(s) }

// util cross-plataforma simples
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := out.ReadFrom(in); err != nil {
		return err
	}
	return out.Sync()
}

func dirOnPath(dir string) (bool, error) {
	p := os.Getenv("PATH")
	sep := string(os.PathListSeparator)
	for _, ent := range strings.Split(p, sep) {
		if filepath.Clean(ent) == filepath.Clean(dir) {
			return true, nil
		}
	}
	return false, nil
}

func isWindows() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") || (filepath.Separator == '\\')
}
