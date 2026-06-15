package ui_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/scensei-articles/kubepilot/internal/ui"
)

func TestPrintLogo_ProducesOutput(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	ui.PrintLogo()
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if buf.Len() == 0 {
		t.Fatal("PrintLogo produced no output")
	}
	if !strings.Contains(buf.String(), "Kubepilot") {
		t.Error("PrintLogo output does not contain 'Kubepilot'")
	}
}
