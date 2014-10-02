function Grid(game, size) {
    this.game = game;
    this.size = size; // (x, y) of grid dimensions.
    this.div = createDiv('grid', game.div);
    this.cells = this.createEmptyCells();

    resizeDiv(this.div, vScale(size, game.cellSize));
}

// Create an x by y grid of empty cells.
Grid.prototype.createEmptyCells = function() {
    var cells = [];
    for (var x = 0; x < this.size.x; x++) {
        cells.push([]);
        for (var y = 0; y < this.size.y; y++) {
            cells[x].push(new Cell(this, vec(x, y)));
        }
    }
    return cells;
}

// Insert a tile at a given position.
Grid.prototype.insertTile = function(tile, position) {
    tile.grid.cells[tile.position.x][tile.position.y].tile = null;
    this.cells[position.x][position.y].tile = tile;
    tile.position = position;
    tile.grid = this;
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
// Return the cell or null.
Grid.prototype.addToNearestEmptyCell = function(tile, position) {
    if (!this.isWithinBounds(position)) {
        return null;
    }
    // Calculate nearest neighbor.
    var relative = vSub(position, getDivPos(this.div));
    var coords = vScale(relative, 1 / this.game.cellSize);
    var intCoords = vec(Math.floor(coords.x), Math.floor(coords.y));
    var nearest = this.cells[intCoords.x][intCoords.y];

    // Make sure nearest is empty.
    if (!nearest.tile) {
        this.insertTile(tile, intCoords);
        return this.cells[intCoords.x][intCoords.y];
    } else {
        return null;
    }
}

Grid.prototype.addToFirstEmptyCell = function(tile) {
    for (var i = 0; i < this.size.x; i++) {
        for (var j = 0; j < this.size.y; j++) {
            var cell = this.cells[i][j];
            if (!cell.tile) {
                console.log('found first empty');
                this.insertTile(tile, vec(i, j));
                return cell;
            }
        }
    }
}


function Cell(grid, position) {
    this.grid = grid;
    this.position = position;
    this.div = createDiv('cell', grid.div);
    this.tile = null;

    // Customize cell div.
    cellSize = grid.game.cellSize;
    var gridCorner = getDivPos(grid.div);
    var newPosition = vAdd(gridCorner, vScale(this.position, cellSize));
    moveDiv(this.div, newPosition);
    resizeDiv(this.div, vec(cellSize, cellSize));
}

