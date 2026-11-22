package journey

import "testing"

// reach floods the walkable cells from p, stopping at markers unless open.
func reach(m RegionMap, from Point, open func(stage int) bool) map[Point]bool {
	seen := map[Point]bool{from: true}
	queue := []Point{from}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if n, ok := m.Stages[p]; ok && !open(n) && p != from {
			continue // a closed tribulation: reached, but not passed
		}
		for _, d := range []Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			q := Point{p.X + d.X, p.Y + d.Y}
			if !seen[q] && m.At(q) != TileBlock {
				seen[q] = true
				queue = append(queue, q)
			}
		}
	}
	return seen
}

func TestEveryRegionMapIsWalkableAndGated(t *testing.T) {
	for i, r := range content.Regions {
		m := BuildMap(i)
		if len(m.Stages) != r.To-r.From+1 || len(m.Towns) != len(r.TownsIn()) {
			t.Errorf("%s: %d stages, %d towns", r.ID, len(m.Stages), len(m.Towns))
		}
		all := reach(m, m.Start, func(int) bool { return true })
		if !all[m.End] || !all[m.Lair()] {
			t.Errorf("%s: the road doesn't run from the east door to the west", r.ID)
		}
		for p := range m.Towns {
			if !all[p] {
				t.Errorf("%s: a town can't be reached", r.ID)
			}
		}
		// With nothing cleared, the first tribulation stops the pilgrim.
		shut := reach(m, m.Start, func(int) bool { return false })
		if shut[m.End] {
			t.Errorf("%s: the west door can be reached without clearing a tribulation", r.ID)
		}
		var firstStage Point
		for p, n := range m.Stages {
			if n == r.From {
				firstStage = p
			}
		}
		if !shut[firstStage] {
			t.Errorf("%s: the first tribulation can't be reached", r.ID)
		}
		if BuildMap(i).Tiles[m.H/2][m.W/2] != m.Tiles[m.H/2][m.W/2] {
			t.Errorf("%s: maps differ between builds", r.ID)
		}
	}
}
