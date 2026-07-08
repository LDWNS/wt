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

# build
```bash
go build -o wt
```
