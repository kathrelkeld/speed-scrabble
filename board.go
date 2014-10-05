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
					c.add(elt)
				}
			}
		}
	}
	log.Println("Found", len(result), "components")
	log.Println(result)
	return result
}

// Verify that the given component list has valid words.
func (b Board) verifyComponent(comp map[Vec]struct{}) bool {
	// Component must be 2 or more tiles.
	if len(comp) <= 1 {
		return false
	}
	//across := make(map[Vec]bool)
	//down := make(map[Vec]bool)
	for v := range comp {
		log.Println(v)
	}
	return true
}

func verifyBoard(board [][]string) bool {
	Board(board).findComponents()
	maxX := len(board)
	maxY := len(board[0])
	var i, j, k int
	for i = 0; i < maxX; i++ {
		for j = 0; j < maxY; j++ {
			if board[i][j] != "" {
				if i == 0 || board[i-1][j] == "" {
					if !(i == maxX || board[i+1][j] == "") {
						var word string = board[i][j] + board[i+1][j]
						for k = i + 2; k < maxX; k++ {
							if board[k][j] != "" {
								word += board[k][j]
							} else {
								break
							}
						}
						if !verifyWord(word) {
							return false
						}
					}
				}
				if j == 0 || board[i][j-1] == "" {
					if !(j == maxY || board[i][j+1] == "") {
						var word string = board[i][j] + board[i][j+1]
						for k = j + 2; k < maxY; k++ {
							if board[i][k] != "" {
								word += board[i][k]
							} else {
								break
							}
						}
						if !verifyWord(word) {
							return false
						}
					}
				}
			}
		}
	}
	return true
}
