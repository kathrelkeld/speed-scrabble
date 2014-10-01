function Tile(position, value) {
    this.x = position.x;
    this.y = position.y;
    this.tileWidth = 100;
    this.value = value;
    this.div = this.createDiv();

    this.moveListener = this.moveVisibleTile.bind(this);
    this.upListener = this.mouseUp.bind(this);
}

// Update position of tile.
Tile.prototype.updatePosition = function(position) {
    this.x = position.x;
    this.y = position.y;
}

// Create tile div.
Tile.prototype.createDiv = function() {
    var div = document.createElement('div');
    div.className = 'tile';
    div.innerHTML = this.value;
    div.addEventListener('mousedown', this.mouseDown.bind(this));
    document.getElementById('gameboard').appendChild(div);
    return div;
}

Tile.prototype.mouseDown = function(e) {
    window.addEventListener('mousemove', this.moveListener);
    window.addEventListener('mouseup', this.upListener);
}

// Move the visible position of the tile.
Tile.prototype.moveVisibleTile = function(e) {
    var rect = this.div.getBoundingClientRect();
    this.div.style.left = rect.left + e.movementX;
    this.div.style.top = rect.top + e.movementY;
}

Tile.prototype.mouseUp = function(e) {
    window.removeEventListener('mousemove', this.moveListener);
    window.removeEventListener('mouseup', this.upListener);
}
