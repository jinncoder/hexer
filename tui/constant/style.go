package constant

import (
	"sync"

	serveConfig "github.com/archimoebius/hexer/cli/config/serve"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	ErrorHeaderText,
	TitleStyle lipgloss.Style
	ListDefaultItemStyle list.DefaultItemStyles
	InactiveTabStyle     lipgloss.Style
	ActiveTabStyle       lipgloss.Style
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
)

var lock = &sync.Mutex{}
var style *Styles = nil

func NewStyles(lg *lipgloss.Renderer) *Styles {
	lock.Lock()
	defer lock.Unlock()

	if style == nil {
		s := Styles{}

		s.Base = lg.NewStyle().
			Padding(1, 4, 0, 1).
			Align(lipgloss.Center, lipgloss.Center)

		s.HeaderText = lg.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: serveConfig.Setting.Theme.Style.HeaderText.Foreground.Light,
				Dark:  serveConfig.Setting.Theme.Style.HeaderText.Foreground.Dark,
			}).
			Bold(true).
			Padding(0, 1, 0, 2)

		s.Status = lg.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: serveConfig.Setting.Theme.Style.Status.Foreground.Light,
				Dark:  serveConfig.Setting.Theme.Style.Status.Foreground.Dark,
			}).
			PaddingLeft(1).
			MarginTop(1)

		s.StatusHeader = lg.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: serveConfig.Setting.Theme.Style.StatusHeader.Foreground.Light,
				Dark:  serveConfig.Setting.Theme.Style.StatusHeader.Foreground.Dark,
			}).
			Bold(true)

		s.ErrorHeaderText = s.HeaderText.
			Foreground(lipgloss.AdaptiveColor{
				Light: serveConfig.Setting.Theme.Style.ErrorHeaderText.Foreground.Light,
				Dark:  serveConfig.Setting.Theme.Style.ErrorHeaderText.Foreground.Dark,
			})

		s.TitleStyle = lg.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: serveConfig.Setting.Theme.Style.TitleStyle.Foreground.Light,
				Dark:  serveConfig.Setting.Theme.Style.TitleStyle.Foreground.Dark,
			}).
			Background(lipgloss.AdaptiveColor{
				Light: serveConfig.Setting.Theme.Style.TitleStyle.Background.Light,
				Dark:  serveConfig.Setting.Theme.Style.TitleStyle.Background.Dark,
			})

		s.InactiveTabStyle = lg.NewStyle().
			Border(inactiveTabBorder, true).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: serveConfig.Setting.Theme.Style.TabStyle.InactiveHighlightColor.BorderForeground.Light,
				Dark:  serveConfig.Setting.Theme.Style.TabStyle.InactiveHighlightColor.BorderForeground.Dark,
			}).Padding(0, 1)

		s.ActiveTabStyle = s.InactiveTabStyle.
			Border(activeTabBorder, true).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: serveConfig.Setting.Theme.Style.TabStyle.ActiveHighlightColor.BorderForeground.Light,
				Dark:  serveConfig.Setting.Theme.Style.TabStyle.ActiveHighlightColor.BorderForeground.Dark,
			})

		s.ListDefaultItemStyle = list.NewDefaultItemStyles()

		s.ListDefaultItemStyle.NormalTitle = lg.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: serveConfig.Setting.Theme.Style.ListDefaultItemStyle.NormalTitle.Foreground.Light,
				Dark:  serveConfig.Setting.Theme.Style.ListDefaultItemStyle.NormalTitle.Foreground.Dark,
			}).
			Padding(0, 0, 0, 2)

		s.ListDefaultItemStyle.NormalDesc = s.ListDefaultItemStyle.NormalTitle.Foreground(lipgloss.AdaptiveColor{
			Light: serveConfig.Setting.Theme.Style.ListDefaultItemStyle.NormalDescription.Foreground.Light,
			Dark:  serveConfig.Setting.Theme.Style.ListDefaultItemStyle.NormalDescription.Foreground.Dark,
		})

		s.ListDefaultItemStyle.SelectedTitle = s.ListDefaultItemStyle.SelectedTitle.Foreground(lipgloss.AdaptiveColor{
			Light: serveConfig.Setting.Theme.Style.ListDefaultItemStyle.SelectedTitle.Foreground.Light,
			Dark:  serveConfig.Setting.Theme.Style.ListDefaultItemStyle.SelectedTitle.Foreground.Dark,
		}).BorderForeground(lipgloss.AdaptiveColor{
			Light: serveConfig.Setting.Theme.Style.ListDefaultItemStyle.SelectedTitle.BorderForeground.Light,
			Dark:  serveConfig.Setting.Theme.Style.ListDefaultItemStyle.SelectedTitle.BorderForeground.Dark,
		})
		s.ListDefaultItemStyle.SelectedDesc = s.ListDefaultItemStyle.SelectedTitle.Foreground(lipgloss.AdaptiveColor{
			Light: serveConfig.Setting.Theme.Style.ListDefaultItemStyle.SelectedDescription.Foreground.Light,
			Dark:  serveConfig.Setting.Theme.Style.ListDefaultItemStyle.SelectedDescription.Foreground.Dark,
		})

		s.ListDefaultItemStyle.DimmedTitle = s.ListDefaultItemStyle.DimmedTitle.Foreground(lipgloss.AdaptiveColor{
			Light: serveConfig.Setting.Theme.Style.ListDefaultItemStyle.DimmedTitle.Foreground.Light,
			Dark:  serveConfig.Setting.Theme.Style.ListDefaultItemStyle.DimmedTitle.Foreground.Dark,
		})
		s.ListDefaultItemStyle.DimmedDesc = s.ListDefaultItemStyle.DimmedTitle.Foreground(lipgloss.AdaptiveColor{
			Light: serveConfig.Setting.Theme.Style.ListDefaultItemStyle.DimmedDescription.Foreground.Light,
			Dark:  serveConfig.Setting.Theme.Style.ListDefaultItemStyle.DimmedDescription.Foreground.Dark,
		})

		style = &s
	}

	return style
}

func AppBoundaryView(styles *Styles, width int, text string, wsc string) string {
	if wsc == "" {
		wsc = "_"
	}

	return lipgloss.PlaceHorizontal(
		width,
		lipgloss.Left,
		styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars(wsc),
		lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{
			Light: serveConfig.Setting.Theme.Style.Status.Foreground.Light,
			Dark:  serveConfig.Setting.Theme.Style.Status.Foreground.Dark,
		}),
	)
}

func AppErrorBoundaryView(styles *Styles, width int, text string) string {
	return lipgloss.PlaceHorizontal(
		width,
		lipgloss.Left,
		styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{
			Light: serveConfig.Setting.Theme.Style.ErrorHeaderText.Foreground.Light,
			Dark:  serveConfig.Setting.Theme.Style.ErrorHeaderText.Foreground.Dark,
		}),
	)
}
