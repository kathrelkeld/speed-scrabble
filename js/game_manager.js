function GameManager(size, startTiles) {
    this.size = size;
    this.startTiles = startTiles;
    this.cellSize = 64;
    this.div = document.getElementById('gameboard');

    this.setup()
}

// Setup game.
GameManager.prototype.setup = function() {
    this.tray = new Grid(this, { x: this.startTiles, y: 1 });
    this.tray.div.style.backgroundColor = 'red';
    this.grid = new Grid(this, { x: this.size, y: this.size });
    this.addStartTiles();
}

// Add initial tiles.
GameManager.prototype.addStartTiles = function() {
    var value = 'A';
    for (var i=0; i<4; i++) {
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
    //if (this.tray.addToNearestEmptyCell(tile, position)) {
        //return;
    //}
    this.tray.addToFirstEmptyCell(tile);
}


window.requestAnimationFrame(function() {
    console.log("Starting Game!")
    new GameManager(12, 4);
});
