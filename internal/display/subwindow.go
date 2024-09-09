package display

import (
	"fmt"
	"strings"

	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
)

type Style struct {
	tl rune
	tr rune
	bl rune
	br rune
	h  rune
	v  rune
}

var (
	StyleRegular    = Style{'┌', '┐', '└', '┘', '─', '│'}
	StyleBold       = Style{'┏', '┓', '┗', '┛', '━', '┃'}
	StyleRounded    = Style{'╭', '╮', '╰', '╯', '─', '│'}
	StyleDouble     = Style{'╔', '╗', '╚', '╝', '═', '║'}
	StyleDotted     = Style{'┌', '┐', '└', '┘', '┄', '┆'}
	StyleDottedBold = Style{'┏', '┓', '┗', '┛', '┅', '┇'}

	StyleDefault = StyleRounded
)

func max(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

type SubWindow struct {
	Width  int
	Height int
	Lines  []string
	style  *Style
}

func NewSubWindowConformLines(style *Style, lines ...string) SubWindow {
	if style == nil {
		style = &StyleDefault
	}

	max_w := 0

	for _, l := range lines {
		max_w = max(max_w, len(l))
	}

	return SubWindow{max_w + 2, len(lines) + 2, lines, style}
}

func NewSubWindowStr(width int, height int, msg string, style *Style) SubWindow {
	if style == nil {
		style = &StyleDefault
	}

	lines := strings.Split(msg, "\n")
	return SubWindow{width, height, lines, style}
}

func NewSubWindowConform(msg string, style *Style) SubWindow {
	w := 0
	y := 3
	max_w := 0

	for _, c := range msg {
		if c == '\n' {
			y++
			w = 0
		}

		w++
		max_w = max(w, max_w)
	}

	return NewSubWindowStr(max_w+2, y, msg, style)
}

func (w *SubWindow) Display(x int, y int) {
	// draw corners
	screen.SetContent(x, y, w.style.tl, nil, config.DefStyle)
	screen.SetContent(x+w.Width-1, y, w.style.tr, nil, config.DefStyle)
	screen.SetContent(x, y+w.Height-1, w.style.bl, nil, config.DefStyle)
	screen.SetContent(x+w.Width-1, y+w.Height-1, w.style.br, nil, config.DefStyle)

	// draw top and bottom bars
	for i := x + 1; i < x+w.Width-1; i++ {
		screen.SetContent(i, y, w.style.h, nil, config.DefStyle)
		screen.SetContent(i, y+w.Height-1, w.style.h, nil, config.DefStyle)
	}

	// draw left and right bars
	for i := y + 1; i < y+w.Height-1; i++ {
		screen.SetContent(x, i, w.style.v, nil, config.DefStyle)
		screen.SetContent(x+w.Width-1, i, w.style.v, nil, config.DefStyle)
	}

	// write message
	for y_off, line := range w.Lines {
		// cant use range counter since its byte offset
		// and runes maybe more than 1 byte wide
		x_off := 0
		for _, c := range fmt.Sprintf("%-*s", w.Width-2, line) {
			screen.SetContent(x+x_off+1, y+y_off+1, c, nil, config.DefStyle)
			x_off++
		}
	}
}

func (sw *SubWindow) DisplayAsTooltip(w BWindow, c *buffer.Cursor) {
	y := c.Y - w.GetView().StartLine.Line
	x := c.GetVisualX()

	bufw, isbufw := w.(*BufWindow)

	if isbufw {
		x += bufw.gutterOffset
	}

	if y-sw.Height < 0 {
		// if not enought room above, draw below cursor
		y += 1
	} else {
		y -= sw.Height
	}

	sw.Display(x, y)
}
