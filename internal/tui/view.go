package tui

import (
	"fmt"
	"strings"
	"time"

	"surge/internal/utils"

	"github.com/charmbracelet/lipgloss"
)

// Define the Layout Ratios
const (
	ListWidthRatio = 0.6 // List takes 60% width
)

func (m RootModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// === Handle Modal States First ===
	// These overlays sit on top of the dashboard or replace it

	if m.state == InputState {
		labelStyle := lipgloss.NewStyle().Width(10).Foreground(ColorGray)
		// Centered popup - compact layout
		hintStyle := lipgloss.NewStyle().MarginLeft(1).Foreground(ColorGray) // Dimmed
		if m.focusedInput == 1 {
			hintStyle = lipgloss.NewStyle().MarginLeft(1).Foreground(ColorNeonPink) // Highlighted
		}
		pathLine := lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Render("Path:"),
			m.inputs[1].View(),
			hintStyle.Render("[Tab] Browse"),
		)

		popup := lipgloss.JoinVertical(lipgloss.Left,
			TitleStyle.Render("ADD DOWNLOAD"),
			"",
			lipgloss.JoinHorizontal(lipgloss.Left, labelStyle.Render("URL:"), m.inputs[0].View()),
			pathLine,
			lipgloss.JoinHorizontal(lipgloss.Left, labelStyle.Render("Filename:"), m.inputs[2].View()),
			"",
			lipgloss.NewStyle().Foreground(ColorGray).Render("[Enter] Start  [Esc] Cancel"),
		)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			PaneStyle.Width(60).Padding(1, 2).Render(popup),
		)
	}

	if m.state == FilePickerState {
		pickerContent := lipgloss.JoinVertical(lipgloss.Left,
			TitleStyle.Render("SELECT DIRECTORY"),
			"",
			lipgloss.NewStyle().Foreground(ColorGray).Render(m.filepicker.CurrentDirectory),
			"",
			m.filepicker.View(),
			"",
			lipgloss.NewStyle().Foreground(ColorGray).Render("[.] Select Here  [H] Downloads  [Enter] Open  [Esc] Cancel"),
		)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			PaneStyle.Width(60).Padding(1, 2).Render(pickerContent),
		)
	}

	if m.state == DuplicateWarningState {
		warningContent := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(ColorNeonPink).Bold(true).Render("⚠ DUPLICATE DETECTED"),
			"",
			lipgloss.NewStyle().Foreground(ColorNeonPurple).Bold(true).Render(truncateString(m.duplicateInfo, 50)),
			"",
			lipgloss.NewStyle().Foreground(ColorGray).Render("[C] Continue  [F] Focus Existing  [X] Cancel"),
		)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(ColorNeonPink).
				Padding(1, 3).
				Render(warningContent),
		)
	}

	// === MAIN DASHBOARD LAYOUT ===

	availableHeight := m.height - 2 // Margin
	availableWidth := m.width - 4   // Margin

	// Top Row Height (Logo + Graph)
	topHeight := 9

	// Bottom Row Height (List + Details)
	bottomHeight := availableHeight - topHeight - 1
	if bottomHeight < 10 {
		bottomHeight = 10
	} // Min height

	// Column Widths
	leftWidth := int(float64(availableWidth) * ListWidthRatio)
	rightWidth := availableWidth - leftWidth - 2 // -2 for spacing

	// --- SECTION 1: HEADER & LOGO (Top Left) ---
	logoText := `
 ██████  ██    ██ ██████   ██████  ███████ 
██       ██    ██ ██   ██ ██       ██      
███████  ██    ██ ██████  ██   ███ █████   
     ██  ██    ██ ██   ██ ██    ██ ██      
███████   ██████  ██   ██  ██████  ███████`

	// Create the header stats
	active, queued, downloaded := m.CalculateStats()
	statsText := fmt.Sprintf("Active: %d  •  Queued: %d  •  Done: %d", active, queued, downloaded)

	headerContent := lipgloss.JoinVertical(lipgloss.Left,
		LogoStyle.Render(logoText),
		lipgloss.NewStyle().Foreground(ColorGray).Render(statsText),
	)

	// Use PaneStyle for consistent borders with the graph box
	headerBox := PaneStyle.
		Width(leftWidth).
		Height(topHeight).
		Render(headerContent)

	// --- SECTION 2: SPEED GRAPH (Top Right) ---
	// Render the Sparkline (account for borders: 2 for left/right border, 2 for padding)
	graphContentWidth := rightWidth - 4
	if graphContentWidth < 10 {
		graphContentWidth = 10
	}
	graphContentHeight := topHeight - 4
	if graphContentHeight < 2 {
		graphContentHeight = 2
	}
	graphContent := renderSparkline(m.SpeedHistory, graphContentWidth, graphContentHeight)

	// Get current speed
	currentSpeed := 0.0
	if len(m.SpeedHistory) > 0 {
		currentSpeed = m.SpeedHistory[len(m.SpeedHistory)-1]
	}
	currentSpeedStr := fmt.Sprintf("%.2f MB/s", currentSpeed)

	graphBox := GraphStyle.
		Width(rightWidth).
		Height(topHeight).
		Render(lipgloss.JoinVertical(lipgloss.Right,
			graphContent,
			lipgloss.NewStyle().Foreground(ColorNeonPink).Bold(true).Render(currentSpeedStr),
		))

	// --- SECTION 3: DOWNLOAD LIST (Bottom Left) ---
	// Tab Bar
	tabBar := renderTabs(m.activeTab)

	// Calculate list height (Total bottom height - Tab bar height)
	listHeight := bottomHeight - 2 // Tabs take ~2 lines
	if listHeight < 5 {
		listHeight = 5
	}

	// Render List
	listContent := m.renderDownloadList(leftWidth-4, listHeight)

	listBox := ListStyle.
		Width(leftWidth).
		Height(bottomHeight).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			tabBar,
			listContent,
		))

	// --- SECTION 4: DETAILS PANE (Bottom Right) ---
	var detailContent string
	filtered := m.getFilteredDownloads()
	if len(filtered) > 0 && m.cursor < len(filtered) {
		detailContent = renderFocusedDetails(filtered[m.cursor], rightWidth-4)
	} else {
		detailContent = lipgloss.Place(rightWidth-4, bottomHeight-4, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(ColorGray).Render("No Download Selected"))
	}

	detailBox := DetailStyle.
		Width(rightWidth).
		Height(bottomHeight).
		Render(detailContent)

	// --- ASSEMBLY ---

	// Top Row
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, headerBox, graphBox)

	// Bottom Row
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, listBox, detailBox)

	// Full Layout
	return lipgloss.JoinVertical(lipgloss.Left,
		topRow,
		bottomRow,
		lipgloss.NewStyle().Foreground(ColorGray).Padding(0, 1).Render(" [G] Add  [P] Pause  [D] Delete  [Q] Quit"),
	)
}

