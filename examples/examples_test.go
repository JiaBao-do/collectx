package examples_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestExamples runs every example program and compares stdout with its
// expected_output.txt, so the examples cannot silently rot.
func TestExamples(t *testing.T) {
	files, err := filepath.Glob("*/expected_output.txt")
	if err != nil || len(files) == 0 {
		t.Fatalf("no examples found: %v", err)
	}
	for _, f := range files {
		dir := filepath.Dir(f)
		t.Run(dir, func(t *testing.T) {
			want, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			out, err := exec.Command("go", "run", "./"+dir).CombinedOutput()
			if err != nil {
				t.Fatalf("go run failed: %v\n%s", err, out)
			}
			norm := func(b []byte) string { return strings.ReplaceAll(strings.TrimSpace(string(b)), "\r\n", "\n") }
			if norm(out) != norm(want) {
				t.Fatalf("output mismatch\n--- got ---\n%s\n--- want ---\n%s", out, want)
			}
		})
	}
}

// TestReadmeQuickstartInSync keeps the README quick start identical to the example.
func TestReadmeQuickstartInSync(t *testing.T) {
	src, err := os.ReadFile("quickstart/main.go")
	if err != nil {
		t.Fatal(err)
	}
	readme, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatal(err)
	}
	// The README block omits the two header comment lines.
	body := string(src)
	body = body[strings.Index(body, "package main"):]
	if !strings.Contains(strings.ReplaceAll(string(readme), "\r\n", "\n"), body) {
		t.Fatal("README quick start differs from examples/quickstart/main.go")
	}
}
