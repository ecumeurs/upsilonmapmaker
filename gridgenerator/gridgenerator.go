// Package gridgenerator provides algorithms for procedural map generation.
// It handles the creation of various terrain types including flat lands, hills, and rivers.
// @spec-link [[mapmaker_board_generation_constraints]]
// @spec-link [[mapmaker_seed_determinism]]
package gridgenerator

import (
	"fmt"

	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapdata/grid/cell"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilonmapdata/grid/position/pattern"
	"github.com/ecumeurs/upsilontools/tools"
)

// GridType defines the architectural style of the generated map.
// GridType defines the architectural style of the generated map.
// It determines which procedural generation algorithm is applied.
// @spec-link [[mapmaker_contract]]
// @spec-link [[mapmaker_vision]]
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

// fillBellowGroundWithDirt ensures that every ground cell has a column of dirt below it.
// This provides visual depth and ensures there are no floating tiles in the 3D view.
func fillBellowGroundWithDirt(g *grid.Grid) {
	// Iterate through all existing cells in the grid.
	for _, c := range g.Cells {
		// Only process non-dirt cells to find the top layer or obstacles.
		if c.Type != cell.Dirt {
			fillColumnBelow(g, c)
		}
	}
	// After filling, ensure the top-most layer is correctly identified as ground.
	ensureNoDirtIsVisibleOnTop(g)
}

// fillColumnBelow creates dirt cells below the given cell down to Z=0.
func fillColumnBelow(g *grid.Grid, c *cell.Cell) {
	// Iterate from the level immediately below the current cell down to the floor.
	for z := c.Position.Z - 1; z >= 0; z-- {
		pos := position.New(c.Position.X, c.Position.Y, z)
		// Maintain a limit of 5 levels of dirt to optimize cell count.
		if z > c.Position.Z-5 {
			if _, ok := g.Cells[pos]; !ok {
				// Create a new dirt cell if none exists at this position.
				d := &cell.Cell{
					Position: pos,
					Type:     cell.Dirt,
				}
				g.Cells[pos] = d
			} else {
				// Replace the existing cell type with dirt if it's already there.
				g.ReplaceCellType(pos, cell.Dirt)
			}
		} else {
			// Remove cells deep below the ground to save memory.
			delete(g.Cells, pos)
		}
	}
}

// ensureNoDirtIsVisibleOnTop converts any top-most dirt cells into ground.
// This prevents "dirt islands" from appearing on the surface of the map.
func ensureNoDirtIsVisibleOnTop(g *grid.Grid) {
	// Scan the entire width and length of the grid.
	for x := 0; x < g.Width; x++ {
		for y := 0; y < g.Length; y++ {
			// Identify the highest cell at the current (x, y) coordinates.
			z := g.TopMostCellAt(x, y)
			pos := position.New(x, y, z)
			if c, ok := g.Cells[pos]; ok && c.Type == cell.Dirt {
				// If the top cell is dirt, it should be ground.
				c.Type = cell.Ground
			}
		}
	}
}


// Generate creates a Grid based on the generator's configuration.
// It selects the appropriate generation algorithm based on the GridType.
func (g *GridGenerator) Generate() (res *grid.Grid) {
	// Select algorithm based on the configured terrain type.
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

	// Post-processing: fill the underground layers with dirt.
	fillBellowGroundWithDirt(res)

	return res
}

