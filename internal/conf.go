package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/gdamore/tcell"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	cmdStart = iota // menu commands
	cmdSearch
	cmdList
	cmdRefresh
	cmdSave
	cmdCancel
	cmdQuit
	cmdEnd
)

var (
	// shortcuts defines what to call in case of a keyboard shortcut press.
	shortcuts = map[tcell.Key]int{}
	// menuItems are the default menu items.
	menuItems = map[int]*menuItem{
		cmdCancel:  &menuItem{"F8", []tcell.Key{tcell.KeyF8}, "", menuCancel, false, true, cancel},
		cmdRefresh: &menuItem{"F5", []tcell.Key{tcell.KeyF5}, "", menuRefresh, true, true, refresh},
		cmdSearch: &menuItem{"F2", []tcell.Key{tcell.KeyF2}, kSearch, titles[kSearch],
			false, true, showSearch},
		cmdList: &menuItem{"F3", []tcell.Key{tcell.KeyF3}, kList, titles[kList], false, true, showList},
		cmdSave: &menuItem{"F6", []tcell.Key{tcell.KeyF6}, "", menuSave, true, true, save},
		cmdQuit: &menuItem{"Esc", []tcell.Key{tcell.KeyEsc}, "", menuQuit, false, true, quit},
	}
)

// menuItem defines default menu items' structure.
type menuItem struct {
	shortcut string
	keys     []tcell.Key
	menuKey  string
	text     string
	positive bool
	isDef    bool
	function func()
}

// initConf initializes configurations, only keyboard shortcuts for now.
func initConf(menuCb func(menuKey, text, shortcut string, isPositive bool)) {
	// loading config file
	if len(confPath) == 0 {
		loadConf(menuCb, nil)
		return
	}
	fp, err := os.Open(confPath)
	if err == nil {
		loadConf(menuCb, fp)
		return
	}
	lgr.Error("opening config file", zap.Error(errors.New(fmt.Sprintf(errConfPath, err.Error()))))
	loadConf(menuCb, nil)
}

// initConf initializes configurations, only keyboard shortcuts for now.
// Shows any errors in GUI in case of an error.
func loadConf(menuCb func(menuKey, text, shortcut string, isPositive bool), fp io.Reader) {
	defer func() {
		for i := cmdStart + 1; i < cmdEnd; i++ {
			m := menuItems[i]
			if m.isDef {
				shortcuts[m.keys[0]] = i
				menuCb(m.menuKey, m.text, m.shortcut, m.positive)
				continue
			}
			sort.Slice(m.keys, func(i, j int) bool {
				return m.keys[i] < m.keys[j]
			})
			m.shortcut = tcell.KeyNames[m.keys[0]]
			shortcuts[m.keys[0]] = i
			for _, key := range m.keys[1:] {
				shortcuts[key] = i
				m.shortcut = fmt.Sprintf("%s; %s", m.shortcut, tcell.KeyNames[key])
			}
			menuCb(m.menuKey, m.text, m.shortcut, m.positive)
		}
	}()

	if fp == nil {
		return
	}
	d := yaml.NewDecoder(fp)
	v := make(map[any]any)
	if err := d.Decode(&v); err != nil {
		writeErrorStr(fmt.Sprintf(errConfParse, err.Error()))
		return
	}

	// loading keyboard shortcuts from config file
	kc, ok := v[cShortcuts].(map[string]any)
	if !ok || len(kc) == 0 {
		return
	}

	mapTextToCmd := map[string]int{}
	for i := cmdStart + 1; i < cmdEnd; i++ {
		mapTextToCmd[menuItems[i].text] = i
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
			notFound[cmdStr] = es
			continue
		}

		m := menuItems[cmd]
		if !m.isDef {
			m.keys = append(m.keys, key)
			continue
		}
		m.isDef = false
		m.keys[0] = key
	}
	if len(kc) > 0 { // some more shortcuts not understood
		notFoundKeys := make([]string, 0, len(kc))
		for key := range kc {
			notFoundKeys = append(notFoundKeys, key)
		}
		writeErrorList(errKeyNotFound, notFoundKeys)
	}
	writeErrorMap(errCmdNotFound, notFound)
}
