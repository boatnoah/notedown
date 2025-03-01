const socket = new WebSocket("ws://localhost:3000/ws");
const clientID = generateUUID();
let editor = document.getElementById("editor");
let localState = [];

editor.addEventListener("input", (event) => {
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
    value: event.data,
    action: "INSERT",
    position: fractionalPos,
  };

  localState.push(operation);
  socket.send(JSON.stringify(operation));

  localState.sort((a, b) => a.position - b.position);

  readLocalState();

  console.log(localState);
});

socket.onopen = () => {
  console.log("Connected to the WebSocket server.");
};

socket.onmessage = (event) => {
  const operation = JSON.parse(event.data);

  console.log("Current operation: ", operation);

  if (operation.clientID !== clientID) {
    localState.push(operation);
    localState.sort((a, b) => a.position - b.position);
    readLocalState();
  }
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

function generateUUID() {
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  array[6] = (array[6] & 0x0f) | 0x40; // Version 4
  array[8] = (array[8] & 0x3f) | 0x80; // Variant 10xx
  return [...array].map((byte) => byte.toString(16).padStart(2, "0")).join("");
}

function readLocalState() {
  const prevCursorPosition = getCursorPosition();

  let data = "";

  for (let i = 0; i < localState.length; i++) {
    data += localState[i].value;
  }

  editor.innerHTML = data;

  if (prevCursorPosition !== 0) {
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
