package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Config holds user-configurable behaviors, loaded from configPath().
// Every field has a default; the config file only needs to set overrides.
type Config struct {
	WorktreeDir   string // template for new worktree dest, {name} substituted, relative to cwd unless absolute
	Icon          string // prompt icon printed by `wt current-repo`
	CloneHost     string // host used to expand "owner/repo" shorthand in `wt clone`
	DefaultBranch string // branch treated as the main worktree/branch
	FzfHeight     string // --height passed to fzf in the picker
}

func defaultConfig() Config {
	return Config{
		WorktreeDir:   "../{name}",
		Icon:          "木",
		CloneHost:     "github.com",
		DefaultBranch: "main",
		FzfHeight:     "40%",
	}
}

// configPath returns ~/.config/wt/config, honoring $XDG_CONFIG_HOME.
func configPath() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "wt", "config"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "wt", "config"), nil
}

// loadConfig reads configPath() over defaultConfig(). A missing file is not
// an error; unknown keys and malformed lines are reported but don't stop
// loading.
func loadConfig() (Config, error) {
	cfg := defaultConfig()

	path, err := configPath()
	if err != nil {
		return cfg, nil
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			fmt.Fprintf(os.Stderr, "wt: %s:%d: expected key=value, got %q\n", path, lineNo, line)
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(strings.Trim(strings.TrimSpace(val), `"`))

		switch key {
		case "worktree_dir":
			cfg.WorktreeDir = val
		case "icon":
			cfg.Icon = val
		case "clone_host":
			cfg.CloneHost = val
		case "default_branch":
			cfg.DefaultBranch = val
		case "fzf_height":
			cfg.FzfHeight = val
		default:
			fmt.Fprintf(os.Stderr, "wt: %s:%d: unknown config key %q\n", path, lineNo, key)
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

const configTemplate = `# wt config — one "key = value" per line, "#" for comments.
# Every key below is optional; omit a line to keep its default.

# Where new worktrees are created by "wt add <name>". "{name}" is
# substituted; relative paths resolve against the current directory.
# worktree_dir = ../{name}

# Icon printed by "wt current-repo" when cwd is inside a linked worktree.
# icon = 木

# Host used to expand "owner/repo" shorthand in "wt clone" into
# git@<clone_host>:owner/repo.git
# clone_host = github.com

# Branch treated as the main branch/worktree.
# default_branch = main

# --height passed to fzf by the worktree picker.
# fzf_height = 40%
`

// cmdConfig implements `wt config [path|edit]`.
func cmdConfig(args []string) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "path":
		fmt.Println(path)
		return nil

	case "edit":
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(configTemplate), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "created %s\n", path)
		} else if err != nil {
			return err
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		e := exec.Command(editor, path)
		e.Stdin = os.Stdin
		e.Stdout = os.Stdout
		e.Stderr = os.Stderr
		return e.Run()

	case "":
		exists := "not found; using defaults"
		if _, err := os.Stat(path); err == nil {
			exists = "loaded"
		}
		fmt.Printf("config: %s (%s)\n\n", path, exists)
		fmt.Printf("worktree_dir   = %s\n", cfg.WorktreeDir)
		fmt.Printf("icon           = %s\n", cfg.Icon)
		fmt.Printf("clone_host     = %s\n", cfg.CloneHost)
		fmt.Printf("default_branch = %s\n", cfg.DefaultBranch)
		fmt.Printf("fzf_height     = %s\n", cfg.FzfHeight)
		return nil

	default:
		return fmt.Errorf("usage: wt config [path|edit]")
	}
}
