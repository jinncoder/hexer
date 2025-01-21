package config

type Theme struct {
	Style struct {
		TitleStyle struct {
			Foreground struct {
				Light string `mapstructure:"light"`
				Dark  string `mapstructure:"dark"`
			} `mapstructure:"foreground"`
			Background struct {
				Light string `mapstructure:"light"`
				Dark  string `mapstructure:"dark"`
			} `mapstructure:"background"`
		} `mapstructure:"title_style"`
		ErrorHeaderText struct {
			Foreground struct {
				Light string `mapstructure:"light"`
				Dark  string `mapstructure:"dark"`
			} `mapstructure:"foreground"`
		} `mapstructure:"error_header_text"`
		HeaderText struct {
			Foreground struct {
				Light string `mapstructure:"light"`
				Dark  string `mapstructure:"dark"`
			} `mapstructure:"foreground"`
		} `mapstructure:"header_text"`
		Status struct {
			Foreground struct {
				Light string `mapstructure:"light"`
				Dark  string `mapstructure:"dark"`
			} `mapstructure:"foreground"`
		} `mapstructure:"status"`
		StatusHeader struct {
			Foreground struct {
				Light string `mapstructure:"light"`
				Dark  string `mapstructure:"dark"`
			} `mapstructure:"foreground"`
		} `mapstructure:"status_header"`
		TabStyle struct {
			InactiveHighlightColor struct {
				BorderForeground struct {
					Light string `mapstructure:"light"`
					Dark  string `mapstructure:"dark"`
				} `mapstructure:"border_foreground"`
			} `mapstructure:"inactive_highlight_color"`
			ActiveHighlightColor struct {
				BorderForeground struct {
					Light string `mapstructure:"light"`
					Dark  string `mapstructure:"dark"`
				} `mapstructure:"border_foreground"`
			} `mapstructure:"active_highlight_color"`
		} `mapstructure:"tab_style"`
		ListDefaultItemStyle struct {
			NormalTitle struct {
				Foreground struct {
					Light string `mapstructure:"light"`
					Dark  string `mapstructure:"dark"`
				} `mapstructure:"foreground"`
			} `mapstructure:"normal_title"`
			NormalDescription struct {
				Foreground struct {
					Light string `mapstructure:"light"`
					Dark  string `mapstructure:"dark"`
				} `mapstructure:"foreground"`
			} `mapstructure:"normal_description"`
			SelectedTitle struct {
				Foreground struct {
					Light string `mapstructure:"light"`
					Dark  string `mapstructure:"dark"`
				} `mapstructure:"foreground"`
				BorderForeground struct {
					Light string `mapstructure:"light"`
					Dark  string `mapstructure:"dark"`
				} `mapstructure:"border_foreground"`
			} `mapstructure:"selected_title"`
			SelectedDescription struct {
				Foreground struct {
					Light string `mapstructure:"light"`
					Dark  string `mapstructure:"dark"`
				} `mapstructure:"foreground"`
			} `mapstructure:"selected_description"`
			DimmedTitle struct {
				Foreground struct {
					Light string `mapstructure:"light"`
					Dark  string `mapstructure:"dark"`
				} `mapstructure:"foreground"`
			} `mapstructure:"dimmed_title"`
			DimmedDescription struct {
				Foreground struct {
					Light string `mapstructure:"light"`
					Dark  string `mapstructure:"dark"`
				} `mapstructure:"foreground"`
			} `mapstructure:"dimmed_description"`
		} `mapstructure:"list_default_item_style"`
	} `mapstructure:"style"`
}

func DefaultTheme() *Theme {
	return new(Theme)
}
