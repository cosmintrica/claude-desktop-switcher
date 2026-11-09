#!/bin/bash
# Builds "Claude Desktop Switcher.app" for macOS as a universal (Apple silicon
# and Intel) menu bar app under installer/dist/macos.
#
#   scripts/build-macos-app.sh [version] [--install]
#
# The version defaults to v<upstream version>-fork-<commit>, like the Windows
# fork builds. --install also copies the app to ~/Applications.
set -euo pipefail

repository="$(cd "$(dirname "$0")/.." && pwd)"
version=""
install=false
for arg in "$@"; do
	case "$arg" in
	--install) install=true ;;
	*) version="$arg" ;;
	esac
done
if [ -z "$version" ]; then
	upstream="$(sed -n 's/^## \[\([0-9][0-9.]*\)\].*/\1/p' "$repository/CHANGELOG.md" | head -1)"
	version="$upstream-fork-$(git -C "$repository" rev-parse --short HEAD)"
fi

# The bundle's version keys take numbers only; the binary keeps the full one.
bundle_version="${version#v}"
bundle_version="${bundle_version%%-*}"
name="Claude Desktop Switcher"
executable="claude-desktop-switcher"
dist="$repository/installer/dist/macos"
app="$dist/$name.app"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

cd "$repository"
for arch in arm64 amd64; do
	CGO_ENABLED=1 GOOS=darwin GOARCH="$arch" go build -trimpath \
		-ldflags "-s -w -X github.com/TinoA/claude-desktop-switcher/cmd.Version=$version" \
		-o "$work/$arch" .
done

rm -rf "$app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
lipo -create -output "$app/Contents/MacOS/$executable" "$work/arm64" "$work/amd64"
cp "$repository/installer/macos/AppIcon.icns" "$app/Contents/Resources/AppIcon.icns"
cat >"$app/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>$name</string>
	<key>CFBundleDisplayName</key>
	<string>$name</string>
	<key>CFBundleIdentifier</key>
	<string>io.github.cosmintrica.claude-desktop-switcher</string>
	<key>CFBundleExecutable</key>
	<string>$executable</string>
	<key>CFBundleIconFile</key>
	<string>AppIcon</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>$bundle_version</string>
	<key>CFBundleVersion</key>
	<string>$bundle_version</string>
	<key>LSMinimumSystemVersion</key>
	<string>12.0</string>
	<key>LSUIElement</key>
	<true/>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
PLIST

# An ad-hoc signature is enough to run on this Mac; Gatekeeper still asks once
# for a copy downloaded from elsewhere.
codesign --force --deep --sign - "$app"
echo "Built $app (version $version)"

if $install; then
	mkdir -p "$HOME/Applications"
	rm -rf "$HOME/Applications/$name.app"
	cp -R "$app" "$HOME/Applications/"
	echo "Installed $HOME/Applications/$name.app"
fi
