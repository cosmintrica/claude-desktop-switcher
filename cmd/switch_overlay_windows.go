//go:build windows

package cmd

import (
	"errors"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Progress cards look like a short terminal session floating over the screen:
// the command types itself, a bar fills with the operation's steps and the
// current step sits centered below it, all three lines one width. The window
// has no background: every pixel carries its own alpha, and a dark outline
// keeps the text readable over light and dark windows alike.

const (
	cardTickMilliseconds = 50
	cardMinColumns       = 30
	cardPadding          = 18
	cardHaloRadius       = 2
	cardHaloOpacity      = 0.72
	cardFontHeight       = 24
	cardCreepAhead       = 0.12
	cardClassName        = "WindowsClaudeSwapCard"
)

var (
	cardWhite  = cardColor{r: 255, g: 255, b: 255, opacity: 1}
	cardOrange = cardColor{r: 217, g: 119, b: 87, opacity: 1}
	cardTrack  = cardColor{r: 217, g: 119, b: 87, opacity: 0.45}
	cardPeach  = cardColor{r: 242, g: 211, b: 196, opacity: 1}
)

var (
	cardCreateDIBSection = nativeGDI32.NewProc("CreateDIBSection")
	cardTextOut          = nativeGDI32.NewProc("TextOutW")
	cardTextExtent       = nativeGDI32.NewProc("GetTextExtentPoint32W")
	cardTextFace         = nativeGDI32.NewProc("GetTextFaceW")
	cardGdiFlush         = nativeGDI32.NewProc("GdiFlush")
	cardUpdateLayered    = nativeUser32.NewProc("UpdateLayeredWindow")
	cardGetDC            = nativeUser32.NewProc("GetDC")
	cardReleaseDC        = nativeUser32.NewProc("ReleaseDC")
	cardClassOnce        sync.Once
	cardClassErr         error
	cardWindowsMu        sync.Mutex
	cardWindows          = make(map[uintptr]*cardWindow)
	cardWndProc          = windows.NewCallback(cardWindowProc)
)

type cardColor struct {
	r, g, b uint8
	opacity float32
}

type cardText struct {
	column, row int
	text        string
}

type cardLayer struct {
	color cardColor
	texts []cardText
}

type switchOverlay struct {
	hwnd  uintptr
	ready chan error
	done  chan struct{}
	card  *terminalCard
}

func startSwitchOverlay(target string) *switchOverlay {
	return startTerminalCard("claude switch "+target, "closing Claude", 0.04)
}

func startAddPreparationOverlay() *switchOverlay {
	return startTerminalCard("claude account add", "preparing a clean sign-in", 0.1)
}

func startAddCompletionOverlay() *switchOverlay {
	return startTerminalCard("claude account add", "saving the new account", 0.5)
}

func startLoginReopenOverlay() *switchOverlay {
	return startTerminalCard("claude open", "reopening Claude", 0.1)
}

func startAddSuccessOverlay() *switchOverlay {
	return startTerminalCard("claude account add", "account added", 1)
}

func startBackupPreparationOverlay() *switchOverlay {
	return startTerminalCard("claude backup", "preparing the backup", 0.1)
}

func startInitialAccountSaveOverlay() *switchOverlay {
	return startTerminalCard("claude account save", "saving the current account", 0.1)
}

func (o *switchOverlay) Step(step string, progress float64) {
	if o == nil || o.card == nil {
		return
	}
	o.card.setStep(step, progress)
}

func (o *switchOverlay) Close() {
	if o == nil || o.hwnd == 0 {
		return
	}
	nativePostMessage.Call(o.hwnd, nativeWMClose, 0, 0)
	<-o.done
}

type terminalCard struct {
	mu      sync.Mutex
	command []rune
	step    string
	stepAt  float64
	target  float64
	shown   float64
	typed   int
	tick    int
	columns int
	eighths bool
}

func newTerminalCard(command, step string, progress float64) *terminalCard {
	prompt := []rune("$ " + command)
	card := &terminalCard{command: prompt, columns: max(len(prompt)+1, cardMinColumns)}
	card.setStep(step, progress)
	if progress >= 1 {
		card.shown, card.typed = 1, len(prompt)
	}
	return card
}

func (c *terminalCard) setStep(step string, progress float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.step = step
	c.stepAt = max(c.stepAt, min(progress, 1))
	c.target = max(c.target, c.stepAt)
}

// Each tick types one more character and eases the bar toward its target. The
// target creeps a little past the last step, so a long step still moves
// without reaching the next one.
func (c *terminalCard) advance() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tick++
	if c.typed < len(c.command) {
		c.typed++
	}
	if ceiling := min(c.stepAt+cardCreepAhead, 0.97); c.stepAt < 1 && c.target < ceiling {
		c.target += (ceiling - c.target) * 0.01
	}
	c.shown += (c.target - c.shown) * 0.2
	if c.target-c.shown < 0.001 {
		c.shown = c.target
	}
}