// Helper to render the list
func (m RootModel) renderDownloadList(w, h int) string {
	var rows []string
	filtered := m.getFilteredDownloads()

	// Item height is roughly 2 lines
	itemHeight := 2
	visibleCount := h / itemHeight
	if visibleCount < 1 {
		visibleCount = 1
	}

	start := m.scrollOffset
	end := start + visibleCount
	if end > len(filtered) {
		end = len(filtered)
	}

	for i := start; i < end; i++ {
		if i >= len(filtered) {
			break
		}
		d := filtered[i]
		isSelected := (i == m.cursor)

		// Compact Row Style
		style := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true). // Left border only
			BorderForeground(ColorGray).
			Padding(0, 1).
			Width(w)

		if isSelected {
			style = style.
				BorderForeground(ColorNeonPink).
				Background(lipgloss.Color("#44475a")) // Highlight bg
		}

		// Progress
		pct := 0.0
		if d.Total > 0 {
			pct = float64(d.Downloaded) / float64(d.Total)
		}
		progStr := fmt.Sprintf("%.0f%%", pct*100)
		statusIcon := "⬇"
		if d.done {
			statusIcon = "✔"
		}
		if d.paused {
			statusIcon = "⏸"
		}
		if d.err != nil {
			statusIcon = "✖"
		}

		title := truncateString(d.Filename, w-20)

		row := lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Width(3).Foreground(ColorNeonPurple).Render(statusIcon),
			lipgloss.NewStyle().Width(w-15).Render(title),
			lipgloss.NewStyle().Foreground(ColorNeonCyan).Render(progStr),
		)

		rows = append(rows, style.Render(row))
	}

	// Fill empty space
	for len(rows) < visibleCount {
		rows = append(rows, "")
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// Helper to render the detailed info pane
func renderFocusedDetails(d *DownloadModel, w int) string {
	pct := 0.0
	if d.Total > 0 {
		pct = float64(d.Downloaded) / float64(d.Total)
	}

	// Use your existing Progress Bar but styled
	d.progress.Width = w - 4
	progView := d.progress.ViewAs(pct)

	return lipgloss.JoinVertical(lipgloss.Left,
		PaneTitleStyle.Render("FILE DETAILS"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, StatsLabelStyle.Render("Filename:"), StatsValueStyle.Render(truncateString(d.Filename, w-15))),
		lipgloss.JoinHorizontal(lipgloss.Left, StatsLabelStyle.Render("Status:"), StatsValueStyle.Render(getDownloadStatus(d))),
		lipgloss.JoinHorizontal(lipgloss.Left, StatsLabelStyle.Render("Size:"), StatsValueStyle.Render(fmt.Sprintf("%s / %s", utils.ConvertBytesToHumanReadable(d.Downloaded), utils.ConvertBytesToHumanReadable(d.Total)))),
		"",
		lipgloss.NewStyle().Foreground(ColorNeonCyan).Render("PROGRESS"),
		progView,
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, StatsLabelStyle.Render("Speed:"), StatsValueStyle.Render(fmt.Sprintf("%.2f MB/s", d.Speed/Megabyte))),
		lipgloss.JoinHorizontal(lipgloss.Left, StatsLabelStyle.Render("Conns:"), StatsValueStyle.Render(fmt.Sprintf("%d", d.Connections))),
		lipgloss.JoinHorizontal(lipgloss.Left, StatsLabelStyle.Render("Elapsed:"), StatsValueStyle.Render(d.Elapsed.Round(time.Second).String())),
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, StatsLabelStyle.Render("URL:"), lipgloss.NewStyle().Foreground(ColorGray).Render(truncateString(d.URL, w-10))),
	)
}

