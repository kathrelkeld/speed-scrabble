function Grid(size, width) {
    this.size = size;
    this.tileWidth = 100;
    this.cells = this.empty();
}

Grid.prototype.empty = function() {
    var cells = [];

    for (var x = 0; x < this.size; x++) {
        cells.push([])

        for (var y = 0; y < this.size; y++) {
            cells[x].push(null);
        }
    }

    return cells
}

// Insert a tile.
Grid.prototype.insertTile = function(tile) {
    this.cells[tile.x][tile.y] = tile;
}

// Remove a tile.
Grid.prototype.insertTile = function(tile) {
    this.cells[tile.x][tile.y] = null;
}

// Nearest tile.
Grid.prototype.nearestCell = function(position) {
    var max_position = this.size * this.tileWidth;
    if (position.x >= max_position || position.x < 0 ||
        position.y >= max_position || postiion.y < 0) {
        return null;
    }
    return position.x / this.tileWidth;
}

