function GameManager(size, startTiles) {
  gamemanager = this;
  this.size = size;
  this.startTiles = startTiles;
  this.cellSize = 64;
  this.div = document.getElementById('gameboard');

  this.setup()
}

// Setup game.
GameManager.prototype.setup = function() {
  this.tray = new Grid(this, vec(this.startTiles,1), false);
  this.borderDiv = createDiv("border", this.div);
  resizeDiv(this.borderDiv, vec(1, this.cellSize/2));
  this.grid = new Grid(this, vec(this.size, this.size), true);
  this.addButtonHandlers(this);
  this.requestTiles();
}

// Add a list of tiles.
GameManager.prototype.addTiles = function(tiles) {
  for (var i=0; i<tiles.length; i++) {
    var tile = new Tile(tiles[i], this, this.tray);
    this.tray.addToFirstEmptyCell(tile);
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

// Send request for new Tiles and add them.
GameManager.prototype.requestTiles = function() {
  getJSON("/tiles", function(tiles) {
    gamemanager.addTiles(tiles);
  });
}

// Send request for new tile and add it.
GameManager.prototype.requestNewTile = function() {
  getJSON("/add_tile", function(tile) {
    gamemanager.addTiles([tile]);
  });
}

// Remove game divs from board.
GameManager.prototype.reload = function() {
  console.log("Reloading game board!");
  this.grid.removeAllTiles();
  this.tray.removeAllTiles();
  this.requestTiles()
}

// Add onclick handlers to the various game buttons
GameManager.prototype.addButtonHandlers = function() {
  document.getElementById("reset").onclick = function() {
    gamemanager.grid.sendAllTiles(gamemanager.tray)
  };
  document.getElementById("add_tile").onclick = function() {
    gamemanager.requestNewTile()
  };
  document.getElementById("reload").onclick = function() {
    gamemanager.reload();
  };
  var directions = ["left", "right", "down", "up"];
  directions.forEach(function(entry) {
    document.getElementById(entry).onclick = function() {
        gamemanager.grid.shiftTiles(entry);
    };
  });
}

function newGameManager() {
  new GameManager(12, 12);
}

window.requestAnimationFrame(function() {
  console.log("Starting Game!")
  newGameManager()
});
