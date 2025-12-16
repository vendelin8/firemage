package conf

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/lang"
	"gopkg.in/yaml.v3"
)

const (
	cmdStart = iota // menu commands
	CmdSearch
	CmdList
	CmdRefresh
	CmdSave
	CmdCancel
	CmdQuit
	cmdEnd
)

var ErrConfInvalid = errors.New(lang.ErrConfInvalidS)

var (
	ConfPath string
	LogPath  string
	UseEmu   bool
	KeyPath  string
)

// InitConf initializes configurations, only keyboard shortcuts for now.
func InitConf(menuCb func(menuKey, text, shortcut string, isPositive bool)) error {
	// loading config file
	if len(ConfPath) == 0 {
		return loadConf(menuCb, nil)
	}
	fp, err := os.Open(ConfPath)
	if err == nil {
		return loadConf(menuCb, fp)
	}
	return fmt.Errorf(lang.ErrConfPath, err)
}

// loadConf loads configurations, only keyboard shortcuts for now.
func loadConf(menuCb func(menuKey, text, shortcut string, isPositive bool), fp io.Reader) error {
	defer saveShortcuts(menuCb)

	kc, err := loadYamlConf(fp)
	if err != nil {
		return err
	}

	if len(kc) == 0 {
		return nil
	}

	mapTextToCmd := map[string]int{}
	for i := cmdStart + 1; i < cmdEnd; i++ {
		mapTextToCmd[common.MenuItems[i].Text] = i
	}

	notFound := map[string]struct{}{}           // list of not found key mappings
	for key, shortcut := range tcell.KeyNames { // iterating tcell key names to check matches
		text, ok := kc[shortcut]
		if !ok {
			continue
		}
		cmdStr := text.(string)
		delete(kc, shortcut)
		cmd, ok := mapTextToCmd[cmdStr]
		if !ok {
			notFound[cmdStr] = struct{}{}
			continue
		}

		m := common.MenuItems[cmd]
		if !m.IsDef {
			m.Keys = append(m.Keys, key)
			continue
		}
		m.IsDef = false
		m.Keys[0] = key
	}
	if len(kc) > 0 { // some more shortcuts not understood
		notFoundKeys := make([]string, 0, len(kc))
		for key := range kc {
			notFoundKeys = append(notFoundKeys, key)
		}
		sort.Strings(notFoundKeys)
		return fmt.Errorf(lang.ErrKeyNotFound, strings.Join(notFoundKeys, ", "))
	}
	if len(notFound) > 0 {
		notFoundCmds := make([]string, 0, len(notFound))
		for cmd := range notFound {
			notFoundCmds = append(notFoundCmds, cmd)
		}
		sort.Strings(notFoundCmds)
		return fmt.Errorf(lang.ErrCmdNotFound, strings.Join(notFoundCmds, ", "))
	}
	return nil
}

func loadYamlConf(fp io.Reader) (map[string]any, error) {
	if fp == nil {
		return nil, nil
	}
	d := yaml.NewDecoder(fp)
	v := make(map[any]any)
	if err := d.Decode(&v); err != nil {
		return nil, fmt.Errorf(lang.ErrConfParse, err)
	}

	// loading keyboard shortcuts from config file
	kc, ok := v[lang.CShortcuts].(map[string]any)
	if !ok {
		return nil, ErrConfInvalid
	}

	return kc, nil
}

func saveShortcuts(menuCb func(menuKey, text, shortcut string, isPositive bool)) {
	for i := cmdStart + 1; i < cmdEnd; i++ {
		m := common.MenuItems[i]
		if m.IsDef {
			common.Shortcuts[m.Keys[0]] = i
			menuCb(m.MenuKey, m.Text, m.Shortcut, m.Positive)
			continue
		}
		slices.Sort(m.Keys)
		m.Shortcut = tcell.KeyNames[m.Keys[0]]
		common.Shortcuts[m.Keys[0]] = i
		for _, key := range m.Keys[1:] {
			common.Shortcuts[key] = i
			m.Shortcut = fmt.Sprintf("%s; %s", m.Shortcut, tcell.KeyNames[key])
		}
		menuCb(m.MenuKey, m.Text, m.Shortcut, m.Positive)
	}
}
