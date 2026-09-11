# Worktree manager

Buddy, worktrees are EASY now. Don't even worry about it.

## Usage

**_Crucial_**: `wt clone <ssh url|owner/repo>` - this clones as a bare repository. `owner/repo` shorthand expands to `git@github.com:owner/repo.git`
This creates the desired package structure:
```
wt
├── main            # main branch
├── feat
│   └── create-makefile
└── docs
    └── update-readme
```

### Workflow

`wt add <n> [b]` - sets up a worktree named `<n>`, optionally tracking branch `[b]`.
`wt` - launches `fzf` with the worktrees you can check out.
`wt switch` - alias for bare `wt`; only this form triggers the herdr integration below.
`wt rm` - launches `fzf` with the worktrees you can remove.

Easy right?

## Install

New machine, one command:

```bash
make install
```

This runs, in order:

1. `deps` - checks `git`, `go`, `fzf` are on `$PATH`, fails fast on the missing one
2. `build` - `go build -o wt .`
3. copies `wt` into `$(go env GOBIN)`, falling back to `~/go/bin`

> [!NOTE] 
> `~/go/bin` is where `go install`/this Makefile puts binaries - it's separate from wherever the `go` command itself lives (e.g. `/opt/homebrew/bin/go`). Having `go` on `$PATH` does **not** mean `~/go/bin` is. `make install` checks and warns if it's missing, e.g.:
>
> ```bash
> export PATH="$HOME/go/bin:$PATH"
> ```
>
> Add that to `~/.zshrc` if you see the warning, then `source ~/.zshrc`.

Other targets:

```bash
make build      # just go build -o wt
make deps       # check git, go, fzf are installed
make uninstall  # remove the installed binary
make clean      # remove the local ./wt build artifact
make help       # list targets (also runs on bare `make`)
```

Once `make install` finishes without a PATH warning, `wt` resolves to the real binary. The steps below (shell wrapper, completion) are still manual - `make install` prints them as a reminder each run.

Prefer to build by hand instead of `make`?

```bash
go build -o wt
export PATH=$HOME/<path to cloned directory>/wt:$PATH
```

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

### herdr integration

If [herdr](https://herdr.dev) is installed and you're running inside a herdr pane
(`$HERDR_ENV` set), `wt switch` (an explicit alias for the bare fuzzy-picker) can also ensure
there's one herdr tab per worktree — reusing the existing tab for a worktree instead of opening
a duplicate, and closing it on `wt rm`. Other commands (`add`, `clone`, `link`, ...) still just
`cd`, with no herdr involvement. Extend the wrapper above with:

```bash
# wt — worktree manager (shell wrapper for cd + herdr tab support)
wt() {
  local out exit_code
  out=$(command wt "$@")
  exit_code=$?
  if [[ -n "$out" && -d "$out" ]]; then
    cd "$out"
    if [[ "$1" == "switch" && -n "$HERDR_ENV" ]] && command -v herdr >/dev/null 2>&1; then
      local main_root
      main_root=$(command wt main 2>/dev/null)
      if [[ -n "$main_root" ]]; then
        herdr worktree open --path "$out" --cwd "$main_root" --focus \
          --trust-repository --label "$(basename "$out")" >/dev/null 2>&1
      fi
    fi
  elif [[ -n "$out" ]]; then
    if [[ ( "$1" == "rm" || "$1" == "remove" ) && -n "$HERDR_ENV" ]] \
      && command -v herdr >/dev/null 2>&1 && command -v jq >/dev/null 2>&1; then
      local main_root wsid
      main_root=$(command wt main 2>/dev/null)
      if [[ -n "$main_root" ]]; then
        wsid=$(herdr worktree list --cwd "$main_root" --trust-repository 2>/dev/null \
          | jq -r --arg p "$out" '.result.worktrees[] | select(.path == $p) | .open_workspace_id // empty')
        [[ -n "$wsid" ]] && herdr workspace close "$wsid" >/dev/null 2>&1
      fi
    else
      print -- "$out"
    fi
  fi
  return $exit_code
}
```

`herdr worktree open` is idempotent — calling it again for a path that already has a workspace
open just focuses it, no duplicate tab. All herdr calls are best-effort: if herdr, `jq`,
`$HERDR_ENV` are missing, or you didn't run `wt switch`, `wt` behaves exactly as before (plain
`cd`, no herdr calls).

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

## Configuration

`wt` reads `~/.config/wt/config` (or `$XDG_CONFIG_HOME/wt/config`), a plain `key = value` file, `#` for comments. Every key is optional — omit a line to keep its default.

```bash
wt config          # show effective config + whether the file was found
wt config path     # print the config file path
wt config edit     # create it (with a commented template) and open in $EDITOR
```

| key              | default      | controls                                                        |
|------------------|--------------|------------------------------------------------------------------|
| `worktree_dir`   | `../{name}`  | where `wt add <name>` creates the worktree; `{name}` substituted |
| `icon`           | `木`         | icon `wt current-repo` prints                                    |
| `clone_host`     | `github.com` | host used to expand `owner/repo` shorthand in `wt clone`          |
| `default_branch` | `main`       | branch treated as the repo's main branch                         |
| `fzf_height`     | `40%`        | `--height` passed to fzf in the picker                            |

## Dependencies

- git
- [go](https://go.dev/doc/install)
- [fzf](https://github.com/junegunn/fzf)

`make deps` checks these are on `$PATH` for you.

## All options

```bash
wt                          # fzf picker, enter to cd
wt switch                   # alias for bare "wt"; herdr integration (if configured) only fires on this
wt add <n> [b]              # add worktree at ../<n>, symlink .wt-include dirs
wt rm [name]                # remove worktree (fzf if omitted)
wt clone <url|owner/repo> [name]  # SSH bare clone into ./<name>/.git, fix fetch refspec
                                  # "owner/repo" shorthand expands to git@github.com:owner/repo.git
wt list                     # list all worktrees
wt link                     # symlink .wt-include dirs into current worktree
wt config [path|edit]       # show effective config, or print/edit ~/.config/wt/config
wt current-repo             # print "本 <repo>" if cwd is a linked worktree, for shell prompts
wt completion zsh           # print zsh completion script
```
