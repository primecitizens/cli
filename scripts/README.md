# scripts

Scripts for shell completion

## Placeholders

- `🖖` is the program name, usually the root `Cmd.Name()`.
- `999` is the identifier version of `🖖`.
  - Created by replacing every non-alphanumeric character to `_`.
- `👀` is the name of the Cmd to handle completion request.

__NOTE__: The placeholder `999` is chosen because it doesn't appear in any of existing shell completion scripts and makes syntax highlighting work for development.

## Maintenance Checklist

- [ ] Use string array of arguments for command execution.
  - [ ] for bash, zsh: use `"${arr[@]}"`, NOTE: be sure it's quoted.
  - [ ] for pwsh: use `[System.Collections.Generic.List[string]]@(...)`.
- [ ] Include the executable path as the first arg in dashArgs to `complete` sub-command.
- [ ] Normalize value to `--at` flag to 0-based index of the argument to complete.
- [ ] Format scripts with tools.
  - [ ] for `*.sh`: [run `shfmt --indent=2 <script-file>`](#appendixtools)
  - [ ] for `*.ps1`: [import module `PowerShell-Beautifier`](#appendixtools), then run `Edit-DTWBeautifyScript -IndentType TwoSpaces -StandardOutput -NewLine LF -SourcePath <script-file>` and align comments manually.
- [ ] Format usage text (`*-usage.txt`), ensure line width <= 67.
- [ ] Ensure no unintentional `999` usage (in link/hash values).
- [ ] Ensure `../compsh.go` is up to date when changing placeholders.

## References

- `bash`
  - [https://www.gnu.org/savannah-checkouts/gnu/bash/manual/bash.html#Programmable-Completion](https://www.gnu.org/savannah-checkouts/gnu/bash/manual/bash.html#Programmable-Completion)
- `zsh`
  - [https://zsh.sourceforge.io/Doc/Release/Completion-System.html](https://zsh.sourceforge.io/Doc/Release/Completion-System.html)
  - [https://github.com/zsh-users/zsh-completions/blob/master/zsh-completions-howto.org](https://github.com/zsh-users/zsh-completions/blob/master/zsh-completions-howto.org)

## Appendix.Tools

- `shfmt`: [https://github.com/mvdan/sh](https://github.com/mvdan/sh)
- `PowerShell-Beautifier`: [https://github.com/DTW-DanWard/PowerShell-Beautifier](https://github.com/DTW-DanWard/PowerShell-Beautifier)
  - Import:

    ```powershell
    git clone https://github.com/DTW-DanWard/PowerShell-Beautifier
    Import-Module ./PowerShell-Beautifier/PowerShell-Beautifier.psd1
    ```
