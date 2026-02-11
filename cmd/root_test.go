package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func executeWithArgs(args []string, stdin *bytes.Buffer) (string, error) {
	// Reset flags to defaults before each run.
	filePaths = nil
	dirPath = ""
	recursive = false
	checkPath = ""
	encoding = "base64"

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	if stdin != nil {
		rootCmd.SetIn(stdin)
	}
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestStdinBase64(t *testing.T) {
	input := bytes.NewBufferString("hello")
	out, err := executeWithArgs(nil, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parts := strings.SplitN(out, "  ", 2)
	if len(parts) != 2 {
		t.Fatalf("expected two-space separated output, got: %q", out)
	}
	if strings.TrimSpace(parts[1]) != "-" {
		t.Errorf("expected label '-', got %q", strings.TrimSpace(parts[1]))
	}
	if parts[0] == "" {
		t.Error("expected non-empty checksum")
	}
}

func TestFileBase64(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "testfile.txt")
	if err := os.WriteFile(fp, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := executeWithArgs([]string{"--file", fp}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parts := strings.SplitN(out, "  ", 2)
	if len(parts) != 2 {
		t.Fatalf("expected two-space separated output, got: %q", out)
	}
	if strings.TrimSpace(parts[1]) != fp {
		t.Errorf("expected label %q, got %q", fp, strings.TrimSpace(parts[1]))
	}
}

func TestHexEncoding(t *testing.T) {
	input := bytes.NewBufferString("hello")
	out, err := executeWithArgs([]string{"-e", "hex"}, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parts := strings.SplitN(out, "  ", 2)
	if len(parts) != 2 {
		t.Fatalf("expected two-space separated output, got: %q", out)
	}
	// Hex output should only contain hex characters.
	checksum := parts[0]
	for _, c := range checksum {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("unexpected character in hex output: %c", c)
		}
	}
}

func TestStdinAndFileProduceSameChecksum(t *testing.T) {
	content := "the quick brown fox"

	// Via stdin.
	stdinOut, err := executeWithArgs(nil, bytes.NewBufferString(content))
	if err != nil {
		t.Fatalf("stdin: unexpected error: %v", err)
	}
	stdinChecksum := strings.SplitN(stdinOut, "  ", 2)[0]

	// Via file.
	dir := t.TempDir()
	fp := filepath.Join(dir, "testfile.txt")
	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	fileOut, err := executeWithArgs([]string{"--file", fp}, nil)
	if err != nil {
		t.Fatalf("file: unexpected error: %v", err)
	}
	fileChecksum := strings.SplitN(fileOut, "  ", 2)[0]

	if stdinChecksum != fileChecksum {
		t.Errorf("checksums differ: stdin=%q file=%q", stdinChecksum, fileChecksum)
	}
}

func TestMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	fp1 := filepath.Join(dir, "a.txt")
	fp2 := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(fp1, []byte("alpha"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fp2, []byte("beta"), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := executeWithArgs([]string{"-f", fp1, "-f", fp2}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	// Verify each line has the correct label.
	parts1 := strings.SplitN(lines[0], "  ", 2)
	parts2 := strings.SplitN(lines[1], "  ", 2)
	if parts1[1] != fp1 {
		t.Errorf("line 1: expected label %q, got %q", fp1, parts1[1])
	}
	if parts2[1] != fp2 {
		t.Errorf("line 2: expected label %q, got %q", fp2, parts2[1])
	}
	// Different content should produce different checksums.
	if parts1[0] == parts2[0] {
		t.Error("expected different checksums for different content")
	}
}

func TestDirectory(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("alpha"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("beta"), 0644)

	out, err := executeWithArgs([]string{"-d", dir}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	// ReadDir returns entries sorted by name, so a.txt comes first.
	if !strings.HasSuffix(lines[0], "  "+filepath.Join(dir, "a.txt")) {
		t.Errorf("line 1: unexpected label: %q", lines[0])
	}
	if !strings.HasSuffix(lines[1], "  "+filepath.Join(dir, "b.txt")) {
		t.Errorf("line 2: unexpected label: %q", lines[1])
	}
}

func TestDirectorySkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)
	os.WriteFile(filepath.Join(dir, "subdir", "nested.txt"), []byte("nested"), 0644)

	out, err := executeWithArgs([]string{"-d", dir}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (subdirs skipped), got %d: %q", len(lines), out)
	}
}

func TestDirectoryAndFileCombined(t *testing.T) {
	dir := t.TempDir()
	dirFiles := filepath.Join(dir, "mydir")
	os.MkdirAll(dirFiles, 0755)
	os.WriteFile(filepath.Join(dirFiles, "a.txt"), []byte("alpha"), 0644)

	extra := filepath.Join(dir, "extra.txt")
	os.WriteFile(extra, []byte("extra"), 0644)

	out, err := executeWithArgs([]string{"-d", dirFiles, "-f", extra}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
}

func TestDirectoryAndFileDedup(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "a.txt")
	os.WriteFile(fp, []byte("alpha"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("beta"), 0644)

	// -f points to a file already in -d; should appear only once.
	out, err := executeWithArgs([]string{"-d", dir, "-f", fp}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (deduped), got %d: %q", len(lines), out)
	}
}

func TestFileDedup(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "a.txt")
	os.WriteFile(fp, []byte("alpha"), 0644)

	// Same file passed twice via -f.
	out, err := executeWithArgs([]string{"-f", fp, "-f", fp}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (deduped), got %d: %q", len(lines), out)
	}
}

func TestRecursiveDirectory(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "top.txt"), []byte("top"), 0644)
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "nested.txt"), []byte("nested"), 0644)

	out, err := executeWithArgs([]string{"-d", dir, "-r"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	// WalkDir visits in lexical order: sub/nested.txt then top.txt.
	if !strings.HasSuffix(lines[0], "  "+filepath.Join(dir, "sub", "nested.txt")) {
		t.Errorf("line 1: unexpected label: %q", lines[0])
	}
	if !strings.HasSuffix(lines[1], "  "+filepath.Join(dir, "top.txt")) {
		t.Errorf("line 2: unexpected label: %q", lines[1])
	}
}

func TestRecursiveWithoutDirectoryReturnsError(t *testing.T) {
	_, err := executeWithArgs([]string{"-r"}, nil)
	if err == nil {
		t.Fatal("expected error for --recursive without --directory, got nil")
	}
}

func TestCheckFromFile(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.txt")
	os.WriteFile(fp, []byte("hello"), 0644)

	// Generate checksums.
	sumOut, err := executeWithArgs([]string{"-f", fp}, nil)
	if err != nil {
		t.Fatalf("sum: unexpected error: %v", err)
	}

	// Write checksum output to a file.
	checksumFile := filepath.Join(dir, "checksums.txt")
	os.WriteFile(checksumFile, []byte(sumOut), 0644)

	// Verify with --check.
	checkOut, err := executeWithArgs([]string{"-c", checksumFile}, nil)
	if err != nil {
		t.Fatalf("check: unexpected error: %v", err)
	}
	if !strings.Contains(checkOut, ": OK") {
		t.Errorf("expected OK output, got %q", checkOut)
	}
}

func TestCheckFromStdin(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.txt")
	os.WriteFile(fp, []byte("hello"), 0644)

	// Generate checksums.
	sumOut, err := executeWithArgs([]string{"-f", fp}, nil)
	if err != nil {
		t.Fatalf("sum: unexpected error: %v", err)
	}

	// Verify via stdin with -c -.
	checkOut, err := executeWithArgs([]string{"-c", "-"}, bytes.NewBufferString(sumOut))
	if err != nil {
		t.Fatalf("check: unexpected error: %v", err)
	}
	if !strings.Contains(checkOut, ": OK") {
		t.Errorf("expected OK output, got %q", checkOut)
	}
}

func TestCheckDetectsMismatch(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.txt")
	os.WriteFile(fp, []byte("hello"), 0644)

	// Generate checksums.
	sumOut, err := executeWithArgs([]string{"-f", fp}, nil)
	if err != nil {
		t.Fatalf("sum: unexpected error: %v", err)
	}

	// Modify the file so the checksum no longer matches.
	os.WriteFile(fp, []byte("changed"), 0644)

	checksumFile := filepath.Join(dir, "checksums.txt")
	os.WriteFile(checksumFile, []byte(sumOut), 0644)

	checkOut, err := executeWithArgs([]string{"-c", checksumFile}, nil)
	if err == nil {
		t.Fatal("expected error for mismatched checksum, got nil")
	}
	if !strings.Contains(checkOut, ": FAILED") {
		t.Errorf("expected FAILED output, got %q", checkOut)
	}
}

func TestCheckWithHexEncoding(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.txt")
	os.WriteFile(fp, []byte("hello"), 0644)

	// Generate hex checksums.
	sumOut, err := executeWithArgs([]string{"-f", fp, "-e", "hex"}, nil)
	if err != nil {
		t.Fatalf("sum: unexpected error: %v", err)
	}

	checksumFile := filepath.Join(dir, "checksums.txt")
	os.WriteFile(checksumFile, []byte(sumOut), 0644)

	// Check with matching encoding.
	checkOut, err := executeWithArgs([]string{"-c", checksumFile, "-e", "hex"}, nil)
	if err != nil {
		t.Fatalf("check: unexpected error: %v", err)
	}
	if !strings.Contains(checkOut, ": OK") {
		t.Errorf("expected OK output, got %q", checkOut)
	}
}

func TestCheckMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	fp1 := filepath.Join(dir, "a.txt")
	fp2 := filepath.Join(dir, "b.txt")
	os.WriteFile(fp1, []byte("alpha"), 0644)
	os.WriteFile(fp2, []byte("beta"), 0644)

	sumOut, err := executeWithArgs([]string{"-f", fp1, "-f", fp2}, nil)
	if err != nil {
		t.Fatalf("sum: unexpected error: %v", err)
	}

	checksumFile := filepath.Join(dir, "checksums.txt")
	os.WriteFile(checksumFile, []byte(sumOut), 0644)

	checkOut, err := executeWithArgs([]string{"-c", checksumFile}, nil)
	if err != nil {
		t.Fatalf("check: unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(checkOut), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), checkOut)
	}
	for _, line := range lines {
		if !strings.HasSuffix(line, ": OK") {
			t.Errorf("expected OK, got %q", line)
		}
	}
}

func TestCheckIncompatibleWithFile(t *testing.T) {
	_, err := executeWithArgs([]string{"-c", "checksums.txt", "-f", "file.txt"}, nil)
	if err == nil {
		t.Fatal("expected error for --check with --file, got nil")
	}
}

func TestCheckIncompatibleWithDirectory(t *testing.T) {
	_, err := executeWithArgs([]string{"-c", "checksums.txt", "-d", "/tmp"}, nil)
	if err == nil {
		t.Fatal("expected error for --check with --directory, got nil")
	}
}

func TestNonexistentDirectoryReturnsError(t *testing.T) {
	_, err := executeWithArgs([]string{"-d", "/nonexistent/dir"}, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent directory, got nil")
	}
}

func TestNonexistentFileReturnsError(t *testing.T) {
	_, err := executeWithArgs([]string{"--file", "/nonexistent/path/file.txt"}, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
