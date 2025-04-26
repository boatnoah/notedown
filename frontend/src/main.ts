// src/main.ts
import { initAuth } from "./login";
import { initEditor } from "./ws";

document.addEventListener("DOMContentLoaded", () => {
  switch (location.pathname) {
    case "/editor":
      initEditor();
      break;

    default:
      initAuth();
  }
});
