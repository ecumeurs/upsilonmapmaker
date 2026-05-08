// Package gridgenerator provides specialized terrain generation algorithms.
// This file contains the implementation for flat, hill, and river terrain generation.
// @spec-link [[rule_mapmaker_board_generation_constraints]]
// @spec-link [[rule_mapmaker_seed_determinism]]
package gridgenerator

import (
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapdata/grid/cell"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilonmapdata/grid/position/pattern"
	"github.com/ecumeurs/upsilontools/tools"
)

// generateFlat creates a mostly even terrain with minor random height variations.
// It populates the grid with a base layer of ground cells and applies vertical noise.
func (g *GridGenerator) generateFlat() *grid.Grid {
	// 1. Initialization: Setup the grid container and randomize base dimensions.
	// We use the configured ranges to ensure variety across different matches.
	gr := new(grid.Grid)
	gr.Cells = make(map[position.Position]*cell.Cell)
	gr.Width = g.Width.Random()
	gr.Length = g.Length.Random()
	gr.Height = g.Height.Random()
	obstruction := g.ObstructionRate.Random()

	// 2. Base Height: Determine a uniform starting height for the ground layer.
	// This leaves vertical space for both underground dirt and potential mountains.
	ground_height := tools.RandomInt(1, gr.Height-1)

	// 3. Ground Layer Creation: Fill the map area with a baseline layer of ground.
	// This provides a solid playable surface for unit movement and placement.
	for x := 0; x < gr.Width; x++ {
		for y := 0; y < gr.Length; y++ {
			c := &cell.Cell{
				Position: position.New(x, y, ground_height),
				Type:     cell.Ground,
			}
			gr.Cells[c.Position] = c
		}
	}

	// 4. Noise Application: Introduce minor vertical variation for natural appearance.
	// This breaks the "checkerboard" look of perfectly flat procedurally generated maps.
	for x := 0; x < gr.Width; x++ {
		for y := 0; y < gr.Length; y++ {
			applyHeightVariation(gr, x, y)
		}
	}

	// 5. Obstruction Placement: Randomly scatter obstacles if enabled in configuration.
	// Obstacles serve as cover and line-of-sight blockers for tactical gameplay.
	if g.GenerateObstrcution {
		for i := 0; i < obstruction; i++ {
			placeRandomObstacle(gr)
		}
	}
	return gr
}

// applyHeightVariation introduces a small chance of raising or lowering a specific tile.
// This creates more natural, varied terrain instead of a perfectly flat surface.
// It modifies the Z-coordinate of the topmost ground cell at the given location.
func applyHeightVariation(gr *grid.Grid, x, y int) {
	// 1. Identify the current top-most ground level for this coordinate.
	z := gr.TopMostGroundAt(x, y)
	
	// 2. Use a random roll (0-99) to decide if we should modify this tile.
	// A higher roll threshold ensures that variations are sparse and localized.
	rand := tools.RandomInt(0, 100)
	if rand > 90 {
		// 3. 10% chance to raise the tile by one level to create minor hills.
		// This simulates natural bumps in the terrain surface.
		c := &cell.Cell{
			Position: position.New(x, y, z+1),
			Type:     cell.Ground,
		}
		gr.ReplaceCell(position.New(x, y, z), c)
	} else if rand > 88 {
		// 4. 2% chance to lower the tile by one level to create shallow pits.
		// This adds more organic depth to the generated battlefield.
		c := &cell.Cell{
			Position: position.New(x, y, z-1),
			Type:     cell.Ground,
		}
		gr.ReplaceCell(position.New(x, y, z), c)
	}
	// 5. Variation application complete for this specific (x,y) column.
}

