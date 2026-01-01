package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aiomayo/aiodl/internal/adapter"
	"github.com/aiomayo/aiodl/internal/tui"
)

type formatKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Sort   key.Binding
	Video  key.Binding
	Audio  key.Binding
	All    key.Binding
	Quit   key.Binding
}

func (k formatKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Sort, k.Video, k.Audio, k.All, k.Quit}
}

func (k formatKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Select}, {k.Sort, k.Video, k.Audio, k.All}, {k.Quit}}
}

var defaultFormatKeys = formatKeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "down")),
	Select: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Sort:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort")),
	Video:  key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "video")),
	Audio:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "audio")),
	All:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "all")),
	Quit:   key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
}

type SortMode int

const (
	SortByQuality SortMode = iota
	SortBySize
	SortByBitrate
)

type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterVideoOnly
	FilterAudioOnly
)

type FormatSelectorModel struct {
	title      string
	formats    []adapter.Format
	filtered   []adapter.Format
	table      table.Model
	help       help.Model
	keys       formatKeyMap
	sortMode   SortMode
	filterMode FilterMode
	selected   *adapter.Format
	cancelled  bool
	width      int
	height     int
}

func NewFormatSelector(title string, formats []adapter.Format) FormatSelectorModel {
	columns := []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Type", Width: 8},
		{Title: "Quality", Width: 12},
		{Title: "Ext", Width: 8},
		{Title: "Size", Width: 12},
		{Title: "Bitrate", Width: 12},
	}

	m := FormatSelectorModel{
		title:    title,
		formats:  formats,
		filtered: formats,
		help:     help.New(),
		keys:     defaultFormatKeys,
	}

	rows := m.buildRows(formats)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(15),
	)
	t.SetStyles(table.DefaultStyles())
	m.table = t
	m.applyFiltersAndSort()

	return m
}

func (m FormatSelectorModel) buildRows(formats []adapter.Format) []table.Row {
	var rows []table.Row
	for _, f := range formats {
		ftype := "audio"
		if f.IsVideo() {
			ftype = "video"
		}
		rows = append(rows, table.Row{
			f.ID, ftype, f.Quality, f.Extension,
			tui.FormatSize(f.FileSize), tui.FormatBitrate(f.Bitrate),
		})
	}
	return rows
}

func (m FormatSelectorModel) Init() tea.Cmd {
	return nil
}

func (m FormatSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.table.SetHeight(msg.Height - 10)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.cancelled = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Select):
			if len(m.filtered) > 0 {
				idx := m.table.Cursor()
				if idx < len(m.filtered) {
					m.selected = &m.filtered[idx]
				}
			}
			return m, tea.Quit

		case key.Matches(msg, m.keys.Sort):
			m.sortMode = (m.sortMode + 1) % 3
			m.applyFiltersAndSort()

		case key.Matches(msg, m.keys.Video):
			m.filterMode = FilterVideoOnly
			m.applyFiltersAndSort()

		case key.Matches(msg, m.keys.Audio):
			m.filterMode = FilterAudioOnly
			m.applyFiltersAndSort()

		case key.Matches(msg, m.keys.All):
			m.filterMode = FilterAll
			m.applyFiltersAndSort()
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *FormatSelectorModel) applyFiltersAndSort() {
	var filtered []adapter.Format
	for _, f := range m.formats {
		switch m.filterMode {
		case FilterVideoOnly:
			if f.IsVideo() {
				filtered = append(filtered, f)
			}
		case FilterAudioOnly:
			if !f.IsVideo() {
				filtered = append(filtered, f)
			}
		default:
			filtered = append(filtered, f)
		}
	}

	switch m.sortMode {
	case SortByQuality:
		sort.Slice(filtered, func(i, j int) bool {
			if filtered[i].Height != filtered[j].Height {
				return filtered[i].Height > filtered[j].Height
			}
			return filtered[i].Bitrate > filtered[j].Bitrate
		})
	case SortBySize:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].FileSize > filtered[j].FileSize
		})
	case SortByBitrate:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Bitrate > filtered[j].Bitrate
		})
	}

	m.filtered = filtered
	m.table.SetRows(m.buildRows(filtered))
}

func (m FormatSelectorModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1)
	b.WriteString(titleStyle.Render(m.title) + "\n\n")

	statusParts := []string{fmt.Sprintf("Formats: %d", len(m.filtered))}
	switch m.filterMode {
	case FilterVideoOnly:
		statusParts = append(statusParts, "Filter: Video")
	case FilterAudioOnly:
		statusParts = append(statusParts, "Filter: Audio")
	}
	switch m.sortMode {
	case SortByQuality:
		statusParts = append(statusParts, "Sort: Quality")
	case SortBySize:
		statusParts = append(statusParts, "Sort: Size")
	case SortByBitrate:
		statusParts = append(statusParts, "Sort: Bitrate")
	}

	statusStyle := lipgloss.NewStyle().Faint(true)
	b.WriteString(statusStyle.Render(strings.Join(statusParts, " | ")) + "\n\n")
	b.WriteString(m.table.View() + "\n\n")
	b.WriteString(m.help.View(m.keys))

	return b.String()
}

func (m FormatSelectorModel) Selected() *adapter.Format {
	return m.selected
}

func (m FormatSelectorModel) Cancelled() bool {
	return m.cancelled
}