// generateFlat creates a mostly even terrain with minor random height variations.
func (g *GridGenerator) generateFlat() *grid.Grid {
	gr := new(grid.Grid)
	gr.Cells = make(map[position.Position]*cell.Cell)
	gr.Width = g.Width.Random()
	gr.Length = g.Length.Random()
	gr.Height = g.Height.Random()
	obstruction := g.ObstructionRate.Random()

	// Determine a base height for the ground.
	ground_height := tools.RandomInt(1, gr.Height-1)

	// Create the initial flat layer of ground cells.
	for x := 0; x < gr.Width; x++ {
		for y := 0; y < gr.Length; y++ {
			c := &cell.Cell{
				Position: position.New(x, y, ground_height),
				Type:     cell.Ground,
			}
			gr.Cells[c.Position] = c
		}
	}

	// Apply minor vertical noise to the terrain for a more natural look.
	for x := 0; x < gr.Width; x++ {
		for y := 0; y < gr.Length; y++ {
			applyHeightVariation(gr, x, y)
		}
	}

	// Randomly place obstacles if requested by the configuration.
	if g.GenerateObstrcution {
		for i := 0; i < obstruction; i++ {
			placeRandomObstacle(gr)
		}
	}

	return gr
}

// applyHeightVariation introduces a small chance of raising or lowering a specific tile.
// This creates more natural, varied terrain instead of a perfectly flat surface.
func applyHeightVariation(gr *grid.Grid, x, y int) {
	// Identify the current top-most ground level.
	z := gr.TopMostGroundAt(x, y)
	
	// Use a random roll to decide if we should modify this tile.
	rand := tools.RandomInt(0, 100)
	if rand > 90 {
		// 10% chance to raise the tile by one level.
		c := &cell.Cell{
			Position: position.New(x, y, z+1),
			Type:     cell.Ground,
		}
		gr.ReplaceCell(position.New(x, y, z), c)
	} else if rand > 88 {
		// 2% chance to lower the tile by one level.
		c := &cell.Cell{
			Position: position.New(x, y, z-1),
			Type:     cell.Ground,
		}
		gr.ReplaceCell(position.New(x, y, z), c)
	}
}

// placeRandomObstacle identifies a random ground tile and converts it to an obstacle.
// This is used to add tactical elements like rocks or trees to the generated map.
func placeRandomObstacle(gr *grid.Grid) {
	// Pick random coordinates within the grid boundaries.
	x := tools.RandomInt(0, gr.Width-1)
	y := tools.RandomInt(0, gr.Length-1)
	
	// Find the ground level at those coordinates.
	z := gr.TopMostGroundAt(x, y)
	
	// Convert the ground cell to an obstacle type.
	gr.ReplaceCellType(position.New(x, y, z), cell.Obstacle)
}



// adjascentHeights returns the Z-coordinates of all cells neighboring the given position.
// It queries the grid to find all valid cells in the immediate 3x3 horizontal vicinity.
func adjascentHeights(g *grid.Grid, p position.Position) []int {
	heights := []int{}
	
	// Iterate through all possible neighbor patterns (8 directions).
	for _, pt := range pattern.Neighbours() {
		neighborPos := p.Add(pt)
		
		// Skip if the neighbor is actually the same as the target position.
		if neighborPos.Equals(p) {
			continue
		}
		
		// If a cell exists at the neighbor position, record its vertical coordinate.
		if c, ok := g.Cells[neighborPos]; ok {
			heights = append(heights, c.Position.Z)
		}
	}
	return heights
}

// adjascentNearestHeight returns the lowest height among all neighbors of the given position.
// This is used during slope calculation to ensure the terrain isn't too steep.
func adjascentNearestHeight(g *grid.Grid, p position.Position, target_height int) int {
	// Gather all neighboring heights.
	heights := adjascentHeights(g, p)
	
	// If no neighbors exist, return the target height as a default.
	if len(heights) == 0 {
		return target_height
	}
	
	// Find the minimum height in the gathered list.
	minHeight := 999999
	for _, h := range heights {
		if h < minHeight {
			minHeight = h
		}
	}
	return minHeight
}