// placeRandomObstacle identifies a random ground tile and converts it to an obstacle.
// This is used to add tactical elements like rocks or trees to the generated map.
// It ensures that players have cover and line-of-sight blockers in standard matches.
func placeRandomObstacle(gr *grid.Grid) {
	// 1. Pick random coordinates within the horizontal grid boundaries.
	x := tools.RandomInt(0, gr.Width-1)
	y := tools.RandomInt(0, gr.Length-1)
	
	// 2. Find the topmost ground level at those coordinates.
	z := gr.TopMostGroundAt(x, y)
	
	// 3. Convert the ground cell to an obstacle type to block movement and LOS.
	gr.ReplaceCellType(position.New(x, y, z), cell.Obstacle)
}

// adjascentHeights returns the Z-coordinates of all cells neighboring the given position.
// It queries the grid to find all valid cells in the immediate 3x3 horizontal vicinity.
// This helper is essential for ensuring terrain connectivity and smooth slopes.
func adjascentHeights(g *grid.Grid, p position.Position) []int {
	heights := []int{}
	
	// 1. Iterate through all possible neighbor patterns in 3D space using the Pattern utility.
	for _, pt := range pattern.Neighbours() {
		// 2. Calculate the absolute world position of the neighbor cell.
		neighborPos := p.Add(pt)
		
		// 3. Skip the center point to avoid self-comparison in adjacency checks.
		if neighborPos.Equals(p) {
			continue
		}
		
		// 4. If a cell exists at the neighbor position, record its vertical coordinate.
		if c, ok := g.Cells[neighborPos]; ok {
			heights = append(heights, c.Position.Z)
		}
	}
	// 5. Return the collected list of neighbor heights for further analysis.
	return heights
}

// adjascentNearestHeight returns the lowest height among all neighbors of the given position.
// This is used during slope calculation to ensure the terrain isn't too steep for units.
// It provides a critical boundary for smoothing algorithms during hill generation.
func adjascentNearestHeight(g *grid.Grid, p position.Position, target_height int) int {
	// 1. Gather all neighboring heights from the 3D grid using the adjacency helper.
	heights := adjascentHeights(g, p)
	
	// 2. If no neighbors exist, return the target height as a default safety reference.
	if len(heights) == 0 {
		return target_height
	}
	
	// 3. Find the minimum height (valley point) in the gathered list of neighbors.
	minHeight := 999999
	for _, h := range heights {
		if h < minHeight {
			minHeight = h
		}
	}
	// 4. Return the found minimum height to the caller for slope verification.
	return minHeight
}

// generateHill creates a terrain with several elevated mounds.
// It starts with a flat base and then iteratively adds hills of random sizes and heights.
func (g *GridGenerator) generateHill() *grid.Grid {
	// 1. Base Setup: Temporarily disable obstructions to generate a clean flat base.
	ob := g.GenerateObstrcution
	g.GenerateObstrcution = false
	gr := g.generateFlat()
	g.GenerateObstrcution = ob

	// 2. Hill Creation Loop: Add a random number of hills across the map area.
	// Each hill is defined by a center point and a falloff radius.
	for it := 0; it < tools.RandomInt(1, 5); it++ {
		// Select a random center point and randomize hill dimensions.
		x := tools.RandomInt(0, gr.Width)
		y := tools.RandomInt(0, gr.Length)
		z := gr.TopMostGroundAt(x, y)
		hillsize := tools.RandomInt(8, 12)
		hillheight := tools.Min(z+tools.RandomInt(3, 8), gr.Height-1)

		// 3. Area Processing: Update tiles within the hill's radius.
		for x2 := x - hillsize; x2 < x+hillsize; x2++ {
			for y2 := y - hillsize; y2 < y+hillsize; y2++ {
				if x2 >= 0 && x2 < gr.Width && y2 >= 0 && y2 < gr.Length {
					applyHillSlope(gr, x, y, x2, y2, hillsize, hillheight)
				}
			}
		}
	}

	// 4. Finalizing: Re-add obstructions on top of the newly formed hills.
	if g.GenerateObstrcution {
		placeFinalObstructions(gr, g.ObstructionRate.Random())
	}
	return gr
}

