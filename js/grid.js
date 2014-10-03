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

// Insert a tile at a given position.
Grid.prototype.insertTile = function(tile, position) {
    this.tiles[position.x][position.y] = tile;
    tile.position = position;
    tile.grid = this;
    var gridCorner = getDivPos(this.div);
    var newPosition = vAdd(gridCorner, vScale(position, this.game.cellSize));
    moveDiv(tile.div, newPosition);
    this.centerTiles();
}

// Shift entire board in one direction.
Grid.prototype.shiftTiles = function(direction) {
    var gridCorner = getDivPos(this.div);
    if (direction == 'r' || direction == 'l') {
        var emptyRow = [];
        for (var i=0; i < this.size.y; i++) {
            emptyRow.push(null);
        }
        if (direction == 'r') {
            this.tiles.pop();
            this.tiles.unshift(emptyRow);
        } else {
            this.tiles.shift();
            this.tiles.push(emptyRow);
        }
    }
    for (var i=0; i < this.size.x; i++) {
        if (direction == 'u') {
            this.tiles[i].shift();
            this.tiles[i].push(null);
        } else if (direction == 'd') {
            this.tiles[i].pop();
            this.tiles[i].unshift(null);
        }
        for (var j=0; j < this.size.y; j++) {
            var tile = this.tiles[i][j];
            if (tile) {
                console.log('moving tile', tile.value);
                tile.position = vec(i, j);
                var newPosition = vAdd(gridCorner,
                        vScale(tile.position, this.game.cellSize));
                moveDiv(tile.div, newPosition);
            }
        }
    }
}

// Recenter tiles if needed.
Grid.prototype.centerTiles = function() {
    if (!this.expands) {
        return;
    }
    var left = this.size.x - 1;
    var top = this.size.y - 1;
    var right = 0;
    var bottom = 0;
    for (var i=0; i < this.size.x; i++) {
        for (var j=0; j < this.size.y; j++) {
            var tile = this.tiles[i][j];
            if (tile) {
                left = Math.min(left, tile.position.x);
                right = Math.max(right, tile.position.x);
                top = Math.min(top, tile.position.y);
                bottom = Math.max(bottom, tile.position.y);
            }
        }
    }
    if (left == 0) {
        this.shiftTiles('r');
    }
    if (right == this.size.x - 1) {
        this.shiftTiles('l');
    }
    if (top == 0) {
        this.shiftTiles('d');
    }
    if (bottom == this.size.y - 1) {
        this.shiftTiles('u');
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
        // Return tile to former location.
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
                return vec(i, j);
            }
        }
    }
}