// generateHill creates a terrain with several elevated mounds.
// It starts with a flat base and then iteratively adds hills of random sizes and heights.
func (g *GridGenerator) generateHill() *grid.Grid {
	// Temporarily disable automatic obstruction generation to avoid conflicts during hill creation.
	ob := g.GenerateObstrcution
	g.GenerateObstrcution = false
	gr := g.generateFlat()
	g.GenerateObstrcution = ob

	// Add a random number of hills (between 1 and 4).
	for it := 0; it < tools.RandomInt(1, 5); it++ {
		// Select a random center point for the hill.
		x := tools.RandomInt(0, gr.Width)
		y := tools.RandomInt(0, gr.Length)
		z := gr.TopMostGroundAt(x, y)
		
		// Randomize the dimensions and peak elevation of the hill.
		hillsize := tools.RandomInt(8, 12)
		hillheight := z + tools.RandomInt(3, 8)

		// Ensure the hill doesn't exceed the grid's maximum vertical limit.
		if hillheight >= gr.Height {
			hillheight = gr.Height - 1
		}

		// Iterate through the affected area and update ground heights to form the slope.
		for x2 := x - hillsize; x2 < x+hillsize; x2++ {
			for y2 := y - hillsize; y2 < y+hillsize; y2++ {
				// Only process coordinates within the grid boundaries.
				if x2 >= 0 && x2 < gr.Width && y2 >= 0 && y2 < gr.Length {
					// Calculate distance from center to determine height falloff.
					dist := tools.FloatDistance(float64(x2), float64(y2), float64(x), float64(y))
					
					// Apply a linear falloff based on distance.
					height := hillheight + int(float64(hillheight)*(1.0-dist/float64(hillsize)))
					
					// Smooth the slope by checking adjacent heights to avoid cliffs.
					nearestHeight := adjascentNearestHeight(gr, position.New(x2, y2, height), height)
					if nearestHeight < height-2 {
						height = nearestHeight + 2
					}
					
					// Final vertical boundary check.
					if height >= gr.Height {
						height = gr.Height - 1
					}

					// Update the ground cell at the new calculated height.
					c := cell.Cell{
						Position: position.New(x2, y2, height),
						Type:     cell.Ground,
					}
					gr.ReplaceCell(position.New(x2, y2, gr.TopMostGroundAt(x2, y2)), &c)
				}
			}
		}
	}

	// Re-add obstructions if enabled, placing them on top of the newly formed hills.
	if g.GenerateObstrcution {
		obstruction := g.ObstructionRate.Random()
		for i := 0; i < obstruction; i++ {
			x := tools.RandomInt(0, gr.Width)
			y := tools.RandomInt(0, gr.Length)
			z := gr.TopMostGroundAt(x, y)
			gr.ReplaceCellType(position.New(x, y, z), cell.Obstacle)
		}
	}
	return gr
}


// GenerateRiver generates a river from a flat grid. A river is a path of water cells that are connected to the ground. Height of the river may be lower than the ground.
// A river must run through the whole grid
// There might be a branching in the river, but no more than 2.
// The river width may not be constant, but it is expected to be between 3 and 5 cells wide
// The river is expected to be flat in the middle and to have a slight slope on the sides

// generateRiver creates a water path cutting through the terrain.
// It selects origin and endpoints on the borders and connects them with waypoints.
func (g *GridGenerator) generateRiver() *grid.Grid {
	// Start with a flat base map.
	gr := g.generateFlat()

	// Select a random border position as the river's starting point.
	origin := position.RandomBorderPosition(gr.Width, gr.Length, gr.Height)
	origin, _ = gr.ForcePositionToGround(origin)
	
	// Determine one or more exit points on the grid borders.
	endpoints := []position.Position{}
	for i := 0; i < tools.RandomInt(1, 3); i++ {
		end := position.RandomBorderPosition(gr.Width, gr.Length, gr.Height)
		end, _ = gr.ForcePositionToGround(end)
		endpoints = append(endpoints, end)
	}

	// Randomize the width and depth (water level) of the river.
	width := tools.RandomInt(2, 4)
	waterlevel := tools.Min(tools.RandomInt(1, 3), tools.Max(0, gr.FindLowestLevel()-1))

	// Generate internal waypoints to create a meandering path.
	waypoints := []position.Position{}
	for i := 0; i < tools.RandomInt(1, 3); i++ {
		waypoint := position.RandomPosition(gr.Width, gr.Length, gr.Height)
		waypoint, _ = gr.ForcePositionToGround(waypoint)
		waypoints = append(waypoints, waypoint)
	}

	// Trace the river path from origin through all waypoints.
	currentRiverPosition := origin
	for _, waypoint := range waypoints {
		currentRiverPosition = g.generateRiverSegment(gr, currentRiverPosition, waypoint, width, waterlevel)
	}

	// Finally, connect the river to all selected endpoints.
	for _, endpoint := range endpoints {
		connectRiverToEndpoint(g, gr, endpoint, waterlevel)
	}

	// Add obstructions (like rocks or debris) if enabled, ensuring they don't block the water.
	if g.GenerateObstrcution {
		placeRiverObstructions(g, gr)
	}
	return gr
}