// applyHillSlope calculates and applies the height of a single tile within a hill's radius.
// It uses a linear falloff and smoothing to produce realistic terrain slopes.
func applyHillSlope(gr *grid.Grid, centerX, centerY, x, y, size, peakHeight int) {
	// 1. Calculate distance from center to determine height falloff.
	dist := tools.FloatDistance(float64(x), float64(y), float64(centerX), float64(centerY))
	
	// 2. Apply a linear falloff based on distance from the peak point.
	// Tiles closer to the center will be significantly higher.
	height := peakHeight + int(float64(peakHeight)*(1.0-dist/float64(size)))
	
	// 3. Smooth the slope by checking adjacent heights to avoid impossible vertical cliffs.
	// We limit the maximum vertical drop between adjacent cells.
	nearestHeight := adjascentNearestHeight(gr, position.New(x, y, height), height)
	if nearestHeight < height-2 {
		height = nearestHeight + 2
	}
	
	// 4. Apply vertical bounds and update the grid cell to the new elevation.
	height = tools.Min(height, gr.Height-1)
	c := cell.Cell{
		Position: position.New(x, y, height),
		Type:     cell.Ground,
	}
	gr.ReplaceCell(position.New(x, y, gr.TopMostGroundAt(x, y)), &c)
}

// placeFinalObstructions is a helper to scatter obstacles after terrain is formed.
// This ensures that rocks and trees are placed on the surface of hills.
func placeFinalObstructions(gr *grid.Grid, count int) {
	// 1. Iterate based on the requested obstruction count for the match.
	for i := 0; i < count; i++ {
		// 2. Select random coordinates and find the current ground level.
		x := tools.RandomInt(0, gr.Width)
		y := tools.RandomInt(0, gr.Length)
		z := gr.TopMostGroundAt(x, y)
		// 3. Mark the target cell as an obstacle to block movement.
		gr.ReplaceCellType(position.New(x, y, z), cell.Obstacle)
	}
}

// generateRiver creates a water path cutting through the terrain.
// It selects origin and endpoints on the borders and connects them with waypoints.
func (g *GridGenerator) generateRiver() *grid.Grid {
	// 1. Setup Phase: Start with a flat base map and pick entry/exit points.
	gr := g.generateFlat()
	origin := position.RandomBorderPosition(gr.Width, gr.Length, gr.Height)
	origin, _ = gr.ForcePositionToGround(origin)
	
	// 2. Parameters: Randomize the width and depth of the river flow.
	width := tools.RandomInt(2, 4)
	waterlevel := tools.Min(tools.RandomInt(1, 3), tools.Max(0, gr.FindLowestLevel()-1))

	// 3. Pathing: Create waypoints and endpoints for the river flow.
	// This produces a meandering effect instead of a straight line.
	endpoints := pickRandomBorderPoints(gr, tools.RandomInt(1, 3))
	waypoints := pickRandomInternalPoints(gr, tools.RandomInt(1, 3))

	// 4. Segment Tracing: Connect origin -> waypoints -> endpoints with water cells.
	currentPos := origin
	for _, wp := range waypoints {
		currentPos = g.generateRiverSegment(gr, currentPos, wp, width, waterlevel)
	}
	for _, ep := range endpoints {
		connectRiverToEndpoint(g, gr, ep, waterlevel)
	}

	// 5. Finishing: Add obstructions while ensuring they don't block the water path.
	if g.GenerateObstrcution {
		placeRiverObstructions(g, gr)
	}
	return gr
}

// pickRandomBorderPoints is a helper to find valid exit points on the grid edges.
func pickRandomBorderPoints(gr *grid.Grid, count int) []position.Position {
	pts := []position.Position{}
	// 1. Generate the requested number of border coordinates.
	for i := 0; i < count; i++ {
		p := position.RandomBorderPosition(gr.Width, gr.Length, gr.Height)
		// 2. Ensure the point is snapped to the playable ground level.
		p, _ = gr.ForcePositionToGround(p)
		pts = append(pts, p)
	}
	// 3. Return the collection of exit points.
	return pts
}

