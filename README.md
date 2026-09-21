# Wohta worktree manager

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
`wt rm` - launches `fzf` with the worktrees you can remove.

Easy right?

## Install

New machine, one command:

```bash
make install
```

This runs, in order:

1. `deps` - checks `git`, `go`, `fzf` are on `$PATH`, fails fast on the missing one
2. `build` - `go build -o wohta .`
3. copies `wohta` into `$(go env GOBIN)`, falling back to `~/go/bin`

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
make build      # just go build -o wohta
make deps       # check git, go, fzf are installed
make uninstall  # remove the installed binary
make clean      # remove the local ./wohta build artifact
make help       # list targets (also runs on bare `make`)
```

Once `make install` finishes without a PATH warning, `wohta` resolves to the real binary. The steps below (shell wrapper, completion) are still manual - `make install` prints them as a reminder each run.

Prefer to build by hand instead of `make`?

```bash
go build -o wohta
export PATH=$HOME/<path to cloned directory>/wohta:$PATH
```

## Change directories

To make changing directories work, add this to your `~/.zshrc`.
 
> [!IMPORTANT]
> Highly recommended

```bash
# wt — worktree manager (shell wrapper for cd support)
wt() {
  local out exit_code
  out=$(command wohta "$@")
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

## Sharing files across worktrees (`.wt-include`)

`wt add` and `wt link` read `.wt-include` in the main worktree — one path per line, `#` for comments — and bring each one into the new/current worktree. By default a path is **symlinked** back to the main worktree's copy (good for shared, heavy dirs like `node_modules`). Prefix a line with `copy ` to **copy** it instead, giving the worktree its own independent, non-symlinked file or directory (good for per-worktree files like `.env` that shouldn't be shared or edited in lockstep).

```
# .wt-include
node_modules
.venv
copy .env
copy .env.local
```

`wt include <path...>` appends to `.wt-include` (creating it and git-excluding it via `info/exclude`); pass `--copy` to add the paths in copy mode:

```bash
wt include node_modules          # symlink
wt include --copy .env .env.local  # copy
```

Existing paths at the destination are left alone — neither mode overwrites a file/symlink that's already there.

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
wt add <n> [b]              # add worktree at ../<n>, symlink/copy .wt-include paths
wt rm [name]                # remove worktree (fzf if omitted)
wt clone <url|owner/repo> [name]  # SSH bare clone into ./<name>/.git, fix fetch refspec
                                  # "owner/repo" shorthand expands to git@github.com:owner/repo.git
wt list                     # list all worktrees
wt link                     # symlink/copy .wt-include paths into current worktree
wt include [path...]        # add path(s) to .wt-include; --copy marks them copy-instead-of-link
wt config [path]            # show effective config, or print ~/.config/wt/config
wt current-repo             # print "本 <repo>" if cwd is a linked worktree, for shell prompts
wt completion zsh           # print zsh completion script
```
