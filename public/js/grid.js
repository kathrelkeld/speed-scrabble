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

Grid.prototype.setup = function() {
  // Create x by y grids of cell divs and of no tiles.
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

  // Add event for keypresses
  this.div.addEventListener('mousedown', this.mouseDown.bind(this));
  window.addEventListener('keydown', this.keypress.bind(this));
  this.auraPos = vec(0, 0);
  this.auraDirection = "right";
}

Grid.prototype.mouseDown = function(e) {
  var intCoords = nearestCoordsInDiv(e, this.game.cellSize, this.div);
  this.addHighlight(intCoords);
}

Grid.prototype.addHighlight = function(position) {
  this.removeHighlight();
  this.auraPos = position;
  this.auraDiv = this.cellDivs[position.x][position.y];
  this.auraDiv.classList.add('highlighted');
}

Grid.prototype.removeHighlight = function() {
  if (this.auraDiv != null) {
    this.auraDiv.classList.remove('highlighted');
    this.auraDiv = null;
  }
}

// Take a keyboard event and move the highlight in the given direction.
Grid.prototype.moveHighlight = function(direction) {
  if (direction == "left" && this.auraPos.x > 0) {
    this.addHighlight(vAdd(vec(-1, 0), this.auraPos));
  } else if (direction == "right" && this.auraPos.x < this.size.x -1) {
    this.addHighlight(vAdd(vec(1, 0), this.auraPos));
  } else if (direction == "up" && this.auraPos.y > 0) {
    this.addHighlight(vAdd(vec(0, -1), this.auraPos));
  } else if (direction == "down" && this.auraPos.y < this.size.y -1) {
    this.addHighlight(vAdd(vec(0, 1), this.auraPos));
  }
}

Grid.prototype.removeHighlightTile = function() {
  var curr = this.getPosition(this.auraPos);
  if (curr != null) {
    curr.remove();
    this.game.moveTileToTray(curr);
  }
}

Grid.prototype.keypress = function(e) {
  switch(e.keyCode) {
    case 13: // enter
      e.preventDefault();
      if (this.auraDirection == "down") {
        this.moveHighlight("down");
      } else {
        setMessages("Text enter direction is: vertical");
        this.auraDirection = "down";
      }
      break;
    case 9: // tab
      e.preventDefault();
      if (this.auraDirection == "right") {
        this.moveHighlight("right");
      } else {
        setMessages("Text enter direction is: horizontal");
        this.auraDirection = "right";
      }
      break;
    case 37: // left arrow
      e.preventDefault();
      this.moveHighlight("left");
      break;
    case 39: // right arrow
      e.preventDefault();
      this.moveHighlight("right");
      break;
    case 38: // up arrow
      e.preventDefault();
      this.moveHighlight("up");
      break;
    case 40: // down arrow
      e.preventDefault();
      this.moveHighlight("down");
      break;
    case 8: // backspace
      e.preventDefault();
      this.removeHighlightTile();
    default:
      if (this.auraDiv != null) {
          var key = String.fromCharCode(e.keyCode).toUpperCase();
          var tile = this.game.tray.findByValue(key);
          if (tile != null) {
            this.removeHighlightTile();
            tile.remove();
            this.addTile(tile, this.auraPos);
            this.moveHighlight(this.auraDirection);
          }
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
  this.setPosition(tile.position, null);
}

// Destroy all current tiles.
Grid.prototype.removeAllTiles = function() {
  for (var x = 0; x < this.size.x; x++) {
    for (var y = 0; y < this.size.y; y++) {
      var tile = this.tiles[x][y];
      if (tile) {
        tile.remove();
        removeDiv(tile.div);
      }
    }
  }
}

// Move all tiles to the tray. 
Grid.prototype.sendAllTilesToTray = function(grid) {
  for (var x = 0; x < this.size.x; x++) {
    for (var y = 0; y < this.size.y; y++) {
      var tile = this.tiles[x][y];
      if (tile) {
        tile.remove();
        this.game.moveTileToTray(tile);
      }
    }
  }
}

// Draw the ghost tile.
Grid.prototype.addGhost = function(tile, position) {
  tile.position = position;
  tile.grid = this;
  tile.redraw();
}

// Insert a tile at a given position.
Grid.prototype.addTile = function(tile, position) {
  this.tiles[position.x][position.y] = tile;
  tile.position = position;
  tile.grid = this;
  tile.redraw();
  if (tile.position.x == 0) {
    this.shiftTiles("right");
  } else if (tile.position.x == this.size.x - 1) {
    this.shiftTiles("left");
  }
  if (tile.position.y == 0) {
    this.shiftTiles("down");
  } else if (tile.position.y == this.size.y - 1) {
    this.shiftTiles("up");
  }
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
        this.game.tray.moveTileToTray(removed);
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
        this.game.tray.moveTileToTray(removed);
      }
    }
    for (var j=0; j < this.size.y; j++) {
      var tile = this.tiles[i][j];
      if (tile) {
        tile.position = vec(i, j);
        tile.redraw();
      }
    }
  }
}

// Find the nearest empty cell and return its index position.
Grid.prototype.findNearestEmptyCell = function(position) {
  if (!inBoundsOfDiv(position, this.div)) {
    return null;
  }
  // Calculate the grid cell under this position.
  var intCoords = nearestCoordsInDiv(position, this.game.cellSize, this.div);

  // Check this and immediate surrounding tiles for an empty.
  dirs = [vec(0, 0), vec(1, 0), vec(-1, 0), vec(0, 1), vec(0, -1)];
  for (var d = 0; d < dirs.length; d++) {
    var curr = vAdd(intCoords, dirs[d]);
    if (inBoundsOfSize(curr, this.size) && this.getPosition(curr) == null) {
      return curr;
    }
  }
  return null;
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
  this.div = createDiv('grid', game.div);

  // Customize this div.
  resizeDiv(this.div, vec(1, game.cellSize));

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

Tray.prototype.addGhost = function(tile, index) {
  this.addTile(tile, index);
}

Tray.prototype.removeTile = function(tile) {
  var index = this.tiles.indexOf(tile);
  for (var i = index + 1; i < this.tiles.length; i++) {
    var curr = this.tiles[i];
    curr.position = vAdd(curr.position, vec(-1, 0));
    curr.redraw();
  }
  this.tiles.splice(index, 1);
};

Tray.prototype.removeAllTiles = function() {
  for (var i = 0; i < this.tiles.length; i++) {
      var tile = this.tiles[i];
      removeDiv(tile.div);
  }
  this.tiles = [];
}

Tray.prototype.findNearest = function(position) {
  var nearest = nearestCoordsInDiv(position, this.game.cellSize, this.div);
  if (nearest.y != 0 || nearest.x >= this.tiles.length || nearest.x < 0) {
    return null;
  }
  return nearest.x;
};

Tray.prototype.findByValue = function(value) {
  for (var i = 0; i < this.tiles.length; i++) {
      var tile = this.tiles[i];
      if (tile.value == value) {
        return tile;
      }
  }
  return null;
}
