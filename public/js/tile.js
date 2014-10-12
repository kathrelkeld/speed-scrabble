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

// Redraw a tile in its known position in its known grid.
Tile.prototype.redraw = function() {
  var scaledPosition = vScale(this.position, this.game.cellSize);
  moveTileDivRelative(this.div, scaledPosition, this.grid.div);
}

// Remove a given tile from its parent grid.
Tile.prototype.remove = function() {
  this.grid.removeTile(this);
}

// onMouseDown event handler.  Always active for this tile.
Tile.prototype.mouseDown = function(e) {
  this.game.grid.removeHighlight();
  this.remove();
  this.div.classList.add('moving');
  window.addEventListener('mousemove', this.moveListener);
  window.addEventListener('mouseup', this.upListener);
  this.game.ghostTile.div.classList.remove('hidden');
  this.game.moveGhostTo(this.game.ghostTile, getDivCenter(this.div));
}

// onMouseMove event handler.  Active only during mouse move.
Tile.prototype.mouseMove = function(e) {
  var curr = getDivPos(this.div);
  moveDiv(this.div, vAdd(curr, vec(e.movementX, e.movementY)));
  this.game.ghostTile.remove();
  this.game.moveGhostTo(this.game.ghostTile, getDivCenter(this.div));
}

// onMouseUp event handler.  Active only during mouse move.
Tile.prototype.mouseUp = function(e) {
  // Remove further eventListeners.
  window.removeEventListener('mousemove', this.moveListener);
  window.removeEventListener('mouseup', this.upListener);

  // Dispose of ghost tile.
  this.game.ghostTile.remove();
  this.game.ghostTile.div.classList.add('hidden');
  this.div.classList.remove('moving');

  // Try to place tile where the center of this tile is.
  this.game.moveTileTo(this, getDivCenter(this.div));
}
