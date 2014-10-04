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

// Resize the given div to the given dimensions.
function resizeDiv(div, dimensions) {
  div.style.width = dimensions.x;
  div.style.height = dimensions.y;
}

// Get the position of the given div.
function getDivPos(div) {
  return vec(div.offsetLeft, div.offsetTop);
}

// Remove a div from its parent.
function removeDiv(div) {
  div.parentNode.removeChild(div);
}