func (c *terminalCard) layers() []cardLayer {
	c.mu.Lock()
	defer c.mu.Unlock()
	track := cardLayer{color: cardTrack}
	white := cardLayer{color: cardWhite}
	orange := cardLayer{color: cardOrange}
	peach := cardLayer{color: cardPeach}

	start := (c.columns - len(c.command) - 1) / 2
	typed := c.command[:c.typed]
	if len(typed) > 0 {
		orange.texts = append(orange.texts, cardText{column: start, text: string(typed[:1])})
		white.texts = append(white.texts, cardText{column: start + 1, text: string(typed[1:])})
	}
	if c.tick/8%2 == 0 {
		orange.texts = append(orange.texts, cardText{column: start + len(typed), text: "▌"})
	}

	inner := c.columns - 2
	cells := c.shown * float64(inner)
	full := min(int(cells), inner)
	white.texts = append(white.texts, cardText{column: 0, row: 1, text: "["}, cardText{column: c.columns - 1, row: 1, text: "]"})
	orange.texts = append(orange.texts, cardText{column: 1, row: 1, text: strings.Repeat("█", full)})
	if full < inner {
		track.texts = append(track.texts, cardText{column: 1 + full, row: 1, text: strings.Repeat("░", inner-full)})
		orange.texts = append(orange.texts, cardText{column: 1 + full, row: 1, text: partialCell(cells-float64(full), c.eighths)})
	}

	step := []rune(c.step)
	if len(step) > c.columns {
		step = append(step[:c.columns-1], '…')
	}
	peach.texts = append(peach.texts, cardText{column: (c.columns - len(step)) / 2, row: 2, text: string(step)})
	return []cardLayer{track, white, orange, peach}
}

// Cascadia Mono has eighth blocks for a smooth edge; Consolas only has shades.
func partialCell(fraction float64, eighths bool) string {
	if eighths {
		return []string{"", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}[min(int(fraction*8), 7)]
	}
	switch {
	case fraction >= 2.0/3:
		return "▓"
	case fraction >= 1.0/3:
		return "▒"
	}
	return ""
}

type cardWindow struct {
	card      *terminalCard
	dc        uintptr
	bitmap    uintptr
	oldBitmap uintptr
	font      uintptr
	oldFont   uintptr
	pixels    []byte
	premult   [4][]float32
	width     int
	height    int
	x, y      int
	cellWidth int
	lineStep  int
}

// terminalCardUnless is the terminal card, or a card that shows nothing when the
// aquarium draws the progress itself.
func terminalCardUnless(quiet bool, command, step string, progress float64) *switchOverlay {
	if quiet {
		return &switchOverlay{}
	}
	return startTerminalCard(command, step, progress)
}

func startTerminalCard(command, step string, progress float64) *switchOverlay {
	overlay := &switchOverlay{ready: make(chan error, 1), done: make(chan struct{}), card: newTerminalCard(command, step, progress)}
	go runTerminalCard(overlay)
	if err := <-overlay.ready; err != nil {
		return &switchOverlay{}
	}
	return overlay
}

func runTerminalCard(overlay *switchOverlay) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(overlay.done)
	cardClassOnce.Do(func() { cardClassErr = nativeRegisterClass(cardClassName, cardWndProc) })
	if cardClassErr != nil {
		overlay.ready <- cardClassErr
		return
	}
	window, err := newCardWindow(overlay.card)
	if err != nil {
		overlay.ready <- err
		return
	}
	defer window.release()
	hInstance, _, _ := nativeGetModule.Call(0)
	hwnd, _, _ := nativeCreateWindow.Call(
		nativeWSExLayered|nativeWSExTopmost|nativeWSExToolWindow|nativeWSExNoActivate|nativeWSExTransparent,
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(cardClassName))),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(ProductName))),
		nativeWSPopup,
		uintptr(window.x), uintptr(window.y), uintptr(window.width), uintptr(window.height),
		0, 0, hInstance, 0,
	)
	if hwnd == 0 {
		overlay.ready <- errors.New("could not create the progress card")
		return
	}
	cardWindowsMu.Lock()
	cardWindows[hwnd] = window
	cardWindowsMu.Unlock()
	window.draw(hwnd)
	nativeShowWindow.Call(hwnd, nativeSWShowNoActivate)
	nativeSetTimer.Call(hwnd, nativeTimerID, cardTickMilliseconds, 0)
	overlay.hwnd = hwnd
	overlay.ready <- nil

	var message nativeMSG
	for {
		result, _, _ := nativeGetMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if result == 0 || int32(result) == -1 {
			break
		}
		nativeTranslate.Call(uintptr(unsafe.Pointer(&message)))
		nativeDispatch.Call(uintptr(unsafe.Pointer(&message)))
	}
	cardWindowsMu.Lock()
	delete(cardWindows, hwnd)
	cardWindowsMu.Unlock()
}

