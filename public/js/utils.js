// Return the vector (x, y).
function vec(x, y) {
  return { x: x, y: y };
}

// Add the vector a to the vector b.
function vAdd(a, b) {
  return vec(a.x + b.x, a.y + b.y);
}

// Subtract the vector b from the vector a.
function vSub(a, b) {
  return vec(a.x - b.x, a.y - b.y);
}

// Scale the vector a by the constant c.
function vScale(a, c) {
  return vec(a.x * c, a.y * c);
}

// Return whether the elements of vectors a and b are equivalent.
function vEquals(a, b) {
  return ((a.x == b.x) && (a.y == b.y));
}

// Whether a given vector is positive and less than size in all directions.
function inBoundsOfSize(v, size) {
  return (v.x >= 0 && v.y >= 0 && v.x < size.x && v.y < size.y);
}

// Create a div with the given className and append to given parentDiv.
function createDiv(className, parentDiv) {
  var div = document.createElement('div');
  div.className = className;
  if (parentDiv) {
    parentDiv.appendChild(div);
  }
  return div;
}

// Move the given div to the given position.
function moveDiv(div, position) {
  div.style.left = position.x;
  div.style.top = position.y;
}

// Move the given div to the given position relative to the given parent.
function moveTileDivRelative(div, position, pdiv) {
  var pCorner = getDivPos(pdiv);
  var newPosition = vAdd(pCorner, position);
  moveDiv(div, newPosition);
}

// Resize the given div to the given dimensions.
function resizeDiv(div, dimensions) {
  div.style.width = dimensions.x;
  div.style.height = dimensions.y;
}

// Get the top-left position of the given div.
function getDivPos(div) {
  return vec(div.offsetLeft, div.offsetTop);
}

// Get the bottom-right position of the given div.
function getDivPosBottomRight(div) {
  return vec(div.offsetLeft + div.offsetWidth,
             div.offsetTop + div.offsetHeight);
}

// Return the center of the given div.
function getDivCenter(div) {
  return vec(div.offsetLeft + div.offsetWidth/2,
             div.offsetTop + div.offsetHeight/2);
}

// Whether a given mouse position is in bounds of a given div.
function inBoundsOfDiv(position, div) {
  var topLeft = getDivPos(div);
  var bottomRight = getDivPosBottomRight(div);
  return position.x >= topLeft.x && position.x <= bottomRight.x &&
    position.y >= topLeft.y && position.y <= bottomRight.y;
}

// Calculate the nearest integer coordinates, given the cell size and parent div.
// Does not check whether position is actually within the given div's bounds.
function nearestCoordsInDiv(position, cellSize, div) {
  var relative = vSub(position, getDivPos(div));
  var coords = vScale(relative, 1 / cellSize);
  return vec(Math.floor(coords.x), Math.floor(coords.y));
}

// Remove a div from its parent.
function removeDiv(div) {
  div.parentNode.removeChild(div);
}

// Request json results from given page and apply given handler.
function getJSON(page, handler) {
  var xmlhttp = new XMLHttpRequest();
  xmlhttp.onreadystatechange = function() {
    if (xmlhttp.readyState == 4 && xmlhttp.status == 200){
      response = JSON.parse(xmlhttp.responseText);
      handler(response);
    }
  }
  xmlhttp.open("GET", page, true);
  xmlhttp.send();
}

// Send json results to given page and apply given handler.
function sendAndGetJSON(data, page, handler) {
  var xmlhttp = new XMLHttpRequest();
  xmlhttp.onreadystatechange = function() {
    if (xmlhttp.readyState == 4 && xmlhttp.status == 200){
      response = JSON.parse(xmlhttp.responseText);
      handler(response);
    }
  }
  xmlhttp.open("POST", page, true);
  xmlhttp.send(JSON.stringify(data));
}
