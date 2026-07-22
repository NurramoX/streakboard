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

## Failure modes, in the order we hit them

Every one of these happened in a real session; the fix column is
what actually worked, not a guess.

- **Text capture shows no board.** tmux panes, pty dumps, and
  Ghostty's HTML export are structurally blind to the image (see
  intro). Not fixable — go to pixels.
- **JXA window lookup segfaults** (`osascript -l JavaScript` +
  ObjC bridge exits 139 on CFBridgingRelease; the deepUnwrap
  variant silently returns nothing). Fix: the Swift snippet above.
- **`screencapture -l` says "could not create image from window".**
  Ghostty (TCC-responsible for everything run inside it) lacks
  Screen Recording permission. Fix: System Settings → Privacy &
  Security → Screen Recording → Ghostty. Bonus symptom of the same
  cause: every window lists as `<no title>` — window names are
  TCC-gated too, so don't debug title matching before permission.
- **`-e` runs but no demo process appears** (no `login -flp` child,
  instance idles with a tiny ~76×103 window). That window is a
  hidden Gatekeeper dialog: "Allow Ghostty to execute <binary>?" —
  triggered by binaries under /private/tmp. Fix: build the binary
  inside the repo and pass that path. Diagnose mystery tiny windows
  by screenshotting them.
- **Session restore muddies the window list**: a new instance also
  reopens previous windows (old titles, shells in old cwds), and
  the `-e` surface may attach to a tiny window while a full-size
  restored one sits on another Space. Match windows by owner pid +
  title (= the binary path) + bounds, never by count or recency.
  `onscreen=false` windows capture fine.
- **`--window-width/--window-height` (cells) are unreliable**: often
  ignored; once honored (1184×326 for 118×14) and then shrunk to
  255×120 anyway. Resizing via System Events needs Accessibility
  permission (error -1719 without it). Working fix: launch with
  session restore active and capture the restored full-size window
  the demo attaches to.
- **Tiny windows produce useless evidence**: captures are
  window-sized (a 258×179 window yields a 258×179 PNG with shadow
  and perspective). If the capture is small, fix the window, don't
  squint.
- The demo's data is randomly generated per run; only layout and
  label alignment are stable, not the cell pattern.
