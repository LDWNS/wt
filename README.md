# Worktree manager
Buddy, worktrees are now easy. Don't even worry about it.

## Usage

The workflow is all that matters here.
Follow it and be happy.

### Dependencies
- git
- [go](https://go.dev/doc/install)
- [fzf](https://github.com/junegunn/fzf) (`brew install fzf`)

### Setup

1. **_Crucial_**: `wt clone <ssh url>` - this clones as a bare repository
2. `cd` to repo
3. `git worktree add main`
4. `cd main`

### Workflow

`wt add <n> [b]` - sets up a worktree named `<n>`, optionally tracking branch `[b]`.
`wt rm` - launches `fzf` with the worktrees you can remove.
 
Easy right?

## Build

```bash
go build -o wt
```

> [!NOTE]
> Add this to your path, e.g. `export PATH=$HOME/<path to cloned directory>/wt:$PATH`

## Change directories

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

> [!NOTE]
> Don't forget to run `source .zshrc` or `zsh` for the changes to take effect.

## Prompt integration

`wt current-repo` prints `本 <repo>` when cwd is inside a linked worktree (not the main one), and nothing when it isn't. Add it to `~/.zshrc` for use in `PROMPT`/`RPROMPT` or Starship's `command` module:

```bash
RPROMPT='$(wt current-repo)'
```

The repo name comes from the repo's base dir (parent of `.git`), so it's the same across every worktree in the repo.

## Completion

Add the following line to your ~/.zshrc

```bash
source <(wt completion zsh)
```

> [!NOTE]
> Don't forget to run `source .zshrc` or `zsh` for the changes to take effect.

## All options

```bash
wt                          # fzf picker, enter to cd
wt add <n> [b]              # add worktree at ../<n>, symlink .wt-include dirs
wt rm [name]                # remove worktree (fzf if omitted)
wt clone <url> [name]       # SSH bare clone into ./<name>/.git, fix fetch refspec
wt list                     # list all worktrees
wt link                     # symlink .wt-include dirs into current worktree
wt current-repo             # print "本 <repo>" if cwd is a linked worktree, for shell prompts
wt completion zsh           # print zsh completion script
```
