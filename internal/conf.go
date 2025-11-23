package internal

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/gdamore/tcell/v2"
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

// menuItem defines default menu items' structure.
type menuItem struct {
	shortcut string
	keys     []tcell.Key
	menuKey  string
	text     string
	positive bool
	isDef    bool
	function func() error
}

// initConf initializes configurations, only keyboard shortcuts for now.
func initConf(menuCb func(menuKey, text, shortcut string, isPositive bool)) error {
	// loading config file
	if len(confPath) == 0 {
		return loadConf(menuCb, nil)
	}
	fp, err := os.Open(confPath)
	if err == nil {
		return loadConf(menuCb, fp)
	}
	return fmt.Errorf(errConfPath, err)
}

// initConf initializes configurations, only keyboard shortcuts for now.
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
		if err := writeErrorListGet(errKeyNotFound, notFoundKeys); err != nil {
			return err
		}
	}
	return writeErrorMap(errCmdNotFound, notFound)
}

func loadYamlConf(fp io.Reader) (map[string]any, error) {
	if fp == nil {
		return nil, nil
	}
	d := yaml.NewDecoder(fp)
	v := make(map[any]any)
	if err := d.Decode(&v); err != nil {
		return nil, fmt.Errorf(errConfParse, err)
	}

	// loading keyboard shortcuts from config file
	kc, ok := v[cShortcuts].(map[string]any)
	if !ok {
		return nil, errConfInvalid
	}

	return kc, nil
}

func saveShortcuts(menuCb func(menuKey, text, shortcut string, isPositive bool)) {
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
}
