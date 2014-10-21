function GameManager(size, startTiles) {
  gamemanager = this;
  this.size = size;
  this.startTiles = startTiles;
  this.cellSize = 48;
  this.div = document.getElementById('gameboard');

  this.setup();
}

// Setup game.
GameManager.prototype.setup = function() {
  this.tray = new Tray(this);
  this.borderDiv = createDiv("border", this.div);
  resizeDiv(this.borderDiv, vec(1, this.cellSize/2));
  this.grid = new Grid(this, vec(this.size, this.size));
  this.addButtonHandlers(this);

  // Create ghost tile and hide it.
  this.ghostTile = new Tile("", this, this.tray);
  removeDiv(this.ghostTile.div);
  this.ghostTile.div = createDiv("ghost", this.div);
  resizeDiv(this.ghostTile.div, vec(this.cellSize, this.cellSize));
  hideDiv(this.ghostTile.div);
  console.log("Finished setting up game!");

  // Join game and End game overlay divs.
  this.joinGameDiv = createOverlayDiv(document.body);
  this.populateJoinGameDiv();
  this.endGameDiv = createOverlayDiv(document.body);
}

GameManager.prototype.populateJoinGameDiv = function() {
  var button = createButton("Join Game", this.joinGameDiv)
  button.onclick = function() {
    hideDiv(gamemanager.joinGameDiv);
  };
  showDiv(gamemanager.joinGameDiv);
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
  console.log("Requesting new tiles!");
  websocketAlert("start");
}

// Send request for new letter and add it.
GameManager.prototype.requestNewTile = function() {
  websocketAlert("addTile");
}

// Remove all tiles and start a new game.
GameManager.prototype.reload = function() {
  this.grid.removeAllTiles();
  this.tray.removeAllTiles();
  gamemanager.requestTiles();
}

// Join a game.
GameManager.prototype.joinGame = function() {
  console.log("Join game!");
  websocketRequest("joinGame", function() {});
}

GameManager.prototype.verify = function() {
  var tileValues = this.stringifyTiles();
  websocketSendAndGet("verify", tileValues, function() {})

}

// Send gameboard for verification.
GameManager.prototype.stringifyTiles = function() {
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
  return tileValues;
}

// Take action when a score is received (game over).
GameManager.prototype.displayScore = function(score) {
  if (score["Valid"]) {
    setMessages("You win!");
  } else {
    setMessages("Board is incomplete: score of " + score["Score"]);
  }
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
  document.getElementById("joinGame").onclick = function() {
    gamemanager.joinGame();
  };
  document.getElementById("verify").onclick = function() {
    gamemanager.verify();
  };
}

function websocketHandleMessage(type, data, responseFunc) {
  if (type == "ok" || type == "error" || type == "fail") {
    responseFunc(type, data);
    return;
  }
  switch(type) {
    case "start":
      gamemanager.grid.removeAllTiles();
      gamemanager.tray.removeAllTiles();
      gamemanager.addNewLetters(data);
      break;
    case "addTile":
      gamemanager.addNewLetters([data]);
      break;
    case "score":
      gamemanager.displayScore(data);
      break;
    case "sendBoard":
      var tileValues = gamemanager.stringifyTiles();
      websocketSendAndGet("sendBoard", tileValues, function() {
        //TODO: handle error
      });
      break;
  }
}

function newGameManager() {
  gamemanager = new GameManager(12, 12);
}

function setMessages(text) {
    document.getElementById("messages").innerHTML = text;
}

window.requestAnimationFrame(function() {
  console.log("Starting Game!")
  socket = websocketCreate();
});
