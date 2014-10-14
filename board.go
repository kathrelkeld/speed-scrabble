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

func (v Vec) scale(c int) Vec {
	return Vec{v.x * c, v.y * c}
}

type VecSet map[Vec]struct{}

func (s VecSet) contains(elt Vec) bool {
	_, ok := s[elt]
	return ok
}

func (s VecSet) insert(elt Vec) {
	s[elt] = struct{}{}
}

// Adds the elements of set b to the set a.
func (a VecSet) union(b VecSet) {
	for elt := range b {
		a.insert(elt)
	}
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

// Create a VecSet containing all the tiles present on this board.
func (b Board) createVecSetOfAllTiles() VecSet {
	result := VecSet{}
	for i := 0; i < len(b); i++ {
		for j := 0; j < len(b[0]); j++ {
			v := Vec{i, j}
			if b.value(v) != "" {
				result.insert(v)
			}
		}
	}
	return result
}

// Find the tiles connected to the given tile in this component.
func (s VecSet) findConnectedTiles(v Vec, found VecSet) {
	if !found.contains(v) && s.contains(v) {
		found.insert(v)
		s.findConnectedTiles(v.add(Vec{1, 0}), found)
		s.findConnectedTiles(v.add(Vec{0, 1}), found)
		s.findConnectedTiles(v.add(Vec{-1, 0}), found)
		s.findConnectedTiles(v.add(Vec{0, -1}), found)
	}
}

// Find the connected components in this component.
func (s VecSet) findComponents() []VecSet {
	var result []VecSet
	c := make(VecSet)
	for v := range s {
		if !c.contains(v) {
			thisCompFound := make(VecSet)
			s.findConnectedTiles(v, thisCompFound)
			result = append(result, thisCompFound)
			for elt := range thisCompFound {
				c.insert(elt)
			}
		}
	}
	log.Println("Found", len(result), "components")
	return result
}

// Given a tile and a direction, follow the word in that direction.
// Returns the word, if any.
func (b Board) followWord(v Vec, comp VecSet, d Vec) (bool, VecSet) {
	result := VecSet{}
	next := v.add(d)
	if !comp.contains(next) {
		return true, result
	}
	result.insert(v)
	word := b.value(v)
	for comp.contains(next) {
		result.insert(next)
		word += b.value(next)
		next = next.add(d)
	}
	log.Println("Found supposed word:", word)
	return verifyWord(word), result
}

// Verify that the given component list has valid words.
// Add tiles that are part of valid words to valid and invalid words to invalid.
// These two sets may overlap.
func (b Board) findValidScorableComponents(boardSet, invalid VecSet) []VecSet {
	components := boardSet.findComponents()
	scorable := []VecSet{}
	for _, comp := range components {
		// Component must be 2 or more tiles.
		if len(comp) <= 1 {
			invalid.union(comp)
			continue
		}
		valid := make(VecSet)
		isAllValid := true
		// Check all tiles in the vert and horizontal directions.
		for v := range comp {
			for _, direction := range []Vec{{1, 0}, {0, 1}} {
				// Follow this word if it's at the start in this direction.
				if !comp.contains(v.add(direction.scale(-1))) {
					isWord, wordSet := b.followWord(v, comp, direction)
					if !isWord {
						invalid.union(wordSet)
						isAllValid = false
					} else {
						valid.union(wordSet)
					}
				}
			}
		}
		if len(valid) == len(comp) {
			if !isAllValid {
				// If all tiles were valid but there were invalid words, stop.
				//TODO: brute force the best solution here instead
				continue
			}
			// If everything was good, keep this component.
			scorable = append(scorable, valid)
		} else {
			// If some tiles were invalid, find the components in the valid subset.
			subcomps := valid.findComponents()
			scorable = append(scorable, subcomps...)
		}
	}
	return scorable
}

// Assuming that all the tiles are valid, find the score of this set.
func (b Board) scoreComponent(c VecSet) int {
	score := 0
	for v := range c {
		score += pointValues[b.value(v)]
	}
	return score
}

// Return false if set and game do not agree on tile values.
func (b Board) compareTileValues(c *Client, s VecSet) bool {
	t := c.getAllTilesServed()
	tileCount := make(map[string]int)
	for _, elt := range t {
		tileCount[elt.Value] += 1
	}
	for key := range s {
		tileCount[b.value(key)] -= 1
	}
	for _, count := range tileCount {
		if count != 0 {
			log.Println("Value difference counts:", tileCount)
			return false
		}
	}
	// Return true if all other checks have passed.
	return true
}

// Score this board
func (b Board) scoreBoard(c *Client) Score {
	maxScore := c.getMaxScore()
	result := Score{Valid: false, Score: maxScore}

	// Find the hightest point value of all scorable components
	boardSet := b.createVecSetOfAllTiles()
	invalid := make(VecSet)
	components := b.findValidScorableComponents(boardSet, invalid)
	score := 0
	for _, comp := range components {
		newScore := b.scoreComponent(comp)
		if newScore > score {
			score = newScore
		}
		log.Println("Found component with score", newScore)
	}
	result.Score -= score
	log.Println("Invalid tiles:", invalid)

	// A valid board has a score of 0.
	if result.Score != 0 {
		if result.Score < 0 {
			log.Println("Impossible score: cheating suspected!")
			result.Score = maxScore
		}
		return result
	}
	// A valid board has only one component.
	if len(components) != 1 {
		return result
	}
	comp := components[0]
	// A valid board must contain all the tiles served and no more.
	tilesServedCount := c.getTilesServedCount()
	if len(comp) != tilesServedCount {
		if len(comp) > tilesServedCount {
			log.Println("Impossible number of tiles: cheating suspected!")
			result.Score = maxScore
		}
		return result
	}
	// A valid board must contain exactly the tiles served.
	if !b.compareTileValues(c, comp) {
		log.Println("Impossibily mismatched tiles: cheating suspected!")
		result.Score = maxScore
		return result
	}
	// Return true if all other checks have passed.
	result.Valid = true
	return result
}

// Printable board.
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