// connectRiverToEndpoint ensures the river flows out to the grid boundary.
func connectRiverToEndpoint(g *GridGenerator, gr *grid.Grid, endpoint position.Position, waterlevel int) {
	// Find the nearest existing water cell to the endpoint.
	nearestRiverCell, found := gr.FindNearestCellMatchingPredicate(endpoint, func(c *cell.Cell) bool {
		return c.Type == cell.Water
	})

	if found && nearestRiverCell.Position.Distance(endpoint) > 1 {
		// Generate a final segment to the border if a gap exists.
		g.generateRiverSegment(gr, nearestRiverCell.Position, endpoint, tools.RandomInt(2, 3), waterlevel)
	}
}

// placeRiverObstructions adds obstacles to the grid while avoiding water cells.
func placeRiverObstructions(g *GridGenerator, gr *grid.Grid) {
	obstruction := g.ObstructionRate.Random()
	for i := 0; i < obstruction; i++ {
		x := tools.RandomInt(0, gr.Width-1)
		y := tools.RandomInt(0, gr.Length-1)
		z := gr.TopMostGroundAt(x, y)
		
		// Avoid placing obstacles directly in the river.
		if c, found := gr.CellAt(position.New(x, y, z)); found && c.Type == cell.Water {
			i--
			continue
		}
		gr.ReplaceCellType(position.New(x, y, z), cell.Obstacle)
	}
}

// generateRiverSegment generates a river segment from a start position to an end position.
func (g *GridGenerator) generateRiverSegment(gr *grid.Grid, start position.Position, end position.Position, width int, waterlevel int) position.Position {
	// Find all positions along the path between start and end.
	// Enlarge the path based on the desired river width.
	for _, p := range pattern.PathTo2D(end.Substract(start)).Enlarge(width) {
		pos := position.New(start.X+p.X, start.Y+p.Y, waterlevel)
		
		// Remove any cells above the water level to create the river bed.
		for z := waterlevel; z < gr.Height; z++ {
			delete(gr.Cells, position.New(pos.X, pos.Y, z))
		}
		
		// Place a water cell at the designated level.
		c := &cell.Cell{
			Position: pos,
			Type:     cell.Water,
		}
		gr.Cells[c.Position] = c
	}
	return end
}


// GeneratePlainSquare returns a perfectly flat size×size grid.
// Every cell is of type Ground placed at Z=1. No height variation, no obstructions.
// Intended for small, deterministic testing maps (e.g. the 10×10 web-UI battle map).
func GeneratePlainSquare(w, h int) *grid.Grid {
	// Initialize a new grid structure.
	gr := new(grid.Grid)
	gr.Cells = make(map[position.Position]*cell.Cell)
	gr.Width = w
	gr.Length = h
	gr.Height = 2

	// Fill the grid area with ground cells at a fixed height of 1.
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c := &cell.Cell{
				Position: position.New(x, y, 1),
				Type:     cell.Ground,
			}
			gr.Cells[c.Position] = c
		}
	}

	return gr
}

