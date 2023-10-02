# cli

[![godoc](https://raw.githubusercontent.com/primecitizens/cicd/cicd/cli/godoc.svg)](https://pkg.go.dev/github.com/primecitizens/cli)
![coverage](https://raw.githubusercontent.com/primecitizens/cicd/cicd/cli/coverage.svg)

A novel experience for building application entrypoint.

## Features

- POSIX & GNU style flag parsing, typed and customizable. (`flag*.go`)
- Commands and sub-commands, nothing hidden. (`cmd*.go`)
- Shell completions made straightforward. (`comp*.go`, `scripts/*`)
  - Use `CompCmdShells` to provide shell completion support for `bash`, `zsh` and `pwsh` (powershell).
- From mostly static to highly dynamic, choices available for zero-allocation* and productivity preferences.
  - Choose `ReflectIndexer` for productivity, choose `FuncIndexer` for zero-allocation. (see [Core Concepts](#core-concepts) and module document)
- Solid & decoupled utilities comes with sane abstractions. (`vp*.go`, `rules*.go`)
  - `VP` implementations for both concrete types and reflection types.

*zero-allocation can be achieved by adding proper buffering in `ParseOptions` and `CmdOptions`.

## Examples

see [package examples](https://pkg.go.dev/github.com/primecitizens/cli#pkg-examples)

## Known Limitations

1. Flags having implicit value cannot be followed by value prefixed with hyphen (`-`) if the flag accepts that value (e.g. `--IntSum -1`). To workaround, choose any of following methods:
   - (Method 1) use `=` to assign flag value explicitly (e.g. `--IntSum=-1`)
   - (Method 2) set `ParseOptions.HandleParseError` to handle error of type `*ErrAmbiguousArgs`.

2. Limitations to standalone dash (`--`)
   - Cannot be a flag value (e.g. given `--sep --`, then `--sep` gets nothing). To wrokaround, use `=` to assign dash value (e.g. `--sep=--`)
   - Can never be a positional arg. To workaround, check if `dashArgs == nil` is true, and if it is ture, then there is no standalone dash.
   - Due to limitations mentioned above, you cannot use a hyphen (`-`) as a flag shorthand.

## Core Concepts

- A `FlagFinder` implementation is capable of searching flags known to it by flag name or shorthand, so it represents a set of flags.
- `FlagIndexer` extends the ability of `FlagFinder` with iteration support (`FlagIter`).
  - `MapIndexer`: register flags just like old days. (see [Motivation](#motivation))
  - `FuncIndexer`: provide ad-hoc flag indexing logic.
  - `LevelIndexer`: build flag hierarchies.
  - `ReflectIndexer`: lazily produce flags on request.
  - `MultiIndexer`: collect indexers as one.
  - ... or implement your own `FlagIndexer/FlagFinder` for you own use cases.
- A root command is the receiver of a `(*Cmd).Exec(...)` function call.
- A `Route` is a collection of all `*Cmd`s from the root command to the target command.

## Terminology

Illustration without sub-commands:

```txt
                        dash
                          |
      posArg  flag name   |
         |        |       |
$ ./foo xxx -i --join bar -- other args
            |          |    [ all args after the dash are dashArgs]
            |          |
            |      flag value
            |
 flag shorthand, with implicit value
```

- `args`: all strings provided to a cmd.
  - For a root command in real world, it usually is `os.Args[1:]`
- `flags`: before the first dash, strings interpreted as flag names and flag values.
  - flag (long) name: a string consists of more than one unicode characters.
  - flag shorthand: a string consists of exactly one unicode character and not a hyphen (`-`).
- `subcmds` (sub-commands): before the first dash, ignore flag names and their values, consecutive args matching a serial of `Cmd.Pattern`.
  - In the above illustration, if there is a `Cmd` in root command's `Children` whose `.Pattern` is `xxx`, then the posArg `xxx` becomes subcmd `xxx`.
- `posArgs` (positional args): before the first dash, strings that are not flags and subcmds.
- `dashArgs`: all strings after the first dash.

## Motivation

Most existing command-line libraies forces you to register your flags to some central registries to get things going, these registrations happen at runtime and often involves dynamic memory allocation, even the often highly praised Rust crate `clap` works this way, but under the guise of procedural macros.

While we acknowledge the fact flag registration works fine in most cases, it is obvious to us that larger applications with a fair amount of sub-commands and flags often suffer from such method due to the central builder forcing the application to write a lot of boilerplate code just to work around its own workflow and restrictions, and we definitely want something more flexible.

Hint: You may obtain (almost) the same experience of traditional flag registration from this module by using `MapIndexer` and predefined flag types.

## The Bigger Picture - an alternative `std`

We are building a no-GC, zero-allocation, FFI-friendly, well-structured `std` module (inside [primecitizens/pcz](https://github.com/primecitizens/pcz)) for Golang, exposing all lower bits of the runtime, and compatible with the official go1.21+ toolchain (`go tool compile/link/asm`).

This `cli` module serves as a showcase how we redesign fundamental components in a standard library.

## Roadmap

- [ ] Completion support for `fish`
- [ ] Add more Helper interfaces in addition to `HelperTerminal`.
  - [ ] `HelperMarkdown`
  - [ ] `HelperMandoc`
  - [ ] `HelperYAML`
- [ ] Define a new interface (`Command`) as abstraction of `Cmd` to allow custom implementation.
- [ ] Add cli tool `cligen` to generate flag indexer implementations for struct types.

## LICENSE

```txt
Copyright 2023 The Prime Citizens

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
