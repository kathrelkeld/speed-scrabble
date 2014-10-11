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
  this.tray = new Tray(this);
  this.borderDiv = createDiv("border", this.div);
  resizeDiv(this.borderDiv, vec(1, this.cellSize/2));
  this.grid = new Grid(this, vec(this.size, this.size));
  this.addButtonHandlers(this);
  this.requestTiles();

  // Create ghost tile and hide it.
  this.ghostTile = new Tile("", this, this.tray);
  removeDiv(this.ghostTile.div);
  this.ghostTile.div = createDiv("ghost", this.div);
  resizeDiv(this.ghostTile.div, vec(this.cellSize, this.cellSize));
  this.ghostTile.div.classList.add('hidden');
}

// Add a list of tiles.
GameManager.prototype.addNewLetters = function(letters) {
  for (var i=0; i<letters.length; i++) {
    var tile = new Tile(letters[i], this, this.tray);
    this.moveTileToTray(tile);
  }
}

// Figure out where a tile will land, given its position.
// Returns the grid and the position in the grid.
GameManager.prototype.getNextPosition = function(position) {
  // Try to add to the grid or the tray at this position.
  var nearestBoardCell = this.grid.findNearestEmptyCell(position);
  if (nearestBoardCell) {
    return [this.grid, nearestBoardCell];
  }
  return [this.tray, this.tray.findNearest(position)];
}

// Place a tile, given its position.
GameManager.prototype.moveTileTo = function(tile, position) {
  // Try to add to the grid or the tray at this position.
  var results = this.getNextPosition(position);
  results[0].addTile(tile, results[1]);
}

// Place a ghost, given its position.
GameManager.prototype.moveGhostTo = function(tile, position) {
  // Try to add to the grid or the tray at this position.
  var results = this.getNextPosition(position);
  results[0].addGhost(tile, results[1]);
}

// Place a tile back in the tray.
GameManager.prototype.moveTileToTray = function(tile) {
  // Try to add to the grid or the tray at this position.
  this.tray.addTile(tile, null);
}

// Send request for new Letters and add them.
GameManager.prototype.requestTiles = function() {
  getJSON("/tiles", function(letters) {
    gamemanager.addNewLetters(letters);
  });
}

// Send request for new letter and add it.
GameManager.prototype.requestNewTile = function() {
  getJSON("/add_tile", function(letter) {
    gamemanager.addNewLetters([letter]);
  });
}

// Remove game divs from board.
GameManager.prototype.reload = function() {
  console.log("Reloading game board!");
  this.grid.removeAllTiles();
  this.tray.removeAllTiles();
  this.requestTiles()
}

// Send gameboard for verification.
GameManager.prototype.verifyTiles = function() {
  var tileValues = [];
  for (var i = 0; i < this.grid.size.x; i++) {
      tileValues.push([]);
    for (var j = 0; j < this.grid.size.y; j++) {
      if (this.grid.tiles[i][j]) {
        tileValues[i].push(this.grid.tiles[i][j].value);
      } else {
        tileValues[i].push(null);
      }
    }
  }
  sendAndGetJSON(tileValues, "/verify", function(result) {
    console.log(result);
  });
}

// Add onclick handlers to the various game buttons
GameManager.prototype.addButtonHandlers = function() {
  document.getElementById("reset").onclick = function() {
    gamemanager.grid.sendAllTilesToTray()
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
  document.getElementById("verify").onclick = function() {
    gamemanager.verifyTiles();
  };
}

function newGameManager() {
  new GameManager(12, 12);
}

window.requestAnimationFrame(function() {
  console.log("Starting Game!")
  newGameManager()
});
