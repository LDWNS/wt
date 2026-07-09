package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	cmd := ""
	args := os.Args[1:]
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	var err error
	switch cmd {
	case "":
		err = cmdNavigate()
	case "main":
		err = toMain()
	case "list", "ls":
		err = cmdList()
	case "add":
		err = cmdAdd(args)
	case "rm", "remove":
		err = cmdRm(args)
	case "clone":
		err = cmdClone(args)
	case "link":
		err = cmdLink()
	case "completion":
		err = cmdCompletion(args)
	case "__complete":
		err = cmdComplete(args)
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s — try: wt help\n", cmd)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type worktreeEntry struct {
	path   string
	head   string
	branch string
}

func listWorktrees() ([]worktreeEntry, error) {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("not a git repo")
	}

	var wts []worktreeEntry
	var cur worktreeEntry

	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			if cur.path != "" {
				wts = append(wts, cur)
			}
			cur = worktreeEntry{path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			h := strings.TrimPrefix(line, "HEAD ")
			if len(h) > 7 {
				h = h[:7]
			}
			cur.head = h
		case strings.HasPrefix(line, "branch "):
			cur.branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		}
	}
	if cur.path != "" {
		wts = append(wts, cur)
	}

	return wts, nil
}

func worktreeRoot() (string, error) {
	wts, err := listWorktrees()
	if err != nil {
		return "", err
	}
	if len(wts) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}
	return wts[0].path, nil
}

func gitToplevel() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repo")
	}
	return strings.TrimSpace(string(out)), nil
}

func symlinkIncluded(src, dst string) error {
	f, err := os.Open(filepath.Join(src, ".wt-include"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		println(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		srcDir := filepath.Join(src, line)
		dstLink := filepath.Join(dst, line)
		if _, err := os.Stat(srcDir); err != nil {
			continue
		}
		if _, err := os.Lstat(dstLink); err == nil {
			continue
		}
		if err := os.Symlink(srcDir, dstLink); err != nil {
			fmt.Fprintf(os.Stderr, "  link %s: %v\n", line, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "  linked: %s\n", line)
	}
	return scanner.Err()
}

func fzfSelect(items []worktreeEntry, prompt string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no worktrees")
	}

	var sb strings.Builder
	for _, wt := range items {
		name := filepath.Base(wt.path)
		branch := wt.branch
		if branch == "" {
			branch = "(detached)"
		}
		fmt.Fprintf(&sb, "%s\t%-24s %s  [%s]\n", wt.path, name, wt.head, branch)
	}

	fzf := exec.Command("fzf",
		"--height=40%",
		"--reverse",
		"--prompt="+prompt,
		"--delimiter=\t",
		"--with-nth=2",
	)
	fzf.Stdin = strings.NewReader(sb.String())
	fzf.Stderr = os.Stderr

	out, err := fzf.Output()
	if err != nil {
		return "", nil // cancelled
	}

	selected := strings.TrimSpace(string(out))
	parts := strings.SplitN(selected, "\t", 2)
	if len(parts) < 1 || parts[0] == "" {
		return "", nil
	}
	return parts[0], nil
}

func toMain() error {
	wts, err := listWorktrees()
	if err != nil {
		return err
	}
	for _, v := range wts {
		if v.branch == "main" {
			fmt.Println(v.path)
		}
	}

	return nil
}
func cmdNavigate() error {
	wts, err := listWorktrees()
	if err != nil {
		return err
	}

	path, err := fzfSelect(wts, "worktree: ")
	if err != nil {
		return err
	}
	if path != "" {
		fmt.Println(path)
	}
	return nil
}

func cmdList() error {
	cmd := exec.Command("git", "worktree", "list")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func cmdAdd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: wt add <name> [branch]")
	}
	name := args[0]
	branch := name
	if len(args) > 1 {
		branch = args[1]
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	dest := filepath.Join(filepath.Dir(cwd), name)

	add := exec.Command("git", "-C", cwd, "worktree", "add", dest, branch)
	add.Stdout = os.Stderr
	add.Stderr = os.Stderr
	if err := add.Run(); err != nil {
		add = exec.Command("git", "-C", cwd, "worktree", "add", "-b", branch, dest)
		add.Stdout = os.Stderr
		add.Stderr = os.Stderr
		if err := add.Run(); err != nil {
			return err
		}
	}
	fmt.Println(dest)
	return symlinkIncluded(cwd, dest)
}

func cmdRm(args []string) error {
	wts, err := listWorktrees()
	if err != nil {
		return err
	}

	var wtPath string

	if len(args) == 0 {
		if len(wts) <= 1 {
			return fmt.Errorf("no worktrees to remove")
		}
		wtPath, err = fzfSelect(wts[1:], "remove: ")
		if err != nil || wtPath == "" {
			return err
		}
	} else {
		name := args[0]
		if filepath.IsAbs(name) {
			wtPath = name
		} else {
			for _, wt := range wts {
				if filepath.Base(wt.path) == name {
					wtPath = wt.path
					break
				}
			}
			if wtPath == "" {
				return fmt.Errorf("worktree %q not found", name)
			}
		}
	}

	mainRoot, err := worktreeRoot()
	if err != nil {
		return err
	}

	entries, _ := os.ReadDir(wtPath)
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink == 0 {
			continue
		}
		linkPath := filepath.Join(wtPath, entry.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}
		if strings.HasPrefix(target, mainRoot+string(os.PathSeparator)) {
			os.Remove(linkPath)
			fmt.Fprintf(os.Stderr, "  removed symlink: %s\n", entry.Name())
		}
	}

	rm := exec.Command("git", "worktree", "remove", wtPath)
	rm.Stdout = os.Stderr
	rm.Stderr = os.Stderr
	return rm.Run()
}

func sshURLName(url string) string {
	// git@github.com:user/repo.git -> repo
	base := url
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	} else if idx := strings.Index(base, ":"); idx >= 0 {
		base = base[idx+1:]
	}
	return strings.TrimSuffix(base, ".git")
}

