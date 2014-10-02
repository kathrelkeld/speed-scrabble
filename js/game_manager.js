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
    var tile = new Tile(vec(0, 0), "A", this, this.tray);
    this.tray.insertTile(tile, vec(0, 0));
    var tile = new Tile(vec(1, 0), "B", this, this.tray);
    this.tray.insertTile(tile, vec(1, 0));
    var tile = new Tile(vec(2, 0), "C", this, this.tray);
    this.tray.insertTile(tile, vec(2, 0));
}

// Figure out the place to put a tile in motion.
GameManager.prototype.moveTileTo = function(tile, position) {
    // Try to add to the grid.
    var newCell = this.grid.addToNearestEmptyCell(tile, position);
    if (newCell) {
        return newCell;
    }
    // Try to add to the tray.
    newCell = this.tray.addToNearestEmptyCell(tile, position);
    if (newCell) {
        return newCell;
    }
    // Default to first empty tile of tray.
    return this.tray.addToFirstEmptyCell(tile);
}


window.requestAnimationFrame(function() {
    console.log("Starting Game!")
    new GameManager(12, 4);
});