func getDownloadStatus(d *DownloadModel) string {
	if d.err != nil {
		return "Error"
	}
	if d.paused {
		return "Paused"
	}
	if d.done {
		return "Completed"
	}
	if d.Speed == 0 && d.Downloaded == 0 {
		return "Queued"
	}
	return "Downloading"
}

// Simple Sparkline Generator using Braille patterns
func renderSparkline(data []float64, w, h int) string {
	if len(data) == 0 {
		return ""
	}

	// Find max for scaling
	max := 0.0
	for _, v := range data {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		max = 1
	}

	// Braille characters
	// distinct levels: ' ', '⡀', '⣀', '⣄', '⣤', '⣦', '⣶', '⣷', '⣿'
	levels := []rune{' ', '⡀', '⣀', '⣄', '⣤', '⣦', '⣶', '⣷', '⣿'}

	// Sample the data to fit width
	// We want to show the latest data at the right
	// If we have more pixels (w) than data, we stretch? No, sparklines usually just show available data.

	// Actually, let's just map data points to character columns.
	// We have 40 history points, width might be ~60 chars.

	var s strings.Builder

	// Ensure we don't go out of bounds
	startIndex := 0
	if len(data) > w {
		startIndex = len(data) - w
	}

	visibleData := data[startIndex:]

	for _, val := range visibleData {
		levelIdx := int((val / max) * float64(len(levels)-1))
		if levelIdx < 0 {
			levelIdx = 0
		}
		if levelIdx >= len(levels) {
			levelIdx = len(levels) - 1
		}
		s.WriteRune(levels[levelIdx])
	}
	// Fill remaining width if any (pad left)
	// But usually we just return what we have

	return lipgloss.NewStyle().Foreground(ColorNeonPink).Render(s.String())
}

func (m RootModel) calcTotalSpeed() float64 {
	total := 0.0
	for _, d := range m.downloads {
		total += d.Speed
	}
	return total / Megabyte
}

func (m RootModel) CalculateStats() (active, queued, downloaded int) {
	for _, d := range m.downloads {
		if d.done {
			downloaded++
		} else if d.Speed > 0 {
			active++
		} else {
			queued++
		}
	}
	return
}

func truncateString(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return string(runes[:i]) + "..."
	}
	return s
}

func renderTabs(active int) string {
	tabs := []string{"Queued", "Active", "Done"}
	var rendered []string
	for i, t := range tabs {
		var style lipgloss.Style
		if i == active {
			style = ActiveTabStyle
		} else {
			style = TabStyle
		}
		rendered = append(rendered, style.Render(t))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}
