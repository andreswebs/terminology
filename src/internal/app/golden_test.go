package app_test

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/app"
	"github.com/andreswebs/terminology/internal/output"
)

var update = flag.Bool("update", false, "rewrite golden files")

func runGolden(t *testing.T, name string, argv []string) {
	t.Helper()
	runGoldenCtx(t, name, argv, context.Background())
}

func runGoldenCtx(t *testing.T, name string, argv []string, ctx context.Context) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	cmd := app.Root()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	err := cmd.Run(ctx, argv)

	exitCode := 0
	if err != nil {
		exitCode = output.ExitCodeFor(err)
		output.EmitError(&stderr, cmd.String("format"), err)
	}

	dir := filepath.Join("testdata", name)

	if *update {
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			t.Fatalf("mkdir %s: %v", dir, mkErr)
		}
		writeGolden(t, filepath.Join(dir, "clean.stdout"), stdout.Bytes())
		writeGolden(t, filepath.Join(dir, "clean.stderr"), stderr.Bytes())
		writeGolden(t, filepath.Join(dir, "clean.exit"), []byte(strconv.Itoa(exitCode)+"\n"))
		t.Logf("updated golden files in %s", dir)
		return
	}

	compareGolden(t, filepath.Join(dir, "clean.stdout"), stdout.Bytes())
	compareGolden(t, filepath.Join(dir, "clean.stderr"), stderr.Bytes())

	wantExitRaw, readErr := os.ReadFile(filepath.Join(dir, "clean.exit"))
	if readErr != nil {
		t.Fatalf("read golden exit: %v", readErr)
	}
	wantExit, _ := strconv.Atoi(strings.TrimSpace(string(wantExitRaw)))
	if exitCode != wantExit {
		t.Errorf("exit code = %d, want %d", exitCode, wantExit)
	}
}

func writeGolden(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write golden %s: %v", path, err)
	}
}

func compareGolden(t *testing.T, path string, got []byte) {
	t.Helper()

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}

	want = normGolden(want)
	got = normGolden(got)

	if !bytes.Equal(got, want) {
		t.Errorf("mismatch %s:\ngot:\n%s\nwant:\n%s", path, got, want)
	}
}

func normGolden(b []byte) []byte {
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	cwd, err := os.Getwd()
	if err == nil {
		b = bytes.ReplaceAll(b, []byte(cwd), []byte("$CWD"))
	}
	return b
}
