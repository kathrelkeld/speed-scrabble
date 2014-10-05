function Grid(game, size, expands) {
  this.game = game;
  this.size = size; // (x, y) of grid dimensions.
  this.div = createDiv('grid', game.div);
  this.expands = expands;

  // Customize grid div.
  resizeDiv(this.div, vScale(size, game.cellSize));

  this.setup();
}

// Create a cell div and move it to the right place.
Grid.prototype.createCellDiv = function(position) {
  var div = createDiv('cell', this.div);
  var cellSize = this.game.cellSize;
  var gridCorner = getDivPos(this.div);
  var newPosition = vAdd(gridCorner, vScale(position, cellSize));
  moveDiv(div, newPosition);
  resizeDiv(div, vec(cellSize, cellSize));
  return div;
}

// Create x by y grids of cell divs and of no tiles.
Grid.prototype.setup = function() {
  this.cellDivs = [];
  this.tiles = [];
  for (var x = 0; x < this.size.x; x++) {
    this.cellDivs.push([]);
    this.tiles.push([]);
    for (var y = 0; y < this.size.y; y++) {
      this.cellDivs[x].push(this.createCellDiv(vec(x, y)));
      this.tiles[x].push(null);
    }
  }
}

// Destroy all current tiles.
Grid.prototype.removeAllTiles = function() {
  for (var x = 0; x < this.size.x; x++) {
    for (var y = 0; y < this.size.y; y++) {
      var tile = this.tiles[x][y];
      if (tile) {
        removeDiv(tile.div);
        this.tiles[x][y] = null;
      }
    }
  }
}

// Move all tiles to the given grid.
Grid.prototype.sendAllTiles = function(grid) {
  for (var x = 0; x < this.size.x; x++) {
    for (var y = 0; y < this.size.y; y++) {
      var tile = this.tiles[x][y];
      if (tile) {
        this.tiles[x][y] = null;
        grid.addToFirstEmptyCell(tile);
      }
    }
  }
}

// Insert a tile at a given position.
Grid.prototype.insertTile = function(tile, position) {
  this.tiles[position.x][position.y] = tile;
  tile.position = position;
  tile.grid = this;
  var gridCorner = getDivPos(this.div);
  var newPosition = vAdd(gridCorner, vScale(position, this.game.cellSize));
  moveDiv(tile.div, newPosition);
}

// Shift entire board in one direction.
Grid.prototype.shiftTiles = function(direction) {
  var gridCorner = getDivPos(this.div);
  if (direction == "right" || direction == "left") {
    var emptyRow = [];
    for (var i=0; i < this.size.y; i++) {
      emptyRow.push(null);
    }
    if (direction == "right") {
      var removed = this.tiles.pop();
      this.tiles.unshift(emptyRow);
    } else {
      var removed = this.tiles.shift();
      this.tiles.push(emptyRow);
    }
    for (var i=0; i < removed.length; i++) {
      if (removed[i]) {
        console.log("returning", removed[i].value)
        this.game.tray.addToFirstEmptyCell(removed[i]);
      }
    }
  }
  for (var i=0; i < this.size.x; i++) {
    if (direction == "up") {
      var removed = this.tiles[i].shift();
      this.tiles[i].push(null);
    } else if (direction == "down") {
      var removed = this.tiles[i].pop();
      this.tiles[i].unshift(null);
    }
    if (direction == "up" || direction == "down") {
      if (removed) {
        this.game.tray.addToFirstEmptyCell(removed);
      }
    }
    for (var j=0; j < this.size.y; j++) {
      var tile = this.tiles[i][j];
      if (tile) {
        tile.position = vec(i, j);
        var newPosition = vAdd(gridCorner,
            vScale(tile.position, this.game.cellSize));
        moveDiv(tile.div, newPosition);
      }
    }
  }
}

// Returns true/false if the position is/isn't within the grid.
Grid.prototype.isWithinBounds = function(position) {
  var curr = getDivPos(this.div);
  return position.x >= curr.x &&
    position.x <= curr.x + this.size.x * this.game.cellSize &&
    position.y >= curr.y &&
    position.y <= curr.y + this.size.y * this.game.cellSize;
}

// Add the given tile to the nearest tile to position, if possible.
// If the tile is occupied, return tile to former spot.
// Return true or false for success or failure.
Grid.prototype.addToNearestEmptyCell = function(tile, position) {
  if (!this.isWithinBounds(position)) {
    return null;
  }
  // Calculate nearest neighbor.
  var relative = vSub(position, getDivPos(this.div));
  var coords = vScale(relative, 1 / this.game.cellSize);
  var intCoords = vec(Math.floor(coords.x), Math.floor(coords.y));

  // Make sure nearest is empty and not this tile.
  if (this.tiles[intCoords.x][intCoords.y] == null) {
    this.insertTile(tile, intCoords);
  } else {
    // Return tile to former location if this location isn't the tray.
    if (this == this.game.tray) {
      return false;
    }
    tile.grid.insertTile(tile, tile.position);
  }
  return true;
}

// Find the first available empty cell and add the tile to it.
Grid.prototype.addToFirstEmptyCell = function(tile) {
  for (var i = 0; i < this.size.x; i++) {
    for (var j = 0; j < this.size.y; j++) {
      if (!this.tiles[i][j]) {
        this.insertTile(tile, vec(i, j));
        return
      }
    }
  }
  // If grid is full, expand it.
  this.expand("right");
  this.insertTile(tile, vec(this.size.x - 1, 0));
}

// Expand grid in the given direction.
Grid.prototype.expand = function(direction) {
  if (direction == "right" || direction == "left") {
    var emptyDivRow = [];
    var emptyTileRow = [];
    for (var i=0; i < this.size.y; i++) {
      emptyDivRow.push(this.createCellDiv(vec(this.size.x, i)));
      emptyTileRow.push(null);
    }
    this.size.x += 1;
    this.cellDivs.push(emptyDivRow);
    this.tiles.push(emptyTileRow);
    if (direction == "left") {
      this.shiftTiles("right")
    }
  } else {
    for (var i=0; i < this.size.x; i++) {
      this.cellDivs.push(this.createCellDiv(vec(i, this.size.y)));
      this.tiles.push(null);
    }
    this.size.y += 1;
    if (direction == "up") {
      this.shiftTiles("down")
    }
  }
}
