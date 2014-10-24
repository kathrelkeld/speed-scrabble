function Tile(tileStruct, game, grid) {
  this.game = game;
  this.grid = grid;
  this.position = vec(0, 0);
  this.value = tileStruct["Value"];
  this.points = tileStruct["Points"];
  this.valid = true;
  this.div = createDiv('tile', grid.div);

  // Customize tile div.
  this.div.innerHTML = this.value;
  this.div.addEventListener('mousedown', this.mouseDown.bind(this));
  resizeDiv(this.div, vec(game.cellSize, game.cellSize));
  this.div.style.lineHeight = game.cellSize + "px";

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

Tile.prototype.invalidate = function() {
  if (this.valid == false) {
    return;
  }
  this.valid = false;
  this.div.classList.add("invalid");
}

Tile.prototype.validate = function() {
  if (this.valid == true) {
    return;
  }
  this.valid = true;
  this.div.classList.remove("invalid");
}

// onMouseDown event handler.  Always active for this tile.
Tile.prototype.mouseDown = function(e) {
  gamemanager.grid.validateTiles();

  this.remove();
  this.div.classList.add('moving');

  this.game.grid.removeHighlight();
  this.game.moveGhostTo(this.game.ghostTile, getDivCenter(this.div));
  showDiv(this.game.ghostTile.div);

  window.addEventListener('mousemove', this.moveListener);
  window.addEventListener('mouseup', this.upListener);
}

// onMouseMove event handler.  Active only during mouse move.
Tile.prototype.mouseMove = function(e) {
  this.game.ghostTile.remove();
  this.game.moveGhostTo(this.game.ghostTile, getDivCenter(this.div));

  var curr = getDivPos(this.div);
  moveDiv(this.div, vAdd(curr, vec(e.movementX, e.movementY)));
}

// onMouseUp event handler.  Active only during mouse move.
Tile.prototype.mouseUp = function(e) {
  // Remove further eventListeners.
  window.removeEventListener('mousemove', this.moveListener);
  window.removeEventListener('mouseup', this.upListener);

  // Dispose of ghost tile.
  this.game.ghostTile.remove();
  hideDiv(this.game.ghostTile.div);
  this.div.classList.remove('moving');

  // Try to place tile where the center of this tile is.
  this.game.moveTileTo(this, getDivCenter(this.div));
}
