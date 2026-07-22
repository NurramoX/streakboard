---
name: verify
description: Capture evidence of the streakboard TUI rendering — a headless pty dump of the kitty-graphics stream, or pixel screenshots of Ghostty driven via AppleScript. Use when verifying changes to tui/ or kitty/.
---

# Verifying streakboard in Ghostty

Two levels of evidence:

- **Stream (headless, no permissions)** — the kitty graphics
  transmission passes through the pty, so a raw dump contains the exact
  PNG the app sent, the placement command, and every placeholder cell.
  Proves everything streakboard *emits*; works without a display. Use
  this for most tui/ and kitty/ changes.
- **Pixels (Ghostty's renderer)** — a screenshot of the live window is
  the only proof the terminal actually *composites* the image over the
  placeholder cells. Needed when a change could affect how the image
  reaches the screen (placement semantics, placeholder colors or
  diacritics) — a stream can look right while a terminal shows nothing.

`Render`/PNG changes need neither — write a PNG and look at it.

What never works: any text-grid export. tmux `capture-pane` swallows
the APC even with `allow-passthrough on` (image unrecoverable AND
undisplayable), `write_screen_file:copy,plain` exports only the
placeholder cells, and the HTML export (`write_screen_file:copy,html`)
emits them as literal character entities — no `<img>`, no `data:` URI.
All three verified; not fixable.

## Stream evidence: pty dump

    go build -o ./streakboard-demo ./cmd/streakboard-demo
    python3 .claude/skills/verify/ptydump.py ./streakboard-demo <outdir>

Runs the demo on a 118x14 pty, answers its startup queries, sends q,
and reconstructs every kitty transmission. Then **look at**
`<outdir>/board-from-pty-id1.png` and check the printed placement
(rows=7, cols=2x*weeks*) and placeholder count (7xcols; 742 at 118
columns).

## Pixel evidence: script the running Ghostty (macOS)

Ghostty >= 1.3 has an AppleScript dictionary — drive the instance that
is already running. Never launch a second instance: that resurrects
session-restored windows and turns window lookup into archaeology.

1. Build **inside the repo** (see above; binaries under /private/tmp
   make Ghostty pop a Gatekeeper allow dialog). The path is gitignored;
   delete the binary afterwards.
2. Launch, focus, and size the demo window (focus matters — see
   failure modes):

       osascript <<'EOF'
       tell application "Ghostty"
           set w to new window with configuration {command:"/abs/path/to/streakboard-demo"}
           delay 2
           set t to focused terminal of selected tab of w
           activate window w
           delay 0.5
           perform action "toggle_maximize" on t
           delay 1
           return id of w
       end tell
       EOF

   Keep the returned id and reference the window later as
   `window id "tab-group-..."` — but assign it to a variable first;
   inline `close window id "..."` fails to parse.
3. Drive the TUI only with `perform action "text:<chars>" on t`
   (e.g. `"text:t"` cycles the theme). `send key` silently drops
   character keys everywhere (see failure modes), and `input text`
   arrives as a bracketed paste — a tea.PasteMsg that key handlers
   never see.
4. Map to a CGWindowID with the Swift snippet below — the AppleScript
   window id is Ghostty-internal, not a CGWindowID. The demo window's
   title is 👻 (direct-command surfaces get no shell-integration
   title); pick the full-size one.
5. `screencapture -x -o -l <id> board.png`, then **look at the image**.
6. Tear down: `perform action "text:q" on t` quits the demo, but
   scripted surfaces wait after command exit ("Process exited. Press
   any key…") — follow with `close window w`. Delete the binary.

```swift
// winlist.swift — swift winlist.swift
import CoreGraphics
import Foundation
let list = CGWindowListCopyWindowInfo([.excludeDesktopElements], kCGNullWindowID) as? [[String: Any]] ?? []
for w in list where (w[kCGWindowOwnerName as String] as? String ?? "").lowercased().contains("ghostty") {
    let b = w[kCGWindowBounds as String] as? [String: Any] ?? [:]
    print(w[kCGWindowNumber as String] ?? 0, "|", w[kCGWindowName as String] as? String ?? "<no title>", "|",
          b["Width"] ?? 0, "x", b["Height"] ?? 0)
}
```

## Failure modes

Every one observed in a real session; fixes are what actually worked.

- **`screencapture -l` says "could not create image from window"** has
  two distinct causes. (a) Ghostty (TCC-responsible for everything run
  inside it) lacks Screen Recording permission — then window titles
  also list as `<no title>`; one-time fix in System Settings → Privacy
  & Security → Screen Recording. (b) Permission is fine but the window
  sits on a non-active Space or display — titles are visible and
  `screencapture -R` works while `-l` fails; fix: `activate window`
  first. Check titles to tell the two apart before debugging.
- **`send key` silently drops character keys** ("a", "t"...) on every
  surface type; named keys ("enter") arrive, oddly as LF instead of
  CR, and keybind combos don't fire. Ghostty 1.3.1 bug: the scripted
  KeyEvent carries no text/unshifted codepoint, and the core encodes
  printable keys from exactly those fields (ScriptKeyEventCommand
  .swift). Use `perform action "text:..."`.
- **JXA window lookup segfaults** (`osascript -l JavaScript` + ObjC
  bridge exits 139 on CFBridgingRelease). Use the Swift snippet.
- **`write_screen_file:copy,...` clobbers the clipboard** — save with
  `pbpaste` first, restore with `pbcopy` after.
- **The demo's data is random per run**; only layout and label
  alignment are stable, not the cell pattern — and `t`/`n` regenerate
  the data as well.
