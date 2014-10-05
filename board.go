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

func (a Vec) vAdd(b Vec) Vec {
	return Vec{a.x + b.x, a.y + b.y}
}

type VecSet map[Vec]struct{}

func (s VecSet) contains(elt Vec) bool {
	_, ok := s[elt]
	return ok
}

func (s VecSet) add(elt Vec) {
	s[elt] = struct{}{}
}

type Board [][]string

func (b Board) value(v Vec) string {
	return b[v.x][v.y]
}

// Find the tiles connected to the given tile.
func (b Board) findConnectedTiles(v Vec, found VecSet) {
	if !found.contains(v) &&
		v.x < len(b) && v.y < len(b[0]) &&
		v.x >= 0 && v.y >= 0 && b.value(v) != "" {
		found.add(v)
		b.findConnectedTiles(v.vAdd(Vec{1, 0}), found)
		b.findConnectedTiles(v.vAdd(Vec{0, 1}), found)
		b.findConnectedTiles(v.vAdd(Vec{-1, 0}), found)
		b.findConnectedTiles(v.vAdd(Vec{0, -1}), found)
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
					c.add(elt)
				}
			}
		}
	}
	log.Println("Found", len(result), "components")
	return result
}

// Given a tile and a direction, follow the word in that direction,
// adding each tile to the set of tiles checked in that direction.
func (b Board) followWord(v Vec, comp VecSet, dSet VecSet, d Vec) bool {
	dSet.add(v)
	next := v.vAdd(d)
	if !comp.contains(next) {
		return true
	}
	word := b.value(v)
	for comp.contains(next) {
		dSet.add(next)
		word += b.value(next)
		next = next.vAdd(d)
	}
	log.Println("Found word", word)
	return verifyWord(word)
}

// Verify that the given component list has valid words.
func (b Board) verifyComponent(comp VecSet) bool {
	// Component must be 2 or more tiles.
	if len(comp) <= 1 {
		return false
	}
	across := make(VecSet)
	vertical := make(VecSet)
	for v := range comp {
		if !comp.contains(v.vAdd(Vec{-1, 0})) {
			if !b.followWord(v, comp, across, Vec{1, 0}) {
				return false
			}
		}
		if !comp.contains(v.vAdd(Vec{0, -1})) {
			if !b.followWord(v, comp, vertical, Vec{0, 1}) {
				return false
			}
		}
	}
	return true
}

func (b Board) verifyBoard() bool {
	components := b.findComponents()
	valid := false
	for _, comp := range components {
		valid = b.verifyComponent(comp)
		log.Println("Component was", valid)
	}
	if len(components) != 1 {
		return false
	}
	return valid
}
