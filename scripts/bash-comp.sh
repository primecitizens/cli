# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 The Prime Citizens

# shellcheck shell=bash

__999_debug() {
  [[ -n ${BASH_COMP_DEBUG_FILE-} ]] && echo "[sh] $*" >>"$BASH_COMP_DEBUG_FILE"
}

__start_999() {
  local words cword
  COMPREPLY=()
  # use function from bash-completion package to prepare the arguments properly
  if declare -F _init_completion >/dev/null 2>&1; then
    _init_completion -n "=:" || return
  elif declare -F _comp_initialize >/dev/null 2>&1; then
    # since bash-completion 2.12 (unreleased as of 2023-07-05)
    _comp_initialize -n "=:" || return
  else
    # bash-completion for bash 3 doesn't include _init_completion.
    _get_comp_words_by_ref -n "=:" words cword || return
  fi

  __999_debug "--- completion start ---"
  # eval each word to expand environ as shell would
  for i in "${!words[@]}"; do words[i]="$(eval "echo \"${words[i]}\"")"; done
  __999_debug "\$cword=$cword, \$words[*]: ${words[*]}"

  local -a invoke
  invoke=("${words[0]}" 👀 "bash" "complete" "--at" "$cword" "${COLUMNS},${COMP_TYPE}")
  [[ -n "$BASH_COMP_DEBUG_FILE" ]] && invoke+=("--debug-file" "$BASH_COMP_DEBUG_FILE")
  invoke+=("--" "${words[@]}")

  __999_debug "exec: ${invoke[*]}"

  local visited_firstline
  while IFS=$'\n' read -r line; do
    if [[ -z "$visited_firstline" ]]; then
      visited_firstline="1"
      if [[ $(type -t compopt) != builtin ]]; then
        __999_debug "all options '$line' are not supported (bash ${BASH_VERSINFO[*]})"
        continue
      fi

      while IFS=$',' read -r opt; do
        case "$opt" in
        nospace)
          __999_debug "add option: nospace"
          compopt -o nospace
          ;;
        nosort)
          # requires bash >= 4.4
          if [[ ${BASH_VERSINFO[0]} -lt 4 || (${BASH_VERSINFO[0]} -eq 4 && ${BASH_VERSINFO[1]} -lt 4) ]]; then
            __999_debug "option '$opt' is not supported (bash ${BASH_VERSINFO[*]})"
          else
            __999_debug "add option: nosort"
            compopt -o nosort
          fi
          ;;
        esac
      done < <(echo "$line")
    elif [[ -n "$line" ]]; then
      case "$line" in
      ' '*)
        line="${line:1}"
        __999_debug "run: _filedir $line"
        local -a fsargs cmd=("_filedir")
        read -r -a fsargs < <(echo "$line")
        cmd+=("${fsargs[@]}")
        "${cmd[@]}"
        ;;
      *)
        __999_debug "add completion: $line"
        COMPREPLY+=("$line")
        ;;
      esac
    fi
  done < <("${invoke[@]}" 2>/dev/null)
  __999_debug "done."
}

complete -F __start_999 🖖
