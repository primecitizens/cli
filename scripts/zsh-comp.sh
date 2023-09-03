# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 The Prime Citizens

# shellcheck shell=bash

compdef _999 🖖

__999_debug() {
  [[ -n "$BASH_COMP_DEBUG_FILE" ]] && echo "[sh] $*" >>"$BASH_COMP_DEBUG_FILE"
}

_999() {
  __999_debug "--- completion start ---"
  # eval each word to expand environ as shell would
  for ((i = 1; i <= ${#words}; i++)); do words[i]="$(eval "echo \"${words[i]}\"")"; done
  __999_debug "\$CURRENT=$CURRENT, \$words[*]: ${words[*]}"

  local -a invoke
  invoke=("${words[1]}" 👀 "zsh" "complete" "--at" "$((CURRENT-1))")
  [[ -n "$BASH_COMP_DEBUG_FILE" ]] && invoke+=("--debug-file" "$BASH_COMP_DEBUG_FILE")
  invoke+=("--" "${words[@]}")

  __999_debug "exec: ${invoke[*]}"

  local visited_firstline ret
  local -a completions extra_flags
  while IFS=$'\n' read -r line; do
    if [[ -z "$visited_firstline" ]]; then
      visited_firstline="1"
      while IFS=$',' read -r opt; do
        case "$opt" in
        nospace)
          __999_debug "add option nospace"
          extra_flags+=("-S" "")
          ;;
        nosort)
          __999_debug "add option nosort"
          extra_flags+=("-V")
          ;;
        esac
      done < <(echo "$line")
    elif [[ -n "$line" ]]; then
      case "$line" in
      :*)
        __999_debug "call _arguments ${line:1} ${extra_flags[*]}"
        ret=1
        _arguments "${line:1}" "${extra_flags[@]}" && ret=0
        ;;
      *)
        __999_debug "add completion: ${line}"
        completions+=("$line")
        ;;
      esac
    fi
  done < <("${invoke[@]}" 2>/dev/null)

  _describe 'completions' completions "${extra_flags[@]}"
  __999_debug "done."
  return $ret
}

# shellcheck disable=SC2154
if [[ "${funcstack[1]}" == "_999" ]]; then
  _999
fi
