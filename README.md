# worktree manager

```bash
wt                          # fzf picker, enter to cd
wt add <n> [b]              # add worktree at ../<n>, symlink .wt-include dirs
wt rm [name]                # remove worktree (fzf if omitted)
wt clone <url> [name]       # SSH bare clone into ./<name>/.git, fix fetch refspec
wt list                     # list all worktrees
wt link                     # symlink .wt-include dirs into current worktree
wt completion zsh           # print zsh completion script
```

# Build
```bash
go build -o wt
```
> [!NOTE]
> Add this to your path, e.g. `export PATH=$HOME/<path to cloned directory>/wt:$PATH`

# Change directories
To make changing directories work, add this to your `~/.zshrc`.
```bash
# wt — worktree manager (shell wrapper for cd support)
wt() {
  local out exit_code
  out=$(command wt "$@")
  exit_code=$?
  if [[ -n "$out" && -d "$out" ]]; then
    cd "$out"
  elif [[ -n "$out" ]]; then
    print -- "$out"
  fi
  return $exit_code
}
```

# Completion

Add the following line to your ~/.zshrc
```bash
source <(wt completion zsh)
```
```
