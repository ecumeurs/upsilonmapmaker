package gridgenerator

import (
	"os"
	"testing"

	"github.com/ecumeurs/upsilontools/tools"
)

func TestGridGeneratorFlat(t *testing.T) {
	gg := GridGenerator{}
	gg.Width = tools.NewIntRange(20, 50)
	gg.Length = tools.NewIntRange(20, 50)
	gg.Height = tools.NewIntRange(10, 15)
	gg.GenerateObstrcution = true
	gg.Type = Flat
	gg.ObstructionRate = tools.NewIntRange(10, 50)

	gr := gg.Generate()
	res := gr.GenerateHTML()
	// store res to file result.svg
	os.WriteFile("resultFlat.html", []byte(res), 0644)
}

func TestGridGeneratorHill(t *testing.T) {
	gg := GridGenerator{}
	gg.Width = tools.NewIntRange(20, 50)
	gg.Length = tools.NewIntRange(20, 50)
	gg.Height = tools.NewIntRange(10, 15)
	gg.GenerateObstrcution = true
	gg.Type = Hill
	gg.ObstructionRate = tools.NewIntRange(10, 50)

	gr := gg.Generate()
	res := gr.GenerateHTML()
	// store res to file result.svg
	os.WriteFile("resultHill.html", []byte(res), 0644)
}

func TestGridGeneratorRiver(t *testing.T) {
	gg := GridGenerator{}
	gg.Width = tools.NewIntRange(20, 50)
	gg.Length = tools.NewIntRange(20, 50)
	gg.Height = tools.NewIntRange(10, 15)
	gg.GenerateObstrcution = true
	gg.Type = River
	gg.ObstructionRate = tools.NewIntRange(10, 50)

	gr := gg.Generate()
	res := gr.GenerateHTML()
	// store res to file result.svg
	os.WriteFile("resultRiver.html", []byte(res), 0644)
}

// TestGeneratePlainSquare_10x10 is the happy path: all 100 cells present, all Ground, all at Z=1.
func TestGeneratePlainSquare_10x10(t *testing.T) {
	gr := GeneratePlainSquare(10)
	if gr.Width != 10 {
		t.Fatalf("expected Width=10, got %d", gr.Width)
	}
	if gr.Length != 10 {
		t.Fatalf("expected Length=10, got %d", gr.Length)
	}
	if len(gr.Cells) != 100 {
		t.Fatalf("expected 100 cells, got %d", len(gr.Cells))
	}
	for pos, c := range gr.Cells {
		if c.Type != 1 { // cell.Ground == 1
			t.Errorf("cell at %v is not Ground (got type %d)", pos, c.Type)
		}
		if pos.Z != 1 {
			t.Errorf("cell at %v has Z=%d, want Z=1", pos, pos.Z)
		}
	}
}

// TestGeneratePlainSquare_SingleCell is an edge case: size=1 → exactly one cell.
func TestGeneratePlainSquare_SingleCell(t *testing.T) {
	gr := GeneratePlainSquare(1)
	if len(gr.Cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(gr.Cells))
	}
}

// TestGeneratePlainSquare_Zero is an edge case: size=0 → empty map, no panic.
func TestGeneratePlainSquare_Zero(t *testing.T) {
	gr := GeneratePlainSquare(0)
	if gr == nil {
		t.Fatal("expected non-nil grid for size=0")
	}
	if len(gr.Cells) != 0 {
		t.Fatalf("expected 0 cells, got %d", len(gr.Cells))
	}
}
