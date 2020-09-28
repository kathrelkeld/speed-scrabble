package game

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

// Dictionary of valid words.
type Dict map[string]struct{}

var globalDict Dict

// Called by server.
func InitDictionary() {
	globalDict = loadDictionary("game/sowpods.txt")
}

func loadDictionary(filename string) Dict {
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

type TileSet map[Vec]*Tile

func (s TileSet) contains(elt Vec) bool {
	_, ok := s[elt]
	return ok
}

func (a TileSet) union(b TileSet) {
	for k, v := range b {
		a[k] = v
	}
}

// A Board is a grid of tiles of arbitrary size.
// This type is used to unmarshal a board from the page client.
type Board [][]*Tile

// setOfAllTiles returns a set of all the tiles present on this board.
func (b Board) setOfAllTiles() TileSet {
	result := TileSet{}
	for j := 0; j < len(b); j++ {
		for i := 0; i < len(b[0]); i++ {
			if b[j][i] != nil {
				result[Vec{i, j}] = b[j][i]
			}
		}
	}
	return result
}

type Word struct {
	Start  Vec     // Coordinates of the word start.
	End    Vec     // Coordinates of the word end.
	Value  string  // Value of the word.
	isWord bool    // Whether this word is an actual word or not.
	tiles  TileSet // Tiles in this word.
}

func (w Word) String() string {
	return w.Value
}

type Score struct {
	Win         bool     // Whether the board ends the game or not.
	Pts         int      // The numerical score (lower is better).
	Valid       TileSet  // Tiles which were part of no valid words.
	Invalid     TileSet  // Tiles which were part of no valid words.
	Unconnected TileSet  // Tiles not part of the best scoring component.
	Words       []Word   // Words found in the dictionary.
	Nonwords    []Word   // Words not found in the dictionary.
	Msg         msg.Type // OK or Error message to send to player.
}

func (s *Score) String() string {
	result := "Score:"
	result += fmt.Sprintf("\n    Win: %v", s.Win)
	result += fmt.Sprintf("\n    Pts: %v", s.Pts)
	result += fmt.Sprintf("\n    Valid: %v", s.Valid)
	result += fmt.Sprintf("\n    Unconnected: %v", s.Unconnected)
	result += fmt.Sprintf("\n    Invalid: %v", s.Invalid)
	result += fmt.Sprintf("\n    Words: ")
	for _, w := range s.Words {
		result += w.Value + " "
	}
	result += fmt.Sprintf("\n    Nonwords: ")
	for _, w := range s.Nonwords {
		result += w.Value + " "
	}
	return result
}

type Scorable struct {
	valid       TileSet
	invalid     TileSet
	unconnected TileSet
	words       []Word
	nonwords    []Word
}

func (s *Scorable) String() string {
	result := "Scorable:"
	result += fmt.Sprintf("\n    valid: %v", s.valid)
	result += fmt.Sprintf("\n    invalid: %v", s.invalid)
	result += fmt.Sprintf("\n    unconnected: %v", s.unconnected)
	result += fmt.Sprintf("\n    words: ")
	for _, w := range s.words {
		result += w.Value + " "
	}
	result += fmt.Sprintf("\n    nonwords: ")
	for _, w := range s.nonwords {
		result += w.Value + " "
	}
	return result
}

func (s *Scorable) score() int {
	score := 0
	for _, t := range s.valid {
		score += t.Points
	}
	return score
}

// findConnectedTiles puts all the tiles connected to this location into the given TileSet.
func (s TileSet) findConnectedTiles(v Vec, connected TileSet) {
	if connected.contains(v) {
		return
	}
	if s.contains(v) {
		connected[v] = s[v]
		s.findConnectedTiles(v.add(Vec{1, 0}), connected)
		s.findConnectedTiles(v.add(Vec{0, 1}), connected)
		s.findConnectedTiles(v.add(Vec{-1, 0}), connected)
		s.findConnectedTiles(v.add(Vec{0, -1}), connected)
	}
}

// findDisjointSets returns the disjoint word sections from the given TileSet.
// E.g. if the board has two unconnected sections of tiles, this function returns a len 2 slice.
func (s TileSet) findDisjointSets() []TileSet {
	var comps []TileSet
	used := make(TileSet)
	for loc := range s {
		if used.contains(loc) {
			continue
		}
		connected := make(TileSet)
		s.findConnectedTiles(loc, connected)
		comps = append(comps, connected)
		used.union(connected)
	}
	return comps
}

// followWord
func (comp TileSet) followWord(v Vec, d Vec) Word {
	w := Word{
		Start: v,
		Value: comp[v].Value,
		tiles: make(TileSet),
	}
	w.tiles[v] = comp[v]
	prev := v
	next := v.add(d)
	for comp.contains(next) {
		w.tiles[next] = comp[next]
		w.Value += comp[next].Value
		prev = next
		next = next.add(d)
	}

	w.End = prev
	w.isWord = verifyWord(w.Value)
	return w
}

func (s TileSet) extractScorable() *Scorable {
	overallSc := &Scorable{
		valid:       make(TileSet),
		invalid:     make(TileSet),
		unconnected: make(TileSet),
	}
	comps := s.findDisjointSets()
	fmt.Println("extracting scorable from multiple ", len(comps))
	allScores := []*Scorable{}
	for _, c := range comps {
		allScores = append(allScores, c.extractScorableFromConnected())
	}

	if len(allScores) == 0 {
		return overallSc
	}

	bestScore := allScores[0].score()
	bestScorable := allScores[0]
	for i := 1; i < len(allScores); i++ {
		sc := allScores[i]
		if pts := sc.score(); pts > bestScore {
			bestScore = pts
			bestScorable, sc = sc, bestScorable
		}
		overallSc.unconnected.union(sc.valid)
		overallSc.unconnected.union(sc.unconnected)
		overallSc.invalid.union(sc.invalid)
		overallSc.words = append(overallSc.words, sc.words...)
		overallSc.nonwords = append(overallSc.nonwords, sc.nonwords...)
	}

	overallSc.valid = bestScorable.valid
	overallSc.invalid.union(bestScorable.invalid)
	overallSc.unconnected.union(bestScorable.unconnected)
	overallSc.words = append(overallSc.words, bestScorable.words...)
	overallSc.nonwords = append(overallSc.nonwords, bestScorable.nonwords...)
	return overallSc
}

// Return the best scorable object that can be made from these connected tiles.
func (s TileSet) extractScorableFromConnected() *Scorable {
	fmt.Println("extracting scorable from a single component")
	// Words must be 2 or more tiles.
	sc := &Scorable{
		valid:    make(TileSet),
		invalid:  make(TileSet),
		words:    []Word{},
		nonwords: []Word{},
	}
	if len(s) <= 1 {
		sc.invalid.union(s)
		return sc
	}

	// Check all words in the down and right directions.
	for v := range s {
		for _, direction := range []Vec{{1, 0}, {0, 1}} {
			if s.contains(v.add(direction.scale(-1))) {
				// Ignore if this is not the start of a word.
				continue
			}
			if !s.contains(v.add(direction)) {
				// Ignore "words" that are only one letter.
				continue
			}
			w := s.followWord(v, direction)
			if !w.isWord {
				sc.invalid.union(w.tiles)
				sc.nonwords = append(sc.nonwords, w)
			} else {
				sc.valid.union(w.tiles)
				sc.words = append(sc.words, w)
			}
		}
	}

	if len(sc.valid) == len(s) {
		if len(sc.invalid) != 0 {
			// If all tiles were used in valid words but there were also invalid words,
			// brute force which tiles to drop.
			bruteForce := sc.valid.bruteForceScorable(sc.invalid)
			bruteForce.words = sc.words
			bruteForce.nonwords = sc.nonwords
			sc = bruteForce
		}
	} else {
		// If some tiles in this component were not in any valid word, throw those out and
		// call this function again on the subset.
		subSc := sc.valid.extractScorable()
		for v := range sc.invalid {
			if !sc.valid.contains(v) {
				subSc.invalid[v] = s[v]
			}
		}
		subSc.words = sc.words
		subSc.nonwords = sc.nonwords
		sc = subSc
	}
	return sc
}

func (s TileSet) bruteForceScorable(invalid TileSet) *Scorable {
	fmt.Println("brute forcing a solution")
	bestScore := 0
	var bestChoice *Scorable
	for elt := range invalid {
		attempt := make(TileSet)
		attempt.union(s)
		delete(attempt, elt)
		attemptScorable := attempt.extractScorable()
		attemptScore := attemptScorable.score()
		if attemptScore > bestScore {
			bestScore = attemptScore
			bestChoice = attemptScorable
			bestChoice.invalid[elt] = s[elt]
		}
	}
	return bestChoice
}

// compareTileValues will return false if board is not a subset of tiles served.
// Takes in the tiles that have been served and the tiles on this board.
func compareTileValues(sent []Tile, received TileSet) bool {
	tileCount := make(map[string]int)
	for _, tile := range sent {
		tileCount[tile.Value] += 1
	}
	for _, tile := range received {
		tileCount[tile.Value] -= 1
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
func (b Board) scoreBoard(tilesServed []Tile) *Score {
	boardSet := b.setOfAllTiles()
	fmt.Println("Tiles received:", boardSet)
	fmt.Println("Tiles served:", tilesServed)
	result := &Score{
		Win: true,
		Msg: msg.Score,
	}

	// Calculate score if board is empty.
	maxPts := 0
	for _, elt := range tilesServed {
		maxPts += elt.Points
	}

	// Find the best scoring component.
	best := boardSet.extractScorable()
	result.Pts = maxPts - best.score()
	result.Invalid = best.invalid
	result.Valid = best.valid
	result.Unconnected = best.unconnected
	result.Words = best.words
	result.Nonwords = best.nonwords

	// A winning board has a score of 0.
	if result.Pts != 0 {
		result.Win = false
		if result.Pts < 0 {
			log.Println("Impossible score: cheating suspected!")
			result.Msg = msg.Error
			result.Pts = maxPts
		}
	}
	// A winning board must contain all the tiles served and no more.
	if len(boardSet) != len(tilesServed) {
		result.Win = false
		if len(boardSet) > len(tilesServed) {
			log.Println("Impossible number of tiles: cheating suspected!")
			result.Msg = msg.Error
			result.Pts = maxPts
			return result
		}
	}
	// A winning board must contain exactly the tiles served.
	if !compareTileValues(tilesServed, boardSet) {
		result.Win = false
		log.Println("Impossibly mismatched tiles: cheating suspected!")
		result.Msg = msg.Error
		result.Pts = maxPts
		return result
	}
	fmt.Println(result)
	return result
}

// Printable board.
func (b Board) String() string {
	result := ""
	for j := 0; j < len(b); j++ {
		for i := 0; i < len(b[0]); i++ {
			if b[j][i] == nil {
				result += " "
			} else {
				result += b[j][i].Value
			}
		}
		result += "\n"
	}
	return result
}
