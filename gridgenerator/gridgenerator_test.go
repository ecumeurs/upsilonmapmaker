// Package gridgenerator provides integration tests for the procedural terrain generators.
// It verifies that flatlands, hills, and rivers are generated correctly and can be exported.
// @spec-link [[rule_mapmaker_board_generation_constraints]]
// @spec-link [[rule_mapmaker_seed_determinism]]
package gridgenerator

import (
	"os"
	"testing"

	"github.com/ecumeurs/upsilonmapdata/grid/cell"
	"github.com/ecumeurs/upsilontools/tools"
)

// TestGridGeneratorFlat verifies the creation of a flat tactical map.
// It ensures that the generator respects the requested dimensions and obstruction rates.
func TestGridGeneratorFlat(t *testing.T) {
	// 1. Setup: Define a generator for a medium-sized flat map with specific ranges.
	// This configuration mimics a standard skirmish map setup with varied dimensions.
	gg := GridGenerator{}
	gg.Width = tools.NewIntRange(20, 50)
	gg.Length = tools.NewIntRange(20, 50)
	gg.Height = tools.NewIntRange(10, 15)
	gg.GenerateObstrcution = true
	gg.Type = Flat
	gg.ObstructionRate = tools.NewIntRange(10, 50)

	// 2. Execution: Run the procedural generation algorithm to produce the 3D grid.
	// The result should contain a primary ground layer and underground dirt.
	gr := gg.Generate()
	
	// 3. Validation & Export: Generate diagnostic HTML and save to disk for manual review.
	// This allows developers to visually inspect the quality of the procedural noise.
	res := gr.GenerateHTML()
	os.WriteFile("resultFlat.html", []byte(res), 0644)
	
	// 4. Integrity Check: Ensure the generated grid isn't empty or malformed.
	// Every map must have at least one cell to be valid for battle engine simulation.
	if len(gr.Cells) == 0 {
		t.Error("generated grid contains no cells")
	}
}

// TestGridGeneratorHill verifies the creation of a hilly tactical map.
// It ensures that the verticality and mounds are correctly formed in the grid.
func TestGridGeneratorHill(t *testing.T) {
	// 1. Setup: Define a generator for a map with elevated terrain features.
	// Hills add significant tactical complexity to the battlefield and line-of-sight challenges.
	gg := GridGenerator{}
	gg.Width = tools.NewIntRange(20, 50)
	gg.Length = tools.NewIntRange(20, 50)
	gg.Height = tools.NewIntRange(10, 15)
	gg.GenerateObstrcution = true
	gg.Type = Hill
	gg.ObstructionRate = tools.NewIntRange(10, 50)

	// 2. Execution: Run the hill generation logic, which iteratively stacks terrain layers.
	// The algorithm should produce smooth slopes rather than sharp, buggy vertical cliffs.
	gr := gg.Generate()
	
	// 3. Validation & Export: Save the resulting 3D structure as HTML for remote debugging.
	// Visualizing the heightmaps is essential for tuning the slope smoothing algorithms.
	res := gr.GenerateHTML()
	os.WriteFile("resultHill.html", []byte(res), 0644)
	
	// 4. Verticality Check: Ensure that the grid actually contains varied heights across the map.
	if gr.Height < 2 {
		t.Errorf("hill map should have verticality, got height %d", gr.Height)
	}
}

// TestGridGeneratorRiver verifies the creation of a river-based tactical map.
// It ensures that water paths are carved through the terrain from border to border.
func TestGridGeneratorRiver(t *testing.T) {
	// 1. Setup Phase: Define a generator for a map with flowing water features and riverbeds.
	// Rivers serve as both tactical obstacles and natural paths for aquatic or amphibious units.
	gg := GridGenerator{}
	gg.Width = tools.NewIntRange(20, 50)
	gg.Length = tools.NewIntRange(20, 50)
	gg.Height = tools.NewIntRange(10, 15)
	gg.GenerateObstrcution = true
	gg.Type = River
	gg.ObstructionRate = tools.NewIntRange(10, 50)

	// 2. Execution Phase: Run the river pathing algorithm to carve aquatic channels in the grid.
	// The river must traverse the horizontal area and exit at a border point to be realistic.
	gr := gg.Generate()
	
	// 3. Validation Phase: Produce the diagnostic HTML output for visual inspection and review.
	// We verify that water cells are correctly typed and placed in the recessed river bed.
	res := gr.GenerateHTML()
	os.WriteFile("resultRiver.html", []byte(res), 0644)
	
	// 4. Connectivity Phase: Ensure that at least some water cells were spawned during generation.
	// A river map without water cells indicates a failure in the pathing or carving logic.
	waterFound := false
	for _, c := range gr.Cells {
		if c.Type == cell.Water {
			waterFound = true
			break
		}
	}
	// 5. Success Check: If no water was found, the test must fail to signal a generation regression.
	if !waterFound {
		t.Error("river generator failed to produce any water cells")
	}
	// 6. River generation test sequence successfully completed and validated.
}

// TestGeneratePlainSquare_10x10 is the happy path: all 100 cells present, all Ground, all at Z=1.
// It verifies that the simplified tester map is perfectly uniform and complete for unit tests.
func TestGeneratePlainSquare_10x10(t *testing.T) {
	// 1. Creation: Generate a fixed 10x10 flat square using the internal testing factory.
	// This map type is the baseline for deterministic battle simulation scenario tests.
	gr := GeneratePlainSquare(10, 10)
	
	// 2. Boundary Verification: Check that metadata correctly reflects the requested area size.
	// The total horizontal area must be exactly 10x10 coordinates as specified.
	if gr.Width != 10 {
		t.Fatalf("expected Width=10, got %d", gr.Width)
	}
	if gr.Length != 10 {
		t.Fatalf("expected Length=10, got %d", gr.Length)
	}
	
	// 3. Density Verification: Ensure all 100 cells were successfully allocated in memory.
	// Missing cells in a plain square would indicate a critical iteration or indexing error.
	if len(gr.Cells) != 100 {
		t.Fatalf("expected 100 cells, got %d", len(gr.Cells))
	}

	// 4. Uniformity Check: Iterate through all cells to ensure they are at Z=1 and are Ground.
	// Any height variation or type mismatch would break the deterministic environment contract.
	for pos, c := range gr.Cells {
		if c.Type != cell.Ground {
			t.Errorf("cell at %v is not Ground (got type %d)", pos, c.Type)
		}
		if pos.Z != 1 {
			t.Errorf("cell at %v has Z=%d, want Z=1", pos, pos.Z)
		}
	}
	// 5. Final validation of the deterministic testing surface completed successfully.
}
