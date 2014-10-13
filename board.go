package main

import (
	"bufio"
	"log"
	"os"
)

var globalDict = initDictionary("sowpods.txt")

func initDictionary(filename string) map[string]struct{} {
	log.Println("Initializing dictionary from", filename)
	d := make(map[string]struct{})
	f, err := os.Open(filename)
	if err != nil {
		log.Println("err:", err)
		return d
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		word := scanner.Text()
		d[word] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		log.Println("err:", err)
		return d
	}
	log.Println("Finished initializing dictionary")
	return d
}

func verifyWord(w string) bool {
	_, ok := globalDict[w]
	return ok
}

type Vec struct {
	x int
	y int
}

func (a Vec) add(b Vec) Vec {
	return Vec{a.x + b.x, a.y + b.y}
}

type VecSet map[Vec]struct{}

func (s VecSet) contains(elt Vec) bool {
	_, ok := s[elt]
	return ok
}

func (s VecSet) insert(elt Vec) {
	s[elt] = struct{}{}
}

type Score struct {
	Valid bool
	Score int
}

type Board [][]string

func makeBoard(x, y int, letters ...string) Board {
	b := Board{}
	if len(letters) != x*y {
		log.Println("Letters count did not match given dimensions!")
		return b
	}
	for i := 0; i < x; i++ {
		b = append(b, letters[i*y:(i+1)*y])
	}
	return b
}

func (b Board) value(v Vec) string {
	return b[v.x][v.y]
}

// Find the tiles connected to the given tile.
func (b Board) findConnectedTiles(v Vec, found VecSet) {
	if !found.contains(v) &&
		v.x < len(b) && v.y < len(b[0]) &&
		v.x >= 0 && v.y >= 0 && b.value(v) != "" {
		found.insert(v)
		b.findConnectedTiles(v.add(Vec{1, 0}), found)
		b.findConnectedTiles(v.add(Vec{0, 1}), found)
		b.findConnectedTiles(v.add(Vec{-1, 0}), found)
		b.findConnectedTiles(v.add(Vec{0, -1}), found)
	}
}

// Find the connected components
func (b Board) findComponents() []VecSet {
	var result []VecSet
	maxX := len(b)
	maxY := len(b[0])
	var i, j int
	c := make(VecSet)
	for i = 0; i < maxX; i++ {
		for j = 0; j < maxY; j++ {
			v := Vec{i, j}
			if !c.contains(v) && b.value(v) != "" {
				thisCompFound := make(VecSet)
				b.findConnectedTiles(v, thisCompFound)
				result = append(result, thisCompFound)
				for elt := range thisCompFound {
					c.insert(elt)
				}
			}
		}
	}
	log.Println("Found", len(result), "components")
	return result
}

// Given a tile and a direction, follow the word in that direction.
// Returns false if any problems are found
// (i.e. more than 1 letter yet not a word)
func (b Board) followWord(v Vec, comp VecSet, d Vec) bool {
	next := v.add(d)
	if !comp.contains(next) {
		return true
	}
	word := b.value(v)
	for comp.contains(next) {
		word += b.value(next)
		next = next.add(d)
	}
	log.Println("Found word", word)
	return verifyWord(word)
}

// Verify that the given component list has valid words.
func (b Board) verifyWordsInComponent(comp VecSet) bool {
	// Component must be 2 or more tiles.
	if len(comp) <= 1 {
		return false
	}
	// Check all tiles in the vert and horizontal directions.
	for v := range comp {
		if !comp.contains(v.add(Vec{-1, 0})) {
			if !b.followWord(v, comp, Vec{1, 0}) {
				return false
			}
		}
		if !comp.contains(v.add(Vec{0, -1})) {
			if !b.followWord(v, comp, Vec{0, 1}) {
				return false
			}
		}
	}
	return true
}

func (b Board) scoreComponent(c VecSet) int {
	score := 0
	for v, _ := range c {
		score += pointValues[b.value(v)]
	}
	return score
}

// Compare a game's tiles to the tiles pointed to by a given set.
func (b Board) compareTiles(c *Client, s VecSet) bool {
	// Return false if set and game do not have same count of tiles.
	tilesServed := c.getTilesServedCount()
	if len(s) != tilesServed {
		log.Println("Game and board were not same length")
		return false
	}
	// Return false if set and game do not agree on tile values.
	t := c.getAllTilesServed()
	tileCount := make(map[string]int)
	for _, elt := range t {
		tileCount[elt.Value] += 1
	}
	for key := range s {
		tileCount[b.value(key)] -= 1
	}
	log.Println("Count:", tileCount)
	for key, count := range tileCount {
		if count != 0 {
			log.Println("Game and board did not agree on", key, "values")
			return false
		}
	}
	// Return true if all other checks have passed.
	return true
}

// Return true if this board is a valid soultion
func (b Board) scoreBoard(c *Client) Score {
	result := Score{Valid: false, Score: c.getMaxScore()}

	// Find all components on this board and score them.
	components := b.findComponents()
	score := 0
	for _, comp := range components {
		// A valid component contains only valid words.
		if b.verifyWordsInComponent(comp) {
			newScore := b.scoreComponent(comp)
			if newScore > score {
				score = newScore
			}
			log.Println("Found component with score", newScore)
		}
	}
	result.Score -= score

	// A valid board has a score of 0
	if result.Score != 0 {
		return result
	}
	// A valid board has only one component.
	if len(components) != 1 {
		return result
	}
	// A valid board must contain exactly the tiles served.
	comp := components[0]
	if !b.compareTiles(c, comp) {
		return result
	}
	// Return true if all other checks have passed.
	result.Valid = true
	return result
}

func (b Board) String() string {
	result := ""
	for i := 0; i < len(b); i++ {
		for j := 0; j < len(b[0]); j++ {
			if b[i][j] == "" {
				result += " "
			} else {
				result += b[i][j]
			}
		}
		result += "\n"
	}
	return result
}
