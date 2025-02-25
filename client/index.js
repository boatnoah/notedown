const socket = new WebSocket("ws://localhost:3000/ws");
let editor = document.getElementById("editor");
let localState = [];

editor.addEventListener("input", (event) => {
  socket.send(event.data);
});

socket.onopen = () => {
  console.log("Connected to the WebSocket server.");
  editor.value = "";
};

socket.onmessage = (event) => {
  localState.push(event.data);
  newChanges = true;
  console.log(localState);
  main();
};

socket.onclose = () => {
  console.log("WebSocket connection closed.");
};

socket.onerror = (error) => {
  console.error("WebSocket error:", error);
};

function readLocalState() {
  editor.innerHTML = "";
  localState.forEach((char) => {
    editor.innerHTML += char;
  });
  newChanges = false;
}

function main() {
  if (newChanges) {
    readLocalState();
  }
}
