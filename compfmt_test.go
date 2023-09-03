// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"bytes"
	"testing"

	"github.com/primecitizens/cli/internal/assert"
)

func TestBashCompFmt(t *testing.T) {
	t.Run("WithDescription", func(t *testing.T) {
		fmt := CompFmtBash{
			Cols:     80,
			CompType: '\t',
		}
		testCompFormatter(t, &fmt, CompFmtTestSpec{
			noFsMatch: "" +
				"spaced\\ cmd                                                a somewhat long de...\n" +
				"command-with-a-long-name-to-test-completion-description    short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n",
			dirMatch: "" +
				"spaced\\ cmd                                                a somewhat long de...\n" +
				"command-with-a-long-name-to-test-completion-description    short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				"\x20-d\n",
			fileMatch: "" +
				"spaced\\ cmd                                                a somewhat long de...\n" +
				"command-with-a-long-name-to-test-completion-description    short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				"\x20\x20'file-pattern'\n",
			fileMatch2: "" +
				"spaced\\ cmd                                                a somewhat long de...\n" +
				"command-with-a-long-name-to-test-completion-description    short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				"\x20\x20'file-pattern|ptn2'\n",
		})
	})

	t.Run("NoDescription", func(t *testing.T) {
		fmt := CompFmtBash{
			Cols:     80,
			CompType: '%',
		}
		testCompFormatter(t, &fmt, CompFmtTestSpec{
			noFsMatch: "" +
				"spaced\\ cmd\n" +
				"command-with-a-long-name-to-test-completion-description\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n",
			dirMatch: "" +
				"spaced\\ cmd\n" +
				"command-with-a-long-name-to-test-completion-description\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				"\x20-d\n",
			fileMatch: "" +
				"spaced\\ cmd\n" +
				"command-with-a-long-name-to-test-completion-description\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				"\x20\x20'file-pattern'\n",
			fileMatch2: "" +
				"spaced\\ cmd\n" +
				"command-with-a-long-name-to-test-completion-description\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				"\x20\x20'file-pattern|ptn2'\n",
		})
	})
}

func TestZshCompFmt(t *testing.T) {
	fmt := CompFmtZsh{}

	testCompFormatter(t, fmt, CompFmtTestSpec{
		noFsMatch: "" +
			"spaced cmd:a somewhat long description with other long command.\n" +
			"command-with-a-long-name-to-test-completion-description:short\n" +
			"--flag-name\n" +
			"-s\n" +
			"colon\\:sep\\:value\n" +
			"flag-value\n",
		dirMatch: "" +
			"spaced cmd:a somewhat long description with other long command.\n" +
			"command-with-a-long-name-to-test-completion-description:short\n" +
			"--flag-name\n" +
			"-s\n" +
			"colon\\:sep\\:value\n" +
			"flag-value\n" +
			":*:dirname:_files -/\n",
		fileMatch: "" +
			"spaced cmd:a somewhat long description with other long command.\n" +
			"command-with-a-long-name-to-test-completion-description:short\n" +
			"--flag-name\n" +
			"-s\n" +
			"colon\\:sep\\:value\n" +
			"flag-value\n" +
			":*:filename:_files -g (file-pattern)\n",
		fileMatch2: "" +
			"spaced cmd:a somewhat long description with other long command.\n" +
			"command-with-a-long-name-to-test-completion-description:short\n" +
			"--flag-name\n" +
			"-s\n" +
			"colon\\:sep\\:value\n" +
			"flag-value\n" +
			":*:filename:_files -g (file-pattern|ptn2)\n",
	})
}

func TestPwshCompFmt(t *testing.T) {
	t.Run("ZshStyle", func(t *testing.T) {
		fmt := CompFmtPwsh{
			Mode: "MenuComplete",
		}

		testCompFormatter(t, &fmt, CompFmtTestSpec{
			noFsMatch: "" +
				"spaced` cmd ;a somewhat long description with other long command.\n" +
				"command-with-a-long-name-to-test-completion-description ;short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n",
			dirMatch: "" +
				"spaced` cmd ;a somewhat long description with other long command.\n" +
				"command-with-a-long-name-to-test-completion-description ;short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				";'*'\n",
			fileMatch: "" +
				"spaced` cmd ;a somewhat long description with other long command.\n" +
				"command-with-a-long-name-to-test-completion-description ;short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				";'(file-pattern)'\n",
			fileMatch2: "" +
				"spaced` cmd ;a somewhat long description with other long command.\n" +
				"command-with-a-long-name-to-test-completion-description ;short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				";'(file-pattern|ptn2)'\n",
		})
	})

	t.Run("BashStyle", func(t *testing.T) {
		fmt := CompFmtPwsh{
			Mode: "Complete",
		}

		testCompFormatter(t, &fmt, CompFmtTestSpec{
			noFsMatch: "" +
				"spaced` cmd ;                                                 a somewhat long description with other long command.\n" +
				"command-with-a-long-name-to-test-completion-description ;    short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n",
			dirMatch: "" +
				"spaced` cmd ;                                                 a somewhat long description with other long command.\n" +
				"command-with-a-long-name-to-test-completion-description ;    short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				";'*'\n",
			fileMatch: "" +
				"spaced` cmd ;                                                 a somewhat long description with other long command.\n" +
				"command-with-a-long-name-to-test-completion-description ;    short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				";'(file-pattern)'\n",
			fileMatch2: "" +
				"spaced` cmd ;                                                 a somewhat long description with other long command.\n" +
				"command-with-a-long-name-to-test-completion-description ;    short\n" +
				"--flag-name\n" +
				"-s\n" +
				"colon:sep:value\n" +
				"flag-value\n" +
				";'(file-pattern|ptn2)'\n",
		})
	})
}

type CompFmtTestSpec struct {
	noFsMatch  string
	dirMatch   string
	fileMatch  string
	fileMatch2 string
}

func testCompFormatter(t *testing.T, fmt CompFmt, spec CompFmtTestSpec) {
	items := []CompItem{
		{Value: "spaced cmd",
			Description: "a somewhat long description with other long command."},
		{Value: "command-with-a-long-name-to-test-completion-description",
			Description: "short\nbut\nmulti\nline"},
		{Value: "flag-name", Kind: CompKindFlagName},
		{Value: "s", Kind: CompKindFlagName},
		{Value: "colon:sep:value", Kind: CompKindText},
		{Value: "flag-value", Kind: CompKindFlagValue},
		{Value: "", Kind: CompKindDirs},
		{Value: "file-pattern", Kind: CompKindFiles},
		{Value: "ptn2", Kind: CompKindFiles},
	}

	var buf bytes.Buffer
	assert.NoError(t, fmt.Format(&buf, &CompTask{result: items[:len(items)-3]}))
	assert.Eq(t, spec.noFsMatch, buf.String())

	buf.Reset()
	assert.NoError(t, fmt.Format(&buf, &CompTask{result: items[:len(items)-2]}))
	assert.Eq(t, spec.dirMatch, buf.String())

	buf.Reset()
	assert.NoError(t, fmt.Format(&buf, &CompTask{result: items[:len(items)-1]}))
	assert.Eq(t, spec.fileMatch, buf.String())

	buf.Reset()
	assert.NoError(t, fmt.Format(&buf, &CompTask{result: items}))
	assert.Eq(t, spec.fileMatch2, buf.String())
}
