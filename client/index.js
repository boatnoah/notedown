const socket = new WebSocket("ws://localhost:3000/ws");
const clientID = generateUUID();
let editor = document.getElementById("editor");
let localState = [];

editor.addEventListener("input", (event) => {
  if (event.inputType === "deleteContentBackward") {
    let deletedOperation = findOperation();
    deletedOperation.action = "DELETE";
    socket.send(JSON.stringify(deletedOperation));
    return;
  }

  let cursorPos = getCursorPosition() - 1; // minus is need because
  let fractionalPos;

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
    charID: generateUUID(),
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

function placeCursorAtEnd() {
  const range = document.createRange();
  const selection = window.getSelection();
  range.selectNodeContents(editor);
  range.collapse(false);
  selection.removeAllRanges();
  selection.addRange(range);
}

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

function generateUUID() {
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  array[6] = (array[6] & 0x0f) | 0x40; // Version 4
  array[8] = (array[8] & 0x3f) | 0x80; // Variant 10xx
  return [...array].map((byte) => byte.toString(16).padStart(2, "0")).join("");
}

function readLocalState() {
  let prevCursorPosition = getCursorPosition();

  const sizeOfCurrentState = editor.innerHTML.length;
  const sizeOfLocalState = localState.length;

  if (sizeOfCurrentState > sizeOfLocalState) {
    prevCursorPosition--;
  }

  let data = "";

  for (let i = 0; i < localState.length; i++) {
    data += localState[i].value;
  }

  editor.textContent = data;

  if (prevCursorPosition > 0) {
    //	place at the previous cursor position

    let setpos = document.createRange();

    // Creates object for selection
    let set = window.getSelection();

    setpos.setStart(editor.childNodes[0], prevCursorPosition);

    setpos.collapse(true);

    set.removeAllRanges();

    set.addRange(setpos);
    editor.focus();
  }
}
