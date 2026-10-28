package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoggerHangsEveryValueUnderItsMessage(t *testing.T) {
	var buf bytes.Buffer
	run(newLogger(&buf, false), 0)
	out := buf.String()
	t.Log("\n" + out)
	require.Contains(t, out, "WARN  [installers] 1 file to move\n")
	require.Contains(t, out, "└─ name=go1.27.3.windows-amd64.msi\n")
	require.Contains(t, out, "TRACE scanning")
	require.Contains(t, out, "├─ skipped\n")
	require.NotContains(t, out, "\x1b[", "a buffer is no terminal, so the log is plain")
}

func TestLoggerPrintsJSONLines(t *testing.T) {
	var buf bytes.Buffer
	run(newLogger(&buf, true), 0)
	sc := bufio.NewScanner(&buf)
	var levels []string
	for sc.Scan() {
		var line map[string]any
		require.NoError(t, json.Unmarshal(sc.Bytes(), &line), sc.Text())
		levels = append(levels, line["level"].(string))
	}
	require.Equal(t, "INFO", levels[0])
	require.Equal(t, "TRACE", levels[1])
	require.Equal(t, "ERROR", levels[len(levels)-2])
}
