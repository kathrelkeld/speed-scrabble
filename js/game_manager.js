function GameManager(size, startTiles) {
    this.size = size;
    this.startTiles = startTiles;

    this.setup()
}

// Setup game.
GameManager.prototype.setup = function() {
    this.grid = new Grid(this.size);
    this.addTiles();
}

// Add initial tiles.
GameManager.prototype.addTiles = function() {
    var tile = new Tile({x:0, y:0}, "A");
    this.grid.insertTile(tile);
}



window.requestAnimationFrame(function() {
    console.log("Starting Game!")
    new GameManager(12, 4);
});
