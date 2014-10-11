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
}

// Add a list of tiles.
GameManager.prototype.addNewLetters = function(letters) {
  for (var i=0; i<letters.length; i++) {
    var tile = new Tile(letters[i], this, this.tray);
    this.tray.addTile(tile);
  }
}

// Figure out the place to put a tile in motion.
GameManager.prototype.moveTileTo = function(tile, position) {
  // Try to add to the grid or the tray at this position.
  if (this.grid.addToNearestEmptyCell(tile, position)) {
        return;
      }
  this.tray.addTile(tile);
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
