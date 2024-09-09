package action

import (
	"fmt"
	"strings"

	"github.com/zyedidia/micro/v2/internal/display"
	"github.com/zyedidia/tcell/v2"
)

var ToolTips = Tooltip{false, nil}

type Tooltip struct {
	active bool

	// current active tooltip, !!!MAYBE BE NIL!!!
	curToolTip tooltipType
}

func (t *Tooltip) Display() {
	if !t.active {
		return
	}

	t.curToolTip.Display()
}

// API functions
// - Message(...) // Display mesasge such as definition of doc
// - Choice(...)  // Give user multiple choice for e.g. completion

// Message(...) // Display mesasge such as definition of doc
func (t *Tooltip) Message(msg ...interface{}) {
	t.curToolTip = NewMessageTooltip(msg...)
	t.active = true
}

// - Choice(...) // Give user multiple choice for e.g. completion
func (t *Tooltip) Choice(msg ...interface{}) {
	t.curToolTip = NewChoiceTooltip(msg...)
	t.active = true
}

// Intercepts events going to tab/pane to be handeled by active tooltips.
// May chose to consume event (return nil) or passthrough event to tab (return event).
// Default should be to passthorugh, especially if event is unsued/unhandled
func (t *Tooltip) InterceptEvent(event tcell.Event) tcell.Event {
	if !t.active {
		return event
	}

	rm, e := t.curToolTip.HandleEvent(event)
	if rm {
		t.active = false
		t.curToolTip = nil
	}

	return e
}

func GetTooltip() *Tooltip {
	return &ToolTips
}

type tooltipType interface {
	Display()
	HandleEvent(event tcell.Event) (bool, tcell.Event)
}

// Tooltip for multiple choices, for choicing autocomplete
type choiceTooltip struct {
	choices []string
	seleced int
}

func NewChoiceTooltip(msg ...interface{}) *choiceTooltip {
	if len(msg) == 1 {
		m := strings.Split(fmt.Sprint(msg...), "]]]]")
		return &choiceTooltip{m, 0}
	}

	t := new(choiceTooltip)
	for _, m := range msg {
		s := fmt.Sprint(m)
		for _, l := range strings.Split(s, "\n") {
			t.choices = append(t.choices, l)
		}
	}

	return t
}

func (t *choiceTooltip) Display() {
	lines := make([]string, len(t.choices))
	for i, s := range t.choices {
		c := ' '
		if i == t.seleced {
			c = 'X'
		}

		lines[i] = fmt.Sprintf("[%c] %s", c, s)
	}

	sw := display.NewSubWindowConformLines(nil, lines...)
	c := MainTab().CurPane()
	sw.DisplayAsTooltip(c.BWindow, c.Cursor)
}

// return shutdown, passthrough
func (t *choiceTooltip) HandleEvent(event tcell.Event) (bool, tcell.Event) {
	key, isKey := event.(*tcell.EventKey)

	if !isKey {
		return true, event
	}

	switch key.Key() {
	case tcell.KeyUp:
		t.seleced = (t.seleced - 1) % len(t.choices)
		return false, nil
	case tcell.KeyDown:
		t.seleced = (t.seleced + 1) % len(t.choices)
		return false, nil
	case tcell.KeyTab:
	case tcell.KeyEnter:
		pane := MainTab().CurPane()
		pane.Buf.Insert(pane.Cursor.Loc, t.choices[t.seleced])
	case tcell.KeyEsc:
		return true, nil
	default:
		return false, event

	}

	return false, event
}

// autocomplete for displaying simple messages like types for docs
type messageTooltip struct {
	lines []string
}

func NewMessageTooltip(msg ...interface{}) *messageTooltip {
	t := new(messageTooltip)
	t.lines = strings.Split(fmt.Sprint(msg...), "\n")
	return t
}

func (t *messageTooltip) Display() {

	sw := display.NewSubWindowConformLines(nil, t.lines...)
	c := MainTab().CurPane()
	sw.DisplayAsTooltip(c.BWindow, c.Cursor)
}

// return shutdown, passthrough
func (t *messageTooltip) HandleEvent(event tcell.Event) (bool, tcell.Event) {
	// always close tooltips
	// when user makes input
	_, mouse := event.(*tcell.EventMouse)
	if mouse {
		return false, event
	}

	return true, event
}
