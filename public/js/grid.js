function Grid(game, size) {
  this.game = game;
  this.size = size; // (x, y) of grid dimensions.
  this.div = createDiv('grid', game.div);

  // Customize grid div.
  resizeDiv(this.div, vScale(size, game.cellSize));

  this.setup();
}

// Create a cell div and move it to the right place.
Grid.prototype.createCellDiv = function(position) {
  var div = createDiv('cell', this.div);
  var cellSize = this.game.cellSize;
  moveTileDivRelative(div, vScale(position, cellSize), this.div);
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

// Get the value of the given position in the tile matrix.
Grid.prototype.getPosition = function(position) {
  return this.tiles[position.x][position.y];
}

// Set the value of the given position in the tile matrix.
Grid.prototype.setPosition = function(position, value) {
  this.tiles[position.x][position.y] = value;
}

Grid.prototype.removeTile = function(tile) {
  this.tiles[tile.position.x][tile.position.y] = null;
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
Grid.prototype.addTile = function(tile, position) {
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

// Find the nearest empty cell and return its index position.
Grid.prototype.findNearestEmptyCell = function(position) {
  if (!inBoundsOfDiv(position, this.div)) {
    return null;
  }
  // Calculate nearest neighbor.
  var intCoords = nearestCoordsInDiv(position, this.game.cellSize, this.div);

  //TODO: Find nearest empty tile instead of just returning null if not empty.
  if (this.getPosition(intCoords) == null) {
    return intCoords;
  }
  return null;
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


function Tray(game) {
  this.game = game;
  this.size = vec(1, 1); // (x, y) of dimensions.
  this.div = createDiv('grid', game.div);

  // Customize this div.
  resizeDiv(this.div, vScale(this.size, game.cellSize));

  this.tiles = [];
}

// Add a tile, at the index (integer) if given.
Tray.prototype.addTile = function(tile, index) {
  if (index == null) {
    index = this.tiles.length;
  }
  for (var i = index; i < this.tiles.length; i++) {
    var curr = this.tiles[i];
    curr.position = vec(i + 1, 0);
    curr.redraw();
  }
  this.tiles.splice(index, 0, tile);
  tile.grid = this;
  tile.position = vec(index, 0);
  tile.redraw();
};

Tray.prototype.removeTile = function(tile) {
  var index = this.tiles.indexOf(tile);
  for (var i = index + 1; i < this.tiles.length; i++) {
    var curr = this.tiles[i];
    curr.position = vAdd(curr.position, vec(-1, 0));
    curr.redraw();
  }
  this.tiles.splice(index, 1);
};

Tray.prototype.findNearest = function(position) {
  var nearest = nearestCoordsInDiv(position, this.game.cellSize, this.div);
  if (nearest.y != 0 || nearest.x >= this.tiles.length || nearest.x < 0) {
    return null;
  }
  return nearest.x;
};
