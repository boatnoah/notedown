const socket = new WebSocket("ws://localhost:3000/ws");
let editor = document.getElementById("editor");
let localState = [];
let globalIndex = 0;

editor.addEventListener("input", (event) => {
  let operation = {
    value: event.data,
    action: "INSERT",
    indexPosition: globalIndex,
  };

  if (event.data == "null") {
    operation.action = "DELETE";
  }

  globalIndex++;

  socket.send(JSON.stringify(operation));
});

socket.onopen = () => {
  console.log("Connected to the WebSocket server.");
  editor.value = "";
};

socket.onmessage = (event) => {
  const operation = JSON.parse(event.data);
  console.log(operation);

  if (operation.action === "DELETE") {
    localState.splice(operation.indexPosition, 1);
    readLocalState();
    return;
  }

  localState.push(operation.value);
  console.log(localState);
  readLocalState();
};

socket.onclose = () => {
  console.log("WebSocket connection closed.");
};

socket.onerror = (error) => {
  console.error("WebSocket error:", error);
};

function placeCursorAtEnd() {
  const range = document.createRange();
  const selection = window.getSelection();
  range.selectNodeContents(editor);
  range.collapse(false);
  selection.removeAllRanges();
  selection.addRange(range);
}

function readLocalState() {
  editor.innerHTML = localState.join("");
  placeCursorAtEnd();
}
