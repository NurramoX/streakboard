---
name: verify
description: Capture pixel evidence of the streakboard TUI rendering in Ghostty — the only way to observe the kitty-graphics image. Use when verifying changes to tui/ or kitty/.
---

# Verifying streakboard in Ghostty

Text-level captures (tmux, pty dumps, Ghostty's HTML export) cannot
see the board: the image is composited by the terminal's renderer and
never exists in the text grid. Only a pixel screenshot proves the
board renders. `Render`/PNG changes don't need any of this — write a
PNG and look at it.

## Recipe (macOS)

1. Build the demo **inside the repo** (Gatekeeper blocks Ghostty from
   executing binaries in /tmp — it pops an Allow/Cancel dialog):
   `go build -o ./streakboard-demo ./cmd/streakboard-demo`
   (the path is gitignored; delete the binary afterwards).
2. Launch a second Ghostty instance:
   `/Applications/Ghostty.app/Contents/MacOS/ghostty -e $PWD/streakboard-demo &`
   Requires Ghostty to have macOS Screen Recording permission
   (System Settings → Privacy & Security), or step 4 fails with
   "could not create image from window".
3. Find the window id with the Swift snippet below. The demo window's
   title is the binary path; off-screen windows capture fine. Session
   restore may open extra shell windows — match by title, not count.
4. `screencapture -x -o -l <id> board.png`, then **look at the image**.
5. Kill the instance you launched (its pid is the ghostty process
   whose argv contains `-e`); that closes its windows.

```swift
// winlist.swift — swift winlist.swift
import CoreGraphics
import Foundation
let list = CGWindowListCopyWindowInfo([.excludeDesktopElements], kCGNullWindowID) as? [[String: Any]] ?? []
for w in list where (w[kCGWindowOwnerName as String] as? String ?? "").lowercased().contains("ghostty") {
    print(w[kCGWindowNumber as String] ?? 0, "|", w[kCGWindowName as String] as? String ?? "", "|",
          (w[kCGWindowBounds as String] as? [String: Any])?["Width"] ?? 0)
}
```

## Gotchas

- Don't use the JXA/osascript ObjC bridge for the window list — it
  segfaults on CFBridgingRelease.
- `--window-width/--window-height` (cells) are unreliable: with
  session restore active the demo often attaches to a restored
  full-size window (good — use it); with `--window-save-state=never`
  the window opened at the requested size but was later shrunk.
- The demo's data is randomly generated per run; only layout and
  label alignment are stable, not the cell pattern.