// pickRandomInternalPoints is a helper to find path waypoints within the map boundaries.
func pickRandomInternalPoints(gr *grid.Grid, count int) []position.Position {
	pts := []position.Position{}
	// 1. Generate the requested number of internal coordinates.
	for i := 0; i < count; i++ {
		p := position.RandomPosition(gr.Width, gr.Length, gr.Height)
		// 2. Ensure the point is snapped to the playable ground level.
		p, _ = gr.ForcePositionToGround(p)
		pts = append(pts, p)
	}
	// 3. Return the collection of waypoints.
	return pts
}

// connectRiverToEndpoint ensures the river flows out to the grid boundary correctly.
// It bridges any gaps between the main river body and the map edge.
func connectRiverToEndpoint(g *GridGenerator, gr *grid.Grid, endpoint position.Position, waterlevel int) {
	// 1. Search Phase: Find the nearest existing water cell to the target border endpoint.
	// This identifies where the current river body ends relative to the boundary.
	nearestRiverCell, found := gr.FindNearestCellMatchingPredicate(endpoint, func(c *cell.Cell) bool {
		// We only care about water cells for this pathing calculation.
		return c.Type == cell.Water
	})

	// 2. Connectivity Validation: If a significant gap exists, we must bridge it.
	// A distance greater than 1 indicates the river hasn't reached the border yet.
	if found && nearestRiverCell.Position.Distance(endpoint) > 1 {
		// 3. Path Completion: Generate a final aquatic segment to the border coordinate.
		// This uses a random narrow width to simulate a natural river mouth.
		g.generateRiverSegment(gr, nearestRiverCell.Position, endpoint, tools.RandomInt(2, 3), waterlevel)
	}
	// 4. Connection verification and segment bridging completed.
}

// placeRiverObstructions adds obstacles to the grid while avoiding water cells.
// This ensures that the water path remains navigable for amphibious units and ships.
func placeRiverObstructions(g *GridGenerator, gr *grid.Grid) {
	// 1. Rate Calculation: Determine the total number of obstructions for this map.
	// The rate is derived from the generator's configured IntRange.
	obstruction := g.ObstructionRate.Random()
	
	// 2. Placement Loop: Iteratively attempt to place obstacles in valid locations.
	for i := 0; i < obstruction; i++ {
		// 3. Random Coordinate Selection: Pick a point and find its surface level.
		x := tools.RandomInt(0, gr.Width-1)
		y := tools.RandomInt(0, gr.Length-1)
		z := gr.TopMostGroundAt(x, y)
		
		// 4. Terrain Validation: Check the cell type at the target surface position.
		if c, found := gr.CellAt(position.New(x, y, z)); found && c.Type == cell.Water {
			// 5. Water Conflict: Skip this location to prevent blocking the river flow.
			// We decrement the counter to ensure the requested total count is still met.
			i--
			continue
		}
		// 6. Execution: Mark the valid ground cell as a tactical obstacle (rock/debris).
		gr.ReplaceCellType(position.New(x, y, z), cell.Obstacle)
	}
	// 7. Tactical riverbed hardening sequence complete.
}

// generateRiverSegment generates a river segment from a start position to an end position.
// It carves a channel and fills it with water cells at the specified depth.
func (g *GridGenerator) generateRiverSegment(gr *grid.Grid, start position.Position, end position.Position, width int, waterlevel int) position.Position {
	// 1. Path Calculation: Find all positions along the 2D path between start and end.
	for _, p := range pattern.PathTo2D(end.Substract(start)).Enlarge(width) {
		pos := position.New(start.X+p.X, start.Y+p.Y, waterlevel)
		
		// 2. Bed Carving: Remove any cells above the water level to create a channel.
		// This ensures the water is correctly recessed into the terrain.
		for z := waterlevel; z < gr.Height; z++ {
			delete(gr.Cells, position.New(pos.X, pos.Y, z))
		}
		
		// 3. Water Placement: Insert the water cell at the designated level.
		gr.Cells[pos] = &cell.Cell{
			Position: pos,
			Type:     cell.Water,
		}
	}
	// 4. Return the end position for chain generation.
	return end
}
