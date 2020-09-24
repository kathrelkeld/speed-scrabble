package game

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var globalDict = initDictionary("game/sowpods.txt")

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
	X int
	Y int
}

func (a Vec) add(b Vec) Vec {
	return Vec{a.X + b.X, a.Y + b.Y}
}

func (v Vec) scale(c int) Vec {
	return Vec{v.X * c, v.Y * c}
}

func (v Vec) String() string {
	return fmt.Sprintf("(%v, %v)", v.X, v.Y)
}

type VecSet map[Vec]struct{}

func (s VecSet) contains(elt Vec) bool {
	_, ok := s[elt]
	return ok
}

func (s VecSet) insert(elt Vec) {
	s[elt] = struct{}{}
}

func (s VecSet) MarshalJSON() ([]byte, error) {
	var result []Vec
	for elt := range s {
		result = append(result, elt)
	}
	return json.Marshal(result)
}

func (s VecSet) UnmarshalJSON(b []byte) error {
	var vecs []Vec
	if err := json.Unmarshal(b, &vecs); err != nil {
		return err
	}
	for _, elt := range vecs {
		s.insert(elt)
	}
	return nil
}

// Adds the elements of set b to the set a.
func (a VecSet) union(b VecSet) {
	for elt := range b {
		a.insert(elt)
	}
}

func (s VecSet) String() string {
	if len(s) == 0 {
		return ""
	}
	result := "["
	for elt := range s {
		result += fmt.Sprintf("%v, ", elt)
	}
	return result[:len(result)-2] + "]"
}

type Score struct {
	Valid   bool
	Score   int
	Invalid VecSet
}

type Board [][]string

func (b Board) value(v Vec) string {
	return b[v.X][v.Y]
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
	return verifyWord(word), result
}

func (b Board) bruteForceScorableComponents(valid, invalid VecSet) []VecSet {
	bestScore := 0
	var bestChoice []VecSet
	tempInvalid := make(VecSet)
	for elt := range invalid {
		log.Println("Removing", b.value(elt))
		attempt := make(VecSet)
		attempt.union(valid)
		delete(attempt, elt)
		attemptComps := b.findValidScorableComponents(attempt, tempInvalid)
		attemptScore := b.scoreComponentList(attemptComps)
		if attemptScore > bestScore {
			bestScore = attemptScore
			bestChoice = attemptComps
		}
	}
	log.Println("Brute force result:", bestChoice)
	return bestChoice
}

// Verify that the given component list has valid words.
// Add tiles that are part of valid words to valid and invalid words to invalid.
// These two sets may overlap.
func (b Board) findValidScorableComponents(boardSet, invalidAll VecSet) []VecSet {
	components := boardSet.findComponents()
	scorable := []VecSet{}
	for _, comp := range components {
		// Component must be 2 or more tiles.
		if len(comp) <= 1 {
			invalidAll.union(comp)
			continue
		}
		valid := make(VecSet)
		invalid := make(VecSet)
		// Check all tiles in the vert and horizontal directions.
		for v := range comp {
			for _, direction := range []Vec{{1, 0}, {0, 1}} {
				// Follow this word if it's at the start in this direction.
				if !comp.contains(v.add(direction.scale(-1))) {
					isWord, wordSet := b.followWord(v, comp, direction)
					if !isWord {
						invalid.union(wordSet)
					} else {
						valid.union(wordSet)
					}
				}
			}
		}
		invalidAll.union(invalid)
		if len(valid) == len(comp) {
			if len(invalid) != 0 {
				// If all tiles were valid but there were invalid words, brute force.
				log.Println("Found unremovable tiles!")
				bestChoice := b.bruteForceScorableComponents(valid, invalid)
				scorable = append(scorable, bestChoice...)
				continue
			}
			// If everything was good, keep this component.
			scorable = append(scorable, valid)
		} else {
			// If some tiles were invalid, find the components in the valid subset.
			subcomps := b.findValidScorableComponents(valid, invalidAll)
			scorable = append(scorable, subcomps...)
		}
	}
	return scorable
}

// Assuming that all the tiles are valid, find the max score in this list.
func (b Board) scoreComponentList(cs []VecSet) int {
	bestScore := 0
	for _, comp := range cs {
		newScore := 0
		for v := range comp {
			newScore += pointValues[b.value(v)]
		}
		if newScore > bestScore {
			bestScore = newScore
		}
		log.Println("Found component with score", newScore)
	}
	return bestScore
}

// Return false if set and board have mismatched tile values.
func (b Board) compareTileValues(c *Client, s VecSet) bool {
	t := c.getAllTilesServed()
	tileCount := make(map[string]int)
	for _, elt := range t {
		tileCount[elt.Value] += 1
	}
	for key := range s {
		tileCount[b.value(key)] -= 1
	}
	// Return false if an impossible value is found.
	for _, count := range tileCount {
		if count < 0 {
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
	result := Score{Valid: true, Score: maxScore, Invalid: make(VecSet)}
	boardSet := b.createVecSetOfAllTiles()

	// Find the hightest point value of all scorable components
	components := b.findValidScorableComponents(boardSet, result.Invalid)
	bestScore := b.scoreComponentList(components)
	result.Score -= bestScore
	log.Println("Best score:", bestScore)
	log.Println("Invalid tiles:", result.Invalid)

	// A valid board has a score of 0.
	if result.Score != 0 {
		result.Valid = false
		if result.Score < 0 {
			log.Println("Impossible score: cheating suspected!")
			result.Score = maxScore
		}
	}
	// A valid board must contain all the tiles served and no more.
	tilesServedCount := c.getTilesServedCount()
	if len(boardSet) != tilesServedCount {
		result.Valid = false
		if len(boardSet) > tilesServedCount {
			log.Println("Impossible number of tiles: cheating suspected!")
			result.Score = maxScore
			return result
		}
	}
	// A valid board must contain exactly the tiles served.
	if !b.compareTileValues(c, boardSet) {
		log.Println("Impossibily mismatched tiles: cheating suspected!")
		result.Valid = false
		result.Score = maxScore
		return result
	}
	// Return true if all other checks have passed.
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