func cardWindowProc(hwnd, message, wParam, lParam uintptr) uintptr {
	cardWindowsMu.Lock()
	window := cardWindows[hwnd]
	cardWindowsMu.Unlock()
	if window != nil {
		switch message {
		case nativeWMTimer:
			window.card.advance()
			window.draw(hwnd)
			return 0
		case nativeWMClose:
			nativeKillTimer.Call(hwnd, nativeTimerID)
			nativeDestroyWindow.Call(hwnd)
			return 0
		case nativeWMDestroy:
			nativePostQuit.Call(0)
			return 0
		}
	}
	result, _, _ := nativeDefWindowProc.Call(hwnd, message, wParam, lParam)
	return result
}

type cardBitmapInfo struct {
	size          uint32
	width         int32
	height        int32
	planes        uint16
	bitCount      uint16
	compression   uint32
	sizeImage     uint32
	xPelsPerMeter int32
	yPelsPerMeter int32
	clrUsed       uint32
	clrImportant  uint32
}

type cardPoint struct{ x, y int32 }

type cardSize struct{ cx, cy int32 }

type cardBlend struct{ op, flags, constantAlpha, alphaFormat byte }

func newCardWindow(card *terminalCard) (*cardWindow, error) {
	screen, _, _ := cardGetDC.Call(0)
	if screen == 0 {
		return nil, errors.New("could not reach the screen")
	}
	defer cardReleaseDC.Call(0, screen)
	dc, _, _ := nativeCreateCompatibleDC.Call(screen)
	if dc == 0 {
		return nil, errors.New("could not prepare the progress card")
	}
	window := &cardWindow{card: card, dc: dc}
	eighths := window.selectFont()
	card.mu.Lock()
	card.eighths = eighths
	card.mu.Unlock()
	var cell cardSize
	block := windows.StringToUTF16("█")
	cardTextExtent.Call(dc, uintptr(unsafe.Pointer(&block[0])), 1, uintptr(unsafe.Pointer(&cell)))
	window.cellWidth = int(cell.cx)
	window.lineStep = int(cell.cy) * 13 / 10
	window.width = 2*cardPadding + card.columns*window.cellWidth
	window.height = 2*cardPadding + 2*window.lineStep + int(cell.cy)
	screenWidth, _, _ := nativeGetMetrics.Call(nativeSMCXScreen)
	screenHeight, _, _ := nativeGetMetrics.Call(nativeSMCYScreen)
	window.x = (int(screenWidth) - window.width) / 2
	window.y = (int(screenHeight) - window.height) / 2

	info := cardBitmapInfo{width: int32(window.width), height: -int32(window.height), planes: 1, bitCount: 32}
	info.size = uint32(unsafe.Sizeof(info))
	var bits unsafe.Pointer
	window.bitmap, _, _ = cardCreateDIBSection.Call(dc, uintptr(unsafe.Pointer(&info)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if window.bitmap == 0 || bits == nil {
		window.release()
		return nil, errors.New("could not prepare the progress card")
	}
	window.oldBitmap, _, _ = nativeSelectObject.Call(dc, window.bitmap)
	window.pixels = unsafe.Slice((*byte)(bits), window.width*window.height*4)
	for channel := range window.premult {
		window.premult[channel] = make([]float32, window.width*window.height)
	}
	nativeSetBkMode.Call(dc, nativeTransparent)
	nativeSetTextColor.Call(dc, 0x00FFFFFF)
	return window, nil
}

// Cascadia Mono, the Windows Terminal font, when Windows has it; Consolas is
// present on every Windows since Vista.
func (w *cardWindow) selectFont() bool {
	height := int32(-cardFontHeight)
	for _, face := range []string{"Cascadia Mono", "Consolas"} {
		font, _, _ := nativeCreateFont.Call(
			uintptr(height), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 4, 0x31,
			uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(face))),
		)
		if font == 0 {
			continue
		}
		previous, _, _ := nativeSelectObject.Call(w.dc, font)
		name := make([]uint16, 64)
		cardTextFace.Call(w.dc, uintptr(len(name)), uintptr(unsafe.Pointer(&name[0])))
		if strings.EqualFold(windows.UTF16ToString(name), face) || face == "Consolas" {
			if w.oldFont == 0 {
				w.oldFont = previous
			}
			w.font = font
			return strings.EqualFold(windows.UTF16ToString(name), "Cascadia Mono")
		}
		nativeSelectObject.Call(w.dc, previous)
		nativeDeleteObject.Call(font)
	}
	return false
}

