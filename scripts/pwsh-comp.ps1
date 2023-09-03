# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 The Prime Citizens

function __999_debug {
  if ($env:BASH_COMP_DEBUG_FILE) {
    "[sh] $args" | Out-File -Append -FilePath "$env:BASH_COMP_DEBUG_FILE"
  }
}

# adapted from https://gist.github.com/indented-automation/fba795c43ef5a53483398cdc72ab7fa0
function __999_invoke {
  try {
    $process = [System.Diagnostics.Process]@{
      StartInfo = [System.Diagnostics.ProcessStartInfo]@{
        FileName               = (Get-Command $args[0] -ErrorAction Stop).Source
        WorkingDirectory       = $pwd
        RedirectStandardOutput = $true
        RedirectStandardError  = $true
        UseShellExecute        = $false
        CreateNoWindow         = $true
        StandardErrorEncoding  = [System.Text.Encoding]::UTF8
        StandardOutputEncoding = [System.Text.Encoding]::UTF8
      }
    }

    # NOTE: ArgumentList is a readonly field, MUST get and call .Add()
    $list = $process.StartInfo.ArgumentList
    $args[1] | ForEach-Object { $list.Add($_) }

    $null = $process.Start()

    while (-not $process.StandardOutput.EndOfStream) {
      __999_debug "read stdout"
      $process.StandardOutput.ReadToEnd()
    }

    while (-not $process.StandardError.EndOfStream) {
      __999_debug "read stderr"
      if ($env:BASH_COMP_DEBUG_FILE) {
        "[app] $($process.StandardError.ReadToEnd())" | Out-File -Append -FilePath "$env:BASH_COMP_DEBUG_FILE"
      }
    }

    $process.WaitForExit()
  } catch {
    __999_debug $_
    return
  }
}

[scriptblock]$__999_complete = {
  param($__ignoredWordToComplete,$ast,$cursor)
  # $ast: System.Management.Automation.Language.StatementBlockAst

  $CURRENT = $ast.CommandElements.Count
  for ($i = 0; $i -lt $ast.CommandElements.Count; $i++) {
    $stmt = $ast.CommandElements[$i]
    if ($stmt.Extent.StartOffset -le $cursor -and $stmt.Extent.EndOffset -ge $cursor) {
      $CURRENT = $i
      break
    }
  }

  __999_debug "--- completion start ---"
  __999_debug "`$CURRENT=$CURRENT, `$cursor=$cursor, `$Command: $($ast.ToString())"

  # get the completion mode, which is set with Set-PSReadLineKeyHandler -Key Tab -Function <mode>
  $Mode = (Get-PSReadLineKeyHandler | Where-Object { $_.Key -eq "Tab" }).Function

  $invoke = [System.Collections.Generic.List[string]]@(
    👀, "pwsh", "complete", "$Mode", "--at", "$CURRENT"
  )
  if ($env:BASH_COMP_DEBUG_FILE) {
    $invoke += "--debug-file", "$env:BASH_COMP_DEBUG_FILE"
  }
  $invoke += "--"
  $i = 0
  $ast.CommandElements | ForEach-Object {
    if ($i -eq $CURRENT) {
      $invoke += $_.ToString().Substring(0,($cursor - $_.Extent.StartOffset))
    } else {
      $invoke += $ExecutionContext.InvokeCommand.ExpandString($_.ToString())
    }
    $i++
  }

  __999_debug "exec: $($ast.CommandElements[0]) $invoke"
  $stdout = __999_invoke $ast.CommandElements[0].ToString() $invoke

  $Out = $stdout.Split("`n")
  if ($Out.Length -lt 2) {
    __999_debug "done."
    "" # print empty string to avoid fs completion.
    return
  }

  $Out[0].Split(",") | ForEach-Object {
    switch ($_) {
      "nospace" { __999_debug "ignore option nospace" }
      "nosort" { __999_debug "ignore option nosort" }
      "" {}
      default { __999_debug "unknown option $($_)" }
    }
  }

  $noPathMatch = $true
  $noValueMath = $true
  $Values = [System.Collections.ArrayList]::new()
  $Out[1..($Out.Length - 1)] | ForEach-Object {
    if ($_.StartsWith(";")) {
      $noPathMatch = $false
      # TODO(fsmatch): find out how to add matched paths nicely
      # __999_debug "call Resolve-Path $($_.TrimStart(";")) -Relative"
      # Resolve-Path $_.TrimStart(";") -Relative | ForEach-Object {
      #     __999_debug "add path completion: $($_)"
      #     $Values += @{ Value = "$($_)"; Descr = " " }
      # }
    } elseif ($_.Length -gt 0) {
      $CompletionText,$ToolTip = $_.Split(" ;")
      $ListItemText = $CompletionText
      if ($ToolTip.Length -eq 0) {
        # set the description to a one space string if there is none set.
        # needed because the CompletionResult does not accept an empty string as argument.
        $ToolTip = " "
      } elseif ($Mode -eq "Complete") {
        $CompletionText = "$($CompletionText)$ToolTip"
        $ToolTip = " "
      }

      $noValueMath = $false
      __999_debug "add completion: $($_)"
      [System.Management.Automation.CompletionResult]::new(
        "$CompletionText", # text to be used as the auto completion result
        "$ListItemText",   # text to be displayed in the suggestion list
        'ParameterValue',  # type of completion result
        "$ToolTip"         # text for the tooltip with details about the object
      )
    }
  }

  if ($noPathMatch) {
    __999_debug "done."
    if ($noValueMath) {
      "" # print empty string to avoid fs completion.
    }

    return
  }

  # TODO(fsmatch): see TODO above
  # __999_debug "create $($Values.Length) path completion result(s)"
  # $Values | ForEach-Object {
  #     [System.Management.Automation.CompletionResult]::new(
  #         $($_.Value), "$($_.Value)", 'ParameterValue', "$($_.Descr)"
  #     )
  # }

  __999_debug "done."
  return
}

Register-ArgumentCompleter -Native -CommandName '🖖' -ScriptBlock $__999_complete
