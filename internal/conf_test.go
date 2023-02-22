package internal

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/gdamore/tcell"
)

type confTestCase struct {
	name string
	conf string
	want map[int]*menuItem // expected menu items and functions
	wErr []string          // expected errors
}

func checkConf(t *testing.T, c confTestCase, cmdByText map[string]int) {
	var fp io.Reader
	if len(c.conf) > 0 {
		fp = strings.NewReader(c.conf)
	}
	if c.want == nil {
		c.want = menuItems
	}
	var shortcutNum, i int
	loadConf(func(menuKey, text, shortcut string, positive bool) {
		i++
		wantCmd := cmdByText[text]
		want := c.want[wantCmd]
		if want.shortcut != shortcut {
			t.Errorf("shortcut at %s should be '%s' but it's '%s'", text, want.shortcut,
				shortcut)
		}
		if want.menuKey != menuKey {
			t.Errorf("menuKey at %s should be '%s' but it's '%s'", text, want.menuKey, menuKey)
		}
		if want.text != text {
			t.Errorf("text should be '%s' but it's '%s'", want.text, text)
		}
		if want.positive != positive {
			t.Errorf("positive at %s should be '%t' but it's '%t'", text,
				want.positive, positive)
		}
		for _, key := range want.keys {
			if wantCmd != shortcuts[key] {
				t.Errorf("shortcut at %s should be '%d' but it's '%d'", text,
					wantCmd, shortcuts[key])
			}
			shortcutNum++
		}
	}, fp)
	showErrorsIf()
	checkMsg(t, "shortcut parse", c.wErr...)
	checkLen(t, "shortcut", i, len(c.want), nil, c.want)
	checkLen(t, "shortcut", shortcutNum, len(shortcuts), nil, shortcuts)
	shortcuts = map[tcell.Key]int{}
}

// TestConf tests reading configuration as keyboard shortcuts and bottom menu.
func TestConf(t *testing.T) {
	defer setup()()

	cases := []confTestCase{
		{
			name: "defaults",
		},
		{
			name: "missing_yaml_colon",
			conf: fmt.Sprintf("%[1]s\n  F3: %[2]s\n  F4: %[2]s", cShortcuts, titles[lst]),
			wErr: []string{fmt.Sprintf(errConfParse,
				"yaml: line 2: mapping values are not allowed in this context")},
		},
		{
			name: "empty_yaml",
			conf: fmt.Sprintf("%s:", cShortcuts),
		},
		{
			name: "yaml_command_not_found",
			conf: fmt.Sprintf("%s:\n  F3: notFound", cShortcuts),
			wErr: []string{fmt.Sprintf(errCmdNotFound, "notFound")},
		},
		{
			name: "yaml_commands_not_found",
			conf: fmt.Sprintf("%s:\n  F3: notFound\n  F4: notFound\n  F5: another", cShortcuts),
			wErr: []string{fmt.Sprintf(errCmdNotFound, "another, notFound")},
		},
		{
			name: "yaml_key_not_found",
			conf: fmt.Sprintf("%s:\n  notFound: %s", cShortcuts, titles[lst]),
			wErr: []string{fmt.Sprintf(errKeyNotFound, "notFound")},
		},
		{
			name: "double",
			conf: fmt.Sprintf("%[1]s:\n  F3: %[2]s\n  F4: %[2]s", cShortcuts, titles[lst]),
			want: map[int]*menuItem{
				cmdCancel:  {"F8", []tcell.Key{tcell.KeyF8}, "", menuCancel, false, true, cancel},
				cmdRefresh: {"F5", []tcell.Key{tcell.KeyF5}, "", menuRefresh, true, true, refresh},
				cmdSearch: {"F2", []tcell.Key{tcell.KeyF2}, srch, titles[srch],
					false, true, showSearch},
				cmdList: {"F3; F4", []tcell.Key{tcell.KeyF3, tcell.KeyF4}, lst,
					titles[lst], false, true, showList},
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
			checkConf(t, c, cmdByText)
		})
	}
}
