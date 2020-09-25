package game

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Dictionary of valid words.
var globalDict = initDictionary("game/sowpods.txt")

func initDictionary(filename string) map[string]struct{} {
	d := make(map[string]struct{})
	f, err := os.Open(filename)
	if err != nil {
		log.Println("Could not open dictionary list:", err)
		panic("No dictionary!")
		return d
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		word := scanner.Text()
		d[word] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		log.Println("Could not load dictionary:", err)
		panic("No dictionary!")
		return d
	}
	log.Println("Finished initializing dictionary from", filename)
	return d
}

// verifyWord returns whether the given string is in the loaded dictionary.
func verifyWord(w string) bool {
	_, ok := globalDict[w]
	return ok
}

// Vec is a struct used for vector calculations.
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

// VecSet is a set of Vec.
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

// A Score contains information found when scoring a board.
type Score struct {
	Win     bool   // Whether the board ends the game or not.
	Score   int    // The numerical score (lower is better).
	Invalid VecSet // Tiles which were found to be invalid or not connected.
}

// A Board is a grid of tiles of arbitrary size.
// This type is used to unmarshal a board from the page client.
type Board [][]*Tile

// value returns the value of the board at the given location.
func (b Board) value(v Vec) string {
	t := b[v.Y][v.X]
	if t == nil {
		return ""
	}
	return t.Value
}

// createVecSetOfAllTiles returns a set of all the tiles present on this board.
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

// findConnectedTiles puts all the tiles connected to the given location into the given set.
func (s VecSet) findConnectedTiles(v Vec, found VecSet) {
	if !found.contains(v) && s.contains(v) {
		found.insert(v)
		s.findConnectedTiles(v.add(Vec{1, 0}), found)
		s.findConnectedTiles(v.add(Vec{0, 1}), found)
		s.findConnectedTiles(v.add(Vec{-1, 0}), found)
		s.findConnectedTiles(v.add(Vec{0, -1}), found)
	}
}

// findComponents returns the disjoint word sections for this board.
// E.g. if the board has two unconnected sections of tiles, this function returns a len 2 slice.
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

// followWord takes a starting location, a set of connected tile coordinates, and a direction,
// and returns the coordinates of tiles in the word and a boolean of whether the word is in
// the dictionary.
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

// bruteForceScorableComponents decides which tiles to discard when there is a combination
// of valid and invalid words in a single component.  Recursively remove tiles one by one,
// checking the effect on the final score.
func (b Board) bruteForceScorableComponents(valid, invalid VecSet) []VecSet {
	bestScore := 0
	var bestChoice []VecSet
	tempInvalid := make(VecSet)
	for elt := range invalid {
		attempt := make(VecSet)
		attempt.union(valid)
		delete(attempt, elt)
		attemptComps := b.findValidScorableComponents(attempt, tempInvalid)
		attemptScore := b.bestScore(attemptComps)
		if attemptScore > bestScore {
			bestScore = attemptScore
			bestChoice = attemptComps
		}
	}
	return bestChoice
}

// findValidScorableComponents returns the different tile sets that only form valid words.
// The given invalidAll set will be modified to include invalid tiles from all components.
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
		// Check all words in the down and right directions.
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
				// If all tiles were used in valid words but there were also invalid words,
				// brute force which tiles to drop.
				bestChoice := b.bruteForceScorableComponents(valid, invalid)
				scorable = append(scorable, bestChoice...)
				continue
			}
			// If everything was good, keep this component.
			scorable = append(scorable, valid)
		} else {
			// If some tiles in this component were not in any valid word, throw those out and
			// call this function again on the subset.
			subcomps := b.findValidScorableComponents(valid, invalidAll)
			scorable = append(scorable, subcomps...)
		}
	}
	return scorable
}

// bestScore will return the best score from the given slice of components (assumes they are valid).
func (b Board) bestScore(cs []VecSet) int {
	bestScore := 0
	for _, comp := range cs {
		newScore := 0
		for v := range comp {
			newScore += pointValues[b.value(v)]
		}
		if newScore > bestScore {
			bestScore = newScore
		}
	}
	return bestScore
}

// compareTileValues will return false if board is not a subset of tiles served.
// Takes in the tiles that have been served and the tiles on this board.
func (b Board) compareTileValues(sent []Tile, received VecSet) bool {
	tileCount := make(map[string]int)
	for _, elt := range sent {
		tileCount[elt.Value] += 1
	}
	for key := range received {
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

// scoreBoard returns the overall score for this board given the tiles served.
// This function is called by the client.
func (b Board) scoreBoard(tilesServed []Tile) Score {
	// Calculate score if board is empty.
	maxScore := 0
	for _, elt := range tilesServed {
		maxScore += elt.Points
	}
	result := Score{Win: true, Score: maxScore, Invalid: make(VecSet)}
	boardSet := b.createVecSetOfAllTiles()

	// Find the hightest point value of all scorable components
	components := b.findValidScorableComponents(boardSet, result.Invalid)
	bestScore := b.bestScore(components)
	result.Score = maxScore - bestScore
	log.Println("Best score:", bestScore)
	log.Println("Invalid tiles:", result.Invalid)

	// A winning board has a score of 0.
	if result.Score != 0 {
		result.Win = false
		if result.Score < 0 {
			log.Println("Impossible score: cheating suspected!")
			result.Score = maxScore
		}
	}
	// A winning board must contain all the tiles served and no more.
	tilesServedCount := len(tilesServed)
	if len(boardSet) != tilesServedCount {
		result.Win = false
		if len(boardSet) > tilesServedCount {
			log.Println("Impossible number of tiles: cheating suspected!")
			result.Score = maxScore
			return result
		}
	}
	// A winning board must contain exactly the tiles served.
	if !b.compareTileValues(tilesServed, boardSet) {
		log.Println("Impossibly mismatched tiles: cheating suspected!")
		result.Win = false
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
			if b[i][j] == nil {
				result += " "
			} else {
				result += b[i][j].Value
			}
		}
		result += "\n"
	}
	return result
}
