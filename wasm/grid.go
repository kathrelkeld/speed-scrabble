package main

type Grid struct {
	// Grid is the slice of slices representation of the entire grid.
	Grid [][]*Tile
	// Loc is the canvas coordinates where this grid is drawn.
	Loc  Vec
	Zone int
}

// newInnerGrid creates a slice of slices needed for a Grid struct of the given size.
func newInnerGrid(size Vec) [][]*Tile {
	var b [][]*Tile
	for j := 0; j < size.Y; j++ {
		b = append(b, []*Tile{})
		for i := 0; i < size.X; i++ {
			b[j] = append(b[j], nil)
		}
	}
	return b
}

func (g *Grid) AddTile(t *Tile, idx Vec) {
	t.pickUp()
	g.Set(idx, t)
	t.Zone = g.Zone
	t.Idx = idx
	t.Loc = Add(g.Loc, Mult(mgr.tileSize, idx))
}

func (g *Grid) IdxSize() Vec {
	return Vec{len(g.Grid[0]), len(g.Grid)}
}

func (g *Grid) CanvasEnd() Vec {
	size := g.IdxSize()
	return Add(g.Loc, Mult(mgr.tileSize, size))
}

func (g *Grid) Get(idx Vec) *Tile {
	return g.Grid[idx.Y][idx.X]
}

func (g *Grid) Set(idx Vec, tile *Tile) {
	g.Grid[idx.Y][idx.X] = tile
}

func (g *Grid) InCanvas(l Vec) bool {
	end := g.CanvasEnd()
	return l.inTarget(g.Loc, end)
}

func (g *Grid) InCoords(idx Vec) bool {
	size := g.IdxSize()
	return idx.X >= 0 && idx.X < size.X && idx.Y >= 0 && idx.Y < size.Y
}

func (g *Grid) coords(loc Vec) Vec {
	return Vec{
		(loc.X - g.Loc.X) / mgr.tileSize.X,
		(loc.Y - g.Loc.Y) / mgr.tileSize.Y,
	}
}

func (g *Grid) canvasStart(idx Vec) Vec {
	return Add(g.Loc, Mult(idx, mgr.tileSize))
}
