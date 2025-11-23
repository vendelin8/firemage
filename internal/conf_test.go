package internal

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/require"
)

type confTestCase struct {
	name string
	conf string
	want map[int]*menuItem // expected menu items and functions
	wErr string            // expected errors
}

// TestConf tests reading configuration as keyboard shortcuts and bottom menu.
func TestConf(t *testing.T) {
	defer setup(t)()

	cases := []confTestCase{
		{
			name: "defaults",
		},
		{
			name: "missing_yaml_colon",
			conf: fmt.Sprintf("%[1]s\n  F3: %[2]s\n  F4: %[2]s", cShortcuts, titles[pageLst]),
			wErr: fmt.Sprintf(errConfParse,
				"yaml: line 2: mapping values are not allowed in this context"),
		},
		{
			name: "empty_yaml",
			conf: fmt.Sprintf("%s:", cShortcuts),
			wErr: errConfInvalidStr,
		},
		{
			name: "yaml_command_not_found",
			conf: fmt.Sprintf("%s:\n  F3: notFound", cShortcuts),
			wErr: fmt.Sprintf(errCmdNotFound, "notFound"),
		},
		{
			name: "yaml_commands_not_found",
			conf: fmt.Sprintf("%s:\n  F3: notFound\n  F4: notFound\n  F5: another", cShortcuts),
			wErr: fmt.Sprintf(errCmdNotFound, "another, notFound"),
		},
		{
			name: "yaml_key_not_found",
			conf: fmt.Sprintf("%s:\n  notFound: %s", cShortcuts, titles[pageLst]),
			wErr: fmt.Sprintf(errKeyNotFound, "notFound"),
		},
		{
			name: "double",
			conf: fmt.Sprintf("%[1]s:\n  F3: %[2]s\n  F4: %[2]s", cShortcuts, titles[pageLst]),
			want: map[int]*menuItem{
				cmdCancel:  {"F8", []tcell.Key{tcell.KeyF8}, "", menuCancel, false, true, cancel},
				cmdRefresh: {"F5", []tcell.Key{tcell.KeyF5}, "", menuRefresh, true, true, refresh},
				cmdSearch: {"F2", []tcell.Key{tcell.KeyF2}, pageSrch, titles[pageSrch],
					false, true, showSearch},
				cmdList: {"F3; F4", []tcell.Key{tcell.KeyF3, tcell.KeyF4}, pageLst,
					titles[pageLst], false, true, showList},
				cmdSave: {"F6", []tcell.Key{tcell.KeyF6}, "", menuSave, true, true, save},
				cmdQuit: {"Esc", []tcell.Key{tcell.KeyEsc}, "", menuQuit, false, true, quit},
			},
		},
	}

	cmdByText := map[string]int{}
	for i := cmdStart + 1; i < cmdEnd; i++ {
		cmdByText[menuItems[i].text] = i
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var fp io.Reader
			if len(c.conf) > 0 {
				fp = strings.NewReader(c.conf)
			}
			if c.want == nil {
				c.want = menuItems
			}
			var shortcutNum, i int
			err := loadConf(func(menuKey, text, shortcut string, positive bool) {
				i++
				wantCmd := cmdByText[text]
				want := c.want[wantCmd]
				require.Equal(t, want.shortcut, shortcut)
				require.Equal(t, want.menuKey, menuKey)
				require.Equal(t, want.text, text)
				require.Equal(t, want.positive, positive)
				for _, key := range want.keys {
					require.Equal(t, wantCmd, shortcuts[key])
					shortcutNum++
				}
			}, fp)

			if len(c.wErr) > 0 {
				require.EqualError(t, err, c.wErr)
			} else {
				require.NoError(t, err)
			}

			require.Len(t, c.want, i, "shortcut")
			require.Len(t, shortcuts, shortcutNum, "shortcut")
			shortcuts = map[tcell.Key]int{}
		})
	}
}
