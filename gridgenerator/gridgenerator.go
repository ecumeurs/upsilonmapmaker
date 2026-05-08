// Package gridgenerator provides algorithms for procedural map generation.
// It handles the creation of various terrain types including flat lands, hills, and rivers.
// @spec-link [[rule_mapmaker_board_generation_constraints]]
// @spec-link [[rule_mapmaker_seed_determinism]]
package gridgenerator

import (
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapdata/grid/cell"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilontools/tools"
)

// GridType defines the architectural style of the generated map.
// It determines which procedural generation algorithm is applied.
type GridType int

const (
	// Flat represents a mostly even terrain with minor height variations.
	Flat GridType = 0
	// Hill represents a terrain with several elevated mounds.
	Hill GridType = 1
	// River represents a terrain with water paths cutting through the ground.
	River GridType = 2
	// Mountain represents a terrain with significant verticality and peaks.
	Mountain GridType = 3
)

// GridGenerator holds the configuration and logic for creating a procedurally generated map.
type GridGenerator struct {
	// Width defines the range for the X-axis dimension.
	Width tools.IntRange
	// Length defines the range for the Y-axis dimension.
	Length tools.IntRange
	// Height defines the maximum vertical limit for the Z-axis.
	Height tools.IntRange

	// GenerateObstrcution determines if static obstacles should be spawned.
	GenerateObstrcution bool
	// ObstructionRate defines how many obstacles are placed on the map.
	ObstructionRate tools.IntRange

	// Type specifies the algorithm to use for terrain generation.
	Type GridType
}

// Generate creates a Grid based on the generator's configuration.
// It selects the appropriate generation algorithm based on the GridType.
func (g *GridGenerator) Generate() (res *grid.Grid) {
	// 1. Select algorithm based on the configured terrain type.
	switch g.Type {
	case Flat:
		res = g.generateFlat()
	case Hill:
		res = g.generateHill()
	case River:
		res = g.generateRiver()
	default:
		// Fallback to flat if unknown type.
		res = g.generateFlat()
	}

	// 2. Post-processing: fill the underground layers with dirt to ensure a solid volume.
	fillBellowGroundWithDirt(res)

	return res
}

// fillBellowGroundWithDirt ensures that every ground cell has a column of dirt below it.
// This provides visual depth and ensures there are no floating tiles in the 3D view.
func fillBellowGroundWithDirt(g *grid.Grid) {
	// 1. Iterate through all existing cells in the grid to find surface tiles.
	for _, c := range g.Cells {
		// 2. Only process non-dirt cells to find the top layer or obstacles.
		if c.Type != cell.Dirt {
			// 3. Fill the vertical column below the identified surface cell.
			fillColumnBelow(g, c)
		}
	}
	// 4. After filling, ensure the top-most layer is correctly identified as ground.
	ensureNoDirtIsVisibleOnTop(g)
}

// fillColumnBelow creates dirt cells below the given cell down to Z=0.
// It maintains a finite column height to optimize grid size and memory.
func fillColumnBelow(g *grid.Grid, c *cell.Cell) {
	// 1. Iterate from the level immediately below the current cell down to the floor.
	for z := c.Position.Z - 1; z >= 0; z-- {
		// 2. Delegate the physical cell creation or removal to a depth-aware helper.
		processVerticalDirtCell(g, c.Position.X, c.Position.Y, z, c.Position.Z)
	}
}

// processVerticalDirtCell handles the creation or cleaning of cells at a specific depth.
// It ensures that only a reasonable depth of dirt is maintained for performance.
func processVerticalDirtCell(g *grid.Grid, x, y, z, topZ int) {
	pos := position.New(x, y, z)
	// 1. Maintain a limit of 5 levels of dirt to optimize cell count.
	if z > topZ-5 {
		// 2. Add or update the cell to be of type Dirt in the persistence registry.
		ensureDirtCellAt(g, pos)
	} else {
		// 3. Remove cells deep below the ground to save memory and reduce vertex count.
		delete(g.Cells, pos)
	}
}

// ensureDirtCellAt is a helper to ensure a Dirt cell exists at the given position.
// It handles both new allocations and re-typing of existing cell objects.
func ensureDirtCellAt(g *grid.Grid, pos position.Position) {
	// 1. Check if a cell already exists at this exact 3D coordinate.
	if _, ok := g.Cells[pos]; !ok {
		// 2. Create a new dirt cell if the space is currently empty.
		g.Cells[pos] = &cell.Cell{
			Position: pos,
			Type:     cell.Dirt,
		}
	} else {
		// 3. Replace the existing cell type with dirt if a cell is already registered.
		g.ReplaceCellType(pos, cell.Dirt)
	}
}

// ensureNoDirtIsVisibleOnTop converts any top-most dirt cells into ground.
// This prevents "dirt islands" from appearing on the surface of the map.
func ensureNoDirtIsVisibleOnTop(g *grid.Grid) {
	// 1. Scan every horizontal grid location (X, Y) across the map area.
	for x := 0; x < g.Width; x++ {
		// 2. Delegate the vertical scan for each column to a specialized helper.
		ensureNoDirtVisibleInColumn(g, x)
	}
}

// ensureNoDirtVisibleInColumn is a helper to process a single column, reducing nesting.
// It identifies the surface tile and ensures it has a playable terrain type.
func ensureNoDirtVisibleInColumn(g *grid.Grid, x int) {
	// 1. Iterate through the full length of the grid for the given X coordinate.
	for y := 0; y < g.Length; y++ {
		// 2. Identify the highest cell at the current (x, y) coordinate.
		z := g.TopMostCellAt(x, y)
		pos := position.New(x, y, z)
		// 3. If the top cell is dirt, it must be converted to ground for gameplay and visuals.
		if c, ok := g.Cells[pos]; ok && c.Type == cell.Dirt {
			c.Type = cell.Ground
		}
	}
}

// GeneratePlainSquare returns a perfectly flat size×size grid for testing purposes.
// Every cell is of type Ground placed at Z=1. No height variation, no obstructions.
// This is used for deterministic scenario verification and UI prototyping.
func GeneratePlainSquare(w, h int) *grid.Grid {
	// 1. Initialization: Create a new grid structure with fixed dimensions.
	// 2. The height is set to 2 to allow for a ground layer and a dirt layer.
	gr := new(grid.Grid)
	gr.Cells = make(map[position.Position]*cell.Cell)
	gr.Width = w
	gr.Length = h
	gr.Height = 2

	// 3. Generation Loop: Fill the grid area with ground cells at height 1.
	// 4. This produces a perfectly uniform tactical surface for unit testing.
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c := &cell.Cell{
				Position: position.New(x, y, 1),
				Type:     cell.Ground,
			}
			gr.Cells[c.Position] = c
		}
	}
	// 5. Return the deterministic testing grid for immediate use in test suites.
	return gr
}