func cmdClone(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: wt clone <ssh-url> [name]")
	}
	url := args[0]
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return fmt.Errorf("only SSH clone supported (use git@host:user/repo.git)")
	}

	name := sshURLName(url)
	if len(args) > 1 {
		name = args[1]
	}
	if name == "" {
		return fmt.Errorf("could not derive repo name from URL; pass a name explicitly")
	}

	if err := os.MkdirAll(name, 0o755); err != nil {
		return err
	}

	dotgit := filepath.Join(name, ".git")

	clone := exec.Command("git", "clone", "--bare", "--", url, dotgit)
	clone.Stdout = os.Stderr
	clone.Stderr = os.Stderr
	if err := clone.Run(); err != nil {
		return err
	}

	// bare clone omits the fetch refspec; add it so remote-tracking branches populate
	cfg := exec.Command("git", "-C", dotgit, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	cfg.Stdout = os.Stderr
	cfg.Stderr = os.Stderr
	if err := cfg.Run(); err != nil {
		return err
	}

	fetch := exec.Command("git", "-C", dotgit, "fetch")
	fetch.Stdout = os.Stderr
	fetch.Stderr = os.Stderr
	if err := fetch.Run(); err != nil {
		return err
	}

	cd_name := exec.Command("cd", name)
	cd_name.Stdout = os.Stderr
	cd_name.Stderr = os.Stderr
	if err := cd_name.Run(); err != nil {
		return err
	}

	init_main := exec.Command("git", "worktree", "add", "main")
	init_main.Stdout = os.Stderr
	init_main.Stderr = os.Stderr
	if err := init_main.Run(); err != nil {
		return err
	}

	cd_main := exec.Command("cd", "main")
	cd_main.Stdout = os.Stderr
	cd_main.Stderr = os.Stderr
	if err := cd_main.Run(); err != nil {
		return err
	}

	pwd := exec.Command("pwd")
	pwd.Stdout = os.Stderr
	pwd.Stderr = os.Stderr
	if out, err := pwd.Output(); err == nil {
		fmt.Print(string(out))
	}
	return nil
}

func cmdLink() error {
	current, err := gitToplevel()
	if err != nil {
		return err
	}
	mainRoot, err := worktreeRoot()
	if err != nil {
		return err
	}
	if current == mainRoot {
		return fmt.Errorf("already in main worktree")
	}
	return symlinkIncluded(mainRoot, current)
}

func cmdComplete(args []string) error {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "worktrees":
		wts, err := listWorktrees()
		if err != nil {
			return nil
		}
		for _, wt := range wts[1:] {
			fmt.Println(filepath.Base(wt.path))
		}
	case "branches":
		out, err := exec.Command("git", "branch", "--format=%(refname:short)").Output()
		if err != nil {
			return nil
		}
		fmt.Print(string(out))
	}
	return nil
}

func cmdCompletion(args []string) error {
	shell := "zsh"
	if len(args) > 0 {
		shell = args[0]
	}
	if shell != "zsh" {
		return fmt.Errorf("unsupported shell: %s (only zsh supported)", shell)
	}
	fmt.Print(`#compdef wt

_wt() {
    local state

    _arguments -C '1: :->command' '*: :->args'

    case $state in
        command)
            local commands
            commands=(
                'main:print main worktree path'
                'list:list all worktrees'
                'ls:list all worktrees'
                'add:add a new worktree'
                'rm:remove a worktree'
                'remove:remove a worktree'
                'clone:clone a repo via SSH'
                'link:symlink .wt-include dirs into current worktree'
                'help:show help'
                'completion:print shell completion script'
            )
            _describe 'command' commands
            ;;
        args)
            case $words[2] in
                rm|remove)
                    local worktrees
                    worktrees=($(wt __complete worktrees 2>/dev/null))
                    _values 'worktree' $worktrees
                    ;;
                add)
                    if [[ $CURRENT -eq 3 ]]; then
                        local branches
                        branches=($(wt __complete branches 2>/dev/null))
                        _values 'branch' $branches
                    fi
                    ;;
                completion)
                    _values 'shell' zsh
                    ;;
            esac
            ;;
    esac
}

compdef _wt wt
`)
	return nil
}

func printHelp() {
	fmt.Print(`wt — worktree manager
  wt              fzf picker, enter to cd
  wt add <n> [b]           add worktree at ../<n>, symlink .wt-include dirs
  wt rm [name]             remove worktree (fzf if omitted)
  wt clone <url> [name]   SSH bare clone into ./<name>/.git, fix fetch refspec
  wt list                  list all worktrees
  wt link                  symlink .wt-include dirs into current worktree
  wt completion zsh        print zsh completion script
`)
}
