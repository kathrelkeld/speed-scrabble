function Tile(value, game, grid) {
    this.game = game;
    this.grid = grid;
    this.position = vec(0, 0);
    this.value = value;
    this.div = createDiv('tile', grid.div);

    // Customize tile div.
    this.div.innerHTML = value;
    this.div.addEventListener('mousedown', this.mouseDown.bind(this));
    resizeDiv(this.div, vec(game.cellSize, game.cellSize));

    // Add values needed for mouse movement.
    this.moveListener = this.mouseMove.bind(this);
    this.upListener = this.mouseUp.bind(this);
}

// Remove a given tile from its parent grid.
Tile.prototype.removeTile = function() {
    this.grid.tiles[this.position.x][this.position.y] = null;
}

// onMouseDown event handler.  Always active for this tile.
Tile.prototype.mouseDown = function(e) {
    this.removeTile();
    window.addEventListener('mousemove', this.moveListener);
    window.addEventListener('mouseup', this.upListener);
}

// onMouseMove event handler.  Active only during mouse move.
Tile.prototype.mouseMove = function(e) {
    var curr = getDivPos(this.div);
    moveDiv(this.div, vAdd(curr, vec(e.movementX, e.movementY)));
}

// onMouseUp event handler.  Active only during mouse move.
Tile.prototype.mouseUp = function(e) {
    window.removeEventListener('mousemove', this.moveListener);
    window.removeEventListener('mouseup', this.upListener);
    var halfCellSize = this.game.cellSize / 2;
    var centerPos = vAdd(getDivPos(this.div), vec(halfCellSize, halfCellSize));
    var cell = this.game.moveTileTo(this, centerPos);
}
