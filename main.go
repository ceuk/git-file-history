package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type model struct {
	list         list.Model
	diff         string
	showDiff     bool
	err          error
	viewport     viewport.Model
	ready        bool
	commit       string
	filePath     string
	screenHeight int
}

func getGitCommits(filePath string) ([]list.Item, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%h %s||%cn, %ah", "--", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	items := make([]list.Item, len(lines))
	for i, line := range lines {
		details := strings.Split(line, "||")
		items[i] = item{title: details[0], desc: details[1]}
	}
	return items, nil
}

func getGitDiff(commit, filePath string) (string, error) {
	cmd := fmt.Sprintf("git show %s -- %s | git-split-diffs --color | sed '/─/,$!d' | tail -n +5", commit, filePath)
	output, err := exec.Command(os.Getenv("SHELL"), "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func newModel(filePath string) (model, error) {
	items, err := getGitCommits(filePath)
	if err != nil {
		return model{}, err
	}
	l := list.New(items, list.NewDefaultDelegate(), 200, 14)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.InfiniteScrolling = true
	l.Styles.Title = lipgloss.NewStyle().Bold(true)
	v := viewport.New(0, 0)
	v.HighPerformanceRendering = false
	return model{list: l, filePath: filePath, viewport: v}, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: git-file-history <file_path>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	m, err := newModel(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m,
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)
	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}

func initialRender(m model, screenWidth int, screenHeight int) {
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	// Handle key presses
	case tea.KeyMsg:
		switch msg.String() {
		case "g":
			m.viewport.GotoTop()
		case "G":
			m.viewport.GotoBottom()
			m.viewport, cmd = m.viewport.Update(msg)
		case "J":
			if m.showDiff {
				current := m.list.Index()
				length := len(m.list.Items())
				idx := min(current+1, length-1)
				m.list.Select(idx)
				selectedItem := m.list.SelectedItem().(item)
				m.commit = strings.Fields(selectedItem.Title())[0]
				diff, err := getGitDiff(m.commit, m.filePath)
				if err != nil {
					m.err = err
				} else {
					m.diff = diff
					m.viewport.SetContent(m.diff)
				}
			}
		case "K":
			if m.showDiff {
				current := m.list.Index()
				idx := max(current-1, 0)
				m.list.Select(idx)
				selectedItem := m.list.SelectedItem().(item)
				m.commit = strings.Fields(selectedItem.Title())[0]
				diff, err := getGitDiff(m.commit, m.filePath)
				if err != nil {
					m.err = err
				} else {
					m.diff = diff
					m.viewport.SetContent(m.diff)
				}
			}
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.showDiff {
				m.showDiff = false
			} else {
				return m, tea.Quit
			}
		case "j", "down":
			if !m.showDiff {
				m.list, cmd = m.list.Update(msg)
			}
		case "k", "up":
			if !m.showDiff {
				m.list, cmd = m.list.Update(msg)
			}
		case "enter":
			selectedItem := m.list.SelectedItem().(item)
			m.commit = strings.Fields(selectedItem.Title())[0]
			diff, err := getGitDiff(m.commit, m.filePath)
			if err != nil {
				m.err = err
			} else {
				m.diff = diff
				m.showDiff = true
			}
		}

	// handle window resizes
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		m.screenHeight = msg.Height
		listHeight := m.screenHeight
		if m.showDiff {
			listHeight = int(m.screenHeight / 4)
		}
		if !m.ready {
			m.ready = true
		}
		m.list.SetSize(msg.Width, listHeight)
		m.viewport.Width = msg.Width
		m.viewport.Height = m.screenHeight - int(m.screenHeight/4) - footerHeight - headerHeight
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) headerView() string {
	title := titleStyle.Render(m.filePath)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error()
	}
	if !m.ready {
		return "Loading..."
	}
	if m.showDiff {
		listHeight := int(m.screenHeight / 4)
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		m.list.SetHeight(listHeight)
		m.viewport.Height = m.screenHeight - listHeight - footerHeight - headerHeight
		m.viewport.SetContent(m.diff)
		return fmt.Sprintf("%s\n%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView(), m.list.View())
	} else {
		m.list.SetHeight(m.screenHeight)
		return m.list.View()
	}
}
