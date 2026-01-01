package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aiomayo/aiodl/internal/adapter"
	"github.com/aiomayo/aiodl/internal/tui"
)

type PlaylistItem struct {
	info adapter.MediaInfo
}

func (i PlaylistItem) Title() string       { return i.info.Title }
func (i PlaylistItem) Description() string { return tui.FormatDuration(i.info.Duration) }
func (i PlaylistItem) FilterValue() string { return i.info.Title }

type playlistDelegate struct {
	selected map[string]bool
}

func (d playlistDelegate) Height() int                             { return 1 }
func (d playlistDelegate) Spacing() int                            { return 0 }
func (d playlistDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d playlistDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(PlaylistItem)
	if !ok {
		return
	}

	isCursor := index == m.Index()
	isSelected := d.selected[i.info.ID]

	cursor := "  "
	if isCursor {
		cursor = "> "
	}

	checkbox := "[ ]"
	if isSelected {
		checkbox = "[x]"
	}

	title := i.info.Title
	maxTitleLen := m.Width() - 25
	if maxTitleLen < 20 {
		maxTitleLen = 20
	}
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen-3] + "..."
	}

	duration := ""
	if i.info.Duration > 0 {
		duration = fmt.Sprintf(" (%s)", tui.FormatDuration(i.info.Duration))
	}

	idx := fmt.Sprintf("%3d. ", index+1)

	line := fmt.Sprintf("%s%s%s %s%s", cursor, idx, checkbox, title, duration)
	if isCursor {
		line = lipgloss.NewStyle().Bold(true).Render(line)
	}

	_, _ = fmt.Fprint(w, line)
}

type playlistKeyMap struct {
	Toggle  key.Binding
	All     key.Binding
	None    key.Binding
	Invert  key.Binding
	Confirm key.Binding
	Quit    key.Binding
}

func (k playlistKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Toggle, k.All, k.None, k.Confirm, k.Quit}
}

func (k playlistKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Toggle, k.All, k.None, k.Invert}, {k.Confirm, k.Quit}}
}

var playlistKeys = playlistKeyMap{
	Toggle:  key.NewBinding(key.WithKeys(" ", "x"), key.WithHelp("space/x", "toggle")),
	All:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "all")),
	None:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "none")),
	Invert:  key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "invert")),
	Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	Quit:    key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
}

type PlaylistBrowserModel struct {
	list      list.Model
	items     []adapter.MediaInfo
	selected  map[string]bool
	delegate  *playlistDelegate
	keys      playlistKeyMap
	confirmed bool
	cancelled bool
}

func NewPlaylistBrowser(title string, items []adapter.MediaInfo) PlaylistBrowserModel {
	selected := make(map[string]bool)

	delegate := &playlistDelegate{selected: selected}

	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = PlaylistItem{info: item}
	}

	l := list.New(listItems, delegate, 80, 20)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{playlistKeys.Toggle, playlistKeys.All}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{playlistKeys.Toggle, playlistKeys.All, playlistKeys.None, playlistKeys.Invert}
	}

	return PlaylistBrowserModel{
		list:     l,
		items:    items,
		selected: selected,
		delegate: delegate,
		keys:     playlistKeys,
	}
}

func (m PlaylistBrowserModel) Init() tea.Cmd {
	return nil
}

func (m PlaylistBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.cancelled = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Toggle):
			if item, ok := m.list.SelectedItem().(PlaylistItem); ok {
				m.selected[item.info.ID] = !m.selected[item.info.ID]
			}
			return m, nil

		case key.Matches(msg, m.keys.All):
			for _, item := range m.items {
				m.selected[item.ID] = true
			}
			return m, nil

		case key.Matches(msg, m.keys.None):
			m.selected = make(map[string]bool)
			m.delegate.selected = m.selected
			return m, nil

		case key.Matches(msg, m.keys.Invert):
			for _, item := range m.items {
				m.selected[item.ID] = !m.selected[item.ID]
			}
			return m, nil

		case key.Matches(msg, m.keys.Confirm):
			m.confirmed = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m PlaylistBrowserModel) View() string {
	var b strings.Builder

	count := m.SelectedCount()
	b.WriteString(fmt.Sprintf("%d of %d selected\n\n", count, len(m.items)))
	b.WriteString(m.list.View())

	return b.String()
}

func (m PlaylistBrowserModel) SelectedCount() int {
	count := 0
	for _, v := range m.selected {
		if v {
			count++
		}
	}
	return count
}

func (m PlaylistBrowserModel) SelectedItems() []adapter.MediaInfo {
	var result []adapter.MediaInfo
	for _, item := range m.items {
		if m.selected[item.ID] {
			result = append(result, item)
		}
	}
	return result
}

func (m PlaylistBrowserModel) Cancelled() bool {
	return m.cancelled
}
