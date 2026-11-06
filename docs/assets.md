# Visual assets

The repository keeps source images and the Windows-ready formats together so a
release can be reproduced without downloading design files.

## Source images

| File | Purpose |
|------|---------|
| `cmd/assets/windows-claude-swap-icon-v2.png` | Main application and installer icon source |
| `cmd/assets/windows-claude-swap-tray.png` | Transparent tray icon source |
| `installer/assets/first-run-guide.png` | First-installation guide source |

## Generated files

| File | Generated from | Used by |
|------|----------------|---------|
| `cmd/assets/windows-claude-swap-icon-v2.ico` | Main icon PNG | Installer, shortcuts, Apps list |
| `cmd/assets/windows-claude-swap-tray.ico` | Tray icon PNG | Embedded Windows tray icon |
| `rsrc_windows_amd64.syso`, `rsrc_windows_arm64.syso` | Main icon ICO, by `go run scripts/icon-resource.go` | The program's own icon: Explorer, shortcuts, Task Manager, Startup apps |
| `installer/assets/first-run-guide.bmp` | First-run guide PNG | Native first-run window installed with the app |

Generated assets are committed because the Go embed and Inno Setup build consume
those exact formats. When a source image changes, regenerate its derived file,
keep transparency where applicable, and run the validation command documented
in `CONTRIBUTING.md`. The icon tests require 16, 32, 48, and 256 pixel ICO
frames. The fishbowl is pixel art drawn at 16 pixels: every frame that is a
multiple of 16 is that drawing scaled by whole pixels, and the 20, 24 and 40
pixel frames are shrunk from a sharp larger copy.
