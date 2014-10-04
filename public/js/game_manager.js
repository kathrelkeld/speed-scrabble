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
  this.requestTiles();
}

// Add initial tiles.
GameManager.prototype.addStartTiles = function(tiles) {
  for (var i=0; i<tiles.length; i++) {
    var tile = new Tile(tiles[i], this, this.tray);
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

// Send request for tiles and add them.
GameManager.prototype.requestTiles = function() {

  var xmlhttp = new XMLHttpRequest();
  xmlhttp.onreadystatechange = function() {
    if (xmlhttp.readyState == 4 && xmlhttp.status == 200){
      tiles = JSON.parse(xmlhttp.responseText);
      gameManager.addStartTiles(tiles)
    }
  }
  xmlhttp.open("GET", "/tiles", true);
  xmlhttp.send();
}

window.requestAnimationFrame(function() {
  console.log("Starting Game!")
  gameManager = new GameManager(12, 12);
});
