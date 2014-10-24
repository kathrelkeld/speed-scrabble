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

// Create a div the size of the window and hide it.
function createOverlayDiv(id, parentDiv) {
  overlayDiv = createDiv("overlay", parentDiv);
  overlayDiv.id = id;
  moveDiv(overlayDiv, vec(0, 0));
  sizeToWindow(overlayDiv);
  overlayDiv.classList.add('hidden');
  return overlayDiv;
}

function createButton(id, text, parentDiv) {
  var button = document.createElement("button");
  var textNode = document.createTextNode(text);
  button.appendChild(textNode);
  button.id = id;
  if (parentDiv) {
    parentDiv.appendChild(button);
  }
  return button;
}

function createInput(id, defaultValue, parentDiv) {
  var input = document.createElement("input");
  input.type = "text";
  input.id = id;
	input.defaultValue = defaultValue;
  if (parentDiv) {
    parentDiv.appendChild(input);
  }
  return input;
}

function hideDiv(div) {
  div.classList.add('hidden');
}

function showDiv(div) {
  div.classList.remove('hidden');
}

// Move the given div to the given position.
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

// Resize the given div to the window's dimensions.
function sizeToWindow(div) {
  div.style.width = window.innerWidth;
  div.style.height = window.innerHeight;
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

function websocketCreate() {
  var loc = window.location;
  var wsPrefix = "ws://";
  if (loc.protocol === "https:") {
    wsPrefix = "wss://";
  }
  var socket = new WebSocket(wsPrefix + loc.host + "/connect");
  socket.onopen = function (e) {
    websocketRequest("connect", newGameManager);
  }
  return socket
}

function websocketGet(e, handler) {
  var m = JSON.parse(e.data);
  console.log("Receiving:", m['Type']);
  websocketHandleMessage(m["Type"], m["Data"], handler);
}

// Takes a type, data, and response handler that accepts type and data.
function websocketSendAndGet(type, data, handler) {
  console.log("Sending:", type);
  var message = {
    type: type,
    //TODO: don't hardcode this...
    at: "2014-01-02T15:04:05Z",
    data: data
  };
  socket.send(JSON.stringify(message));
  socket.onmessage = function(e) {
    websocketGet(e, handler);
  }
}

function websocketRequest(type, handler) {
  websocketSendAndGet(type, "", handler)
}

function websocketAlert(type) {
  websocketSendAndGet(type, "", function(type, data) {})
}
