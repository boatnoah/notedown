import Operation from "./types";
import { v4 as uuidv4 } from "uuid";

const socket = new WebSocket("ws://localhost:3000/ws");
const clientID = uuidv4();
let editor = document.getElementById("editor");
let localState: Operation[] = [];
let cursorPos: Operation["charID"];

editor?.addEventListener("input", (event) => {
  let cursorPos = getCursorPosition() - 1; // minus is need because
  let fractionalPos: number;

  if (localState.length === 0) {
    fractionalPos = 0.5;
  } else if (cursorPos === 0) {
    fractionalPos = localState[0].position / 2; // Push closer to the start
  } else if (cursorPos >= localState.length) {
    fractionalPos = localState[localState.length - 1].position + 1; // Append to end
  } else {
    const prev = localState[cursorPos - 1].position;
    const next = localState[cursorPos].position;

    fractionalPos = (prev + next) / 2;
  }

  let operation = {
    clientID: clientID,
    charID: uuidv4(),
    value: event.data,
    action: "INSERT",
    position: fractionalPos,
  };

  console.log("Sending this operation to the server: ", operation);
  socket.send(JSON.stringify(operation));
});

socket.onopen = () => {
  console.log("Connected to the WebSocket server.");
};

socket.onmessage = (event) => {
  const operations = JSON.parse(event.data);
  console.log("received from server", operations);
  localState = operations;
  readLocalState();
};

socket.onclose = () => {
  console.log("WebSocket connection closed.");
};

socket.onerror = (error) => {
  console.error("WebSocket error:", error);
};

function getCursorPosition() {
  if (document.activeElement !== editor) {
    return 0;
  }
  const selection = window.getSelection();
  const range = selection.getRangeAt(0);
  const clonedRange = range.cloneRange();
  clonedRange.selectNodeContents(editor);
  clonedRange.setEnd(range.endContainer, range.endOffset);

  return clonedRange.toString().length;
}

function findOperation() {
  const currentState = editor.innerHTML;
  let i = 0;
  let j = 0;

  while (i < localState.length || j < currentState.length) {
    console.log("Comparing i: ", localState[i].value);
    console.log("Comparing j: ", currentState[j]);
    if (localState[i].value !== currentState[j]) {
      console.log("do i make it chat");
      return localState[i];
    }
    i = i < localState.length ? i + 1 : j + 0;
    j = j < currentState.length ? j + 1 : j + 0;
  }
  return localState[localState.length - 1];
}

function readLocalState() {
  for (let i = 0; i < localState.length; i++) {
    const operation = localState[i];
    const uuid = operation.charID;
    const textContent = document.createTextNode(operation.value);
    const span = document.createElement("span");
    span.appendChild(textContent);
    span.setAttribute("data", uuid);
    editor.appendChild(span);
  }
}
