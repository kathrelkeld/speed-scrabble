function GameManager(size, startTiles) {
    this.size = size;
    this.startTiles = startTiles;
    this.cellSize = 64;
    this.div = document.getElementById('gameboard');

    this.setup()
}

// Setup game.
GameManager.prototype.setup = function() {
    this.tray = new Grid(this, vec(this.startTiles,1), false);
    this.grid = new Grid(this, vec(this.size, this.size), true);
    this.addStartTiles();
}

// Add initial tiles.
GameManager.prototype.addStartTiles = function() {
    var value = 'A';
    for (var i=0; i<12; i++) {
        var nextValue = String.fromCharCode(value.charCodeAt(0) + i)
        var tile = new Tile(nextValue, this, this.tray);
        this.tray.insertTile(tile, vec(i, 0));
    }
}

// Figure out the place to put a tile in motion.
GameManager.prototype.moveTileTo = function(tile, position) {
    // Try to add to the grid or the tray at this position.
    if (this.grid.addToNearestEmptyCell(tile, position) ||
        this.tray.addToNearestEmptyCell(tile, position)) {
        return;
    }
    this.tray.addToFirstEmptyCell(tile);
}


window.requestAnimationFrame(function() {
    console.log("Starting Game!")
    new GameManager(12, 12);
});
