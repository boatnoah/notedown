# Notedown

Feats:

- [x] Detect change to the editor.
- [x] Send raw array buffer over websocket.
  - [ ] Avoid redudant packets from being applied.
- [x] Go distributes it to all clients
- [x] On message convert array buffer to Uint8array
- [x] Apply changes
- [x] Implement awareness protocol
- [ ] when a new client joins figure out a way to render the state for the joined client
- [ ] create an enpoint that gives you all the connections from the connections map

Chores:

- [ ] fix write error