func (w *cardWindow) release() {
	if w.oldBitmap != 0 {
		nativeSelectObject.Call(w.dc, w.oldBitmap)
	}
	if w.bitmap != 0 {
		nativeDeleteObject.Call(w.bitmap)
	}
	if w.oldFont != 0 {
		nativeSelectObject.Call(w.dc, w.oldFont)
	}
	if w.font != 0 {
		nativeDeleteObject.Call(w.font)
	}
	nativeDeleteDC.Call(w.dc)
}

// GDI draws antialiased white glyphs on black but never writes alpha, so each
// layer is drawn alone and its brightness becomes that layer's coverage.
func (w *cardWindow) draw(hwnd uintptr) {
	layers := w.card.layers()
	clear(w.pixels)
	for _, layer := range layers {
		for _, text := range layer.texts {
			for dy := -cardHaloRadius; dy <= cardHaloRadius; dy++ {
				for dx := -cardHaloRadius; dx <= cardHaloRadius; dx++ {
					if dx*dx+dy*dy <= cardHaloRadius*cardHaloRadius && (dx != 0 || dy != 0) {
						w.text(text, dx, dy)
					}
				}
			}
		}
	}
	cardGdiFlush.Call()
	red, green, blue, alpha := w.premult[0], w.premult[1], w.premult[2], w.premult[3]
	for index := range alpha {
		red[index], green[index], blue[index] = 0, 0, 0
		alpha[index] = float32(w.pixels[index*4]) / 255 * cardHaloOpacity
	}
	for _, layer := range layers {
		clear(w.pixels)
		for _, text := range layer.texts {
			w.text(text, 0, 0)
		}
		cardGdiFlush.Call()
		r, g, b := float32(layer.color.r)/255, float32(layer.color.g)/255, float32(layer.color.b)/255
		for index := range alpha {
			coverage := float32(w.pixels[index*4]) / 255 * layer.color.opacity
			if coverage == 0 {
				continue
			}
			keep := 1 - coverage
			red[index] = r*coverage + red[index]*keep
			green[index] = g*coverage + green[index]*keep
			blue[index] = b*coverage + blue[index]*keep
			alpha[index] = coverage + alpha[index]*keep
		}
	}
	for index := range alpha {
		w.pixels[index*4] = uint8(blue[index]*255 + 0.5)
		w.pixels[index*4+1] = uint8(green[index]*255 + 0.5)
		w.pixels[index*4+2] = uint8(red[index]*255 + 0.5)
		w.pixels[index*4+3] = uint8(alpha[index]*255 + 0.5)
	}
	screen, _, _ := cardGetDC.Call(0)
	defer cardReleaseDC.Call(0, screen)
	position := cardPoint{x: int32(w.x), y: int32(w.y)}
	size := cardSize{cx: int32(w.width), cy: int32(w.height)}
	origin := cardPoint{}
	blend := cardBlend{constantAlpha: 255, alphaFormat: 1}
	cardUpdateLayered.Call(hwnd, screen, uintptr(unsafe.Pointer(&position)), uintptr(unsafe.Pointer(&size)), w.dc, uintptr(unsafe.Pointer(&origin)), 0, uintptr(unsafe.Pointer(&blend)), 2)
}

func (w *cardWindow) text(text cardText, dx, dy int) {
	if text.text == "" {
		return
	}
	utf := windows.StringToUTF16(text.text)
	x := cardPadding + text.column*w.cellWidth + dx
	y := cardPadding + text.row*w.lineStep + dy
	cardTextOut.Call(w.dc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(&utf[0])), uintptr(len(utf)-1))
}
