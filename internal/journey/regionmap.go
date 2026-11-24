package journey

import (
	"hash/fnv"
	"math/rand"
	"sort"
)

// Tile is one cell of a region map.
type Tile byte

const (
	TileBlock    Tile = iota // mountain, water, trees: can't be walked
	TilePath                 // the road: safe
	TileWild                 // wild ground beside the road: random encounters
	TileTown                 // a town on the road
	TileStage                // a tribulation; it bars the road until cleared
	TileEast                 // back east, to the previous region
	TileWest                 // on west, to the next region
	TileCrossing             // a river, pass or desert: it bars the road until crossed
)

// Point is a cell on a region map.
type Point struct{ X, Y int }

// RegionMap is a region drawn as tiles: a road winding from the east door,
// where the pilgrim arrives, to the west door, with the region's towns on
// it and a marker for each tribulation. A tribulation's column is walled
// off except for its marker, so the road west opens only as each is
// cleared; the last is the demon's lair.
type RegionMap struct {
	Region Region
	Index  int
	W, H   int
	Tiles  [][]Tile // [y][x]
	Start  Point    // where the pilgrim arrives from the east
	End    Point    // where the pilgrim arrives back from the west
	Towns  map[Point]Town
	Stages map[Point]int

	// The region's obstacle, if its terrain has one, and where it stands.
	Crossing *Crossing
	CrossAt  Point
}

const (
	mapW = 46
	mapH = 13
)

// At returns the tile at p; outside the map is blocked.
func (m RegionMap) At(p Point) Tile {
	if p.X < 0 || p.Y < 0 || p.X >= m.W || p.Y >= m.H {
		return TileBlock
	}
	return m.Tiles[p.Y][p.X]
}

// Lair is the marker of the region's last tribulation.
func (m RegionMap) Lair() Point {
	for p, n := range m.Stages {
		if n == m.Region.To {
			return p
		}
	}
	return m.End
}

// BuildMap draws region i, the same way every time.
func BuildMap(i int) RegionMap {
	r := content.Regions[i]
	h := fnv.New64a()
	h.Write([]byte(r.ID)) //nolint:errcheck // hash writes never fail
	rng := rand.New(rand.NewSource(int64(h.Sum64())))
	m := RegionMap{Region: r, Index: i, W: mapW, H: mapH, Towns: map[Point]Town{}, Stages: map[Point]int{}}
	m.Tiles = make([][]Tile, m.H)
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, m.W)
	}

	// The road's row in each column, wandering as it goes west.
	row := make([]int, m.W)
	row[m.W-2] = m.H / 2
	for x := m.W - 3; x >= 1; x-- {
		row[x] = row[x+1]
		if rng.Intn(100) < 35 {
			row[x] = min(max(row[x]+rng.Intn(3)-1, 2), m.H-3)
		}
	}

	// Events along the road, east to west: in stage order, a town before
	// the tribulation it shares a stage with.
	type event struct {
		stage int
		town  *Town
	}
	var events []event
	for _, t := range r.TownsIn() {
		events = append(events, event{stage: t.Stage, town: &t})
	}
	for n := r.From; n <= r.To; n++ {
		events = append(events, event{stage: n})
	}
	sort.SliceStable(events, func(a, b int) bool {
		if events[a].stage != events[b].stage {
			return events[a].stage < events[b].stage
		}
		return events[a].town != nil && events[b].town == nil
	})
	first, last := m.W-6, 4 // the stretch of road events are spread over
	cols := make([]int, len(events))
	for j := range events {
		cols[j] = first - (j+1)*(first-last)/(len(events)+1)
		// Straighten the road around an event so its column has one road cell.
		row[cols[j]+1], row[cols[j]-1] = row[cols[j]], row[cols[j]]
	}
	// The obstacle goes in the middle of the widest stretch between events.
	crossCol := 0
	if m.Crossing = CrossingFor(r.Terrain); m.Crossing != nil {
		edges := append([]int{m.W - 3}, cols...)
		edges = append(edges, 2)
		gap := 0
		for j := 1; j < len(edges); j++ {
			if d := edges[j-1] - edges[j]; d > gap {
				gap, crossCol = d, edges[j]+d/2
			}
		}
		row[crossCol+1], row[crossCol-1] = row[crossCol], row[crossCol]
	}

	// Carve the road, joining each column to the next with a vertical run.
	for x := m.W - 2; x >= 1; x-- {
		m.Tiles[row[x]][x] = TilePath
		if x < m.W-2 {
			for y := min(row[x], row[x+1]); y <= max(row[x], row[x+1]); y++ {
				m.Tiles[y][x] = TilePath
			}
		}
	}

	// Wild ground in patches beside the road.
	for k := 0; k < m.W/3; k++ {
		cx := 2 + rng.Intn(m.W-4)
		cy := row[cx] + rng.Intn(5) - 2
		rx, ry := 2+rng.Intn(4), 1+rng.Intn(2)
		for y := cy - ry; y <= cy+ry; y++ {
			for x := cx - rx; x <= cx+rx; x++ {
				if x > 0 && y > 0 && x < m.W-1 && y < m.H-1 && m.Tiles[y][x] == TileBlock {
					m.Tiles[y][x] = TileWild
				}
			}
		}
	}

	// Place the events; a tribulation walls off its column.
	for j, e := range events {
		p := Point{cols[j], row[cols[j]]}
		if e.town != nil {
			m.Tiles[p.Y][p.X] = TileTown
			m.Towns[p] = *e.town
			continue
		}
		for y := 0; y < m.H; y++ {
			m.Tiles[y][p.X] = TileBlock
		}
		m.Tiles[p.Y][p.X] = TileStage
		m.Stages[p] = e.stage
	}

	if m.Crossing != nil {
		for y := 0; y < m.H; y++ {
			m.Tiles[y][crossCol] = TileBlock
		}
		m.CrossAt = Point{crossCol, row[crossCol]}
		m.Tiles[m.CrossAt.Y][crossCol] = TileCrossing
	}

	m.Start, m.End = Point{m.W - 2, row[m.W-2]}, Point{1, row[1]}
	if i > 0 {
		m.Tiles[m.Start.Y][m.W-1] = TileEast
	}
	if i < len(content.Regions)-1 {
		m.Tiles[m.End.Y][0] = TileWest
	}
	return m
}
