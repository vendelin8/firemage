package common

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/vendelin8/firemage/internal/lang"
)

const d1 = time.Hour * 24

var (
	ErrNoUsers      = errors.New(lang.ErrNoUsersS)
	ErrWrongDBClaim = errors.New(lang.ErrWrongDBClaimS)
)

var (
	// Frontend instance
	Fe FeIf
	// Firebase instance
	Fb FbIf

	MenuItems map[int]MenuItem
	Shortcuts = make(map[tcell.Key]int)

	defaultClaims = ClaimsMap{}
)

func init() {
	for _, perm := range AllPerms {
		defaultClaims[perm] = &Claim{}
	}
}

// MenuItem defines default menu items' structure.
type MenuItem struct {
	Shortcut string
	Keys     []tcell.Key
	MenuKey  string
	Text     string
	Positive bool
	IsDef    bool
	Function func() error
}

type Claim struct {
	Checked bool
	Date    *time.Time
}

func NewClaimFrom(a any) (*Claim, error) {
	c := &Claim{}
	if checked, ok := a.(bool); ok {
		c.Checked = checked
		return c, nil
	}

	if date, ok := a.(string); ok {
		d, err := time.Parse(DateFormat, date)
		if err != nil {
			return nil, fmt.Errorf("%w: %w %s", ErrWrongDBClaim, err, date)
		}
		c.Date = &d
		return c, nil
	}

	return nil, ErrWrongDBClaim
}

func (c *Claim) ToAny() any {
	if c.Date != nil {
		return c.Date.Format(DateFormat)
	}

	return c.Checked
}

func (c *Claim) FormatDate() string {
	if c.Date == nil {
		return "<???>"
	}

	return c.Date.Format(DateFormat)
}

func (c *Claim) IsZero() bool {
	return c.Date == nil && !c.Checked
}

func (c *Claim) String() string {
	if c.Date != nil {
		return fmt.Sprintf("Claim(%s)", c.Date)
	}

	return fmt.Sprintf("Claim(%t)", c.Checked)
}

func (c *Claim) Differs(d *Claim) bool {
	if c.Checked != d.Checked || (c.Date == nil) != (d.Date == nil) {
		return true
	}

	return d.Date != nil && !c.Date.Truncate(d1).Equal(d.Date.Truncate(d1))
}

func (c *Claim) DiffersType(d *Claim) bool {
	return (d.Date == nil) != (c.Date == nil)
}

type ClaimsMap map[string]*Claim

func NewClaimsMap() *ClaimsMap {
	cs := make(ClaimsMap, len(defaultClaims))
	maps.Copy(cs, defaultClaims)
	return &cs
}

func NewClaimsMapFrom(as map[string]any) (*ClaimsMap, error) {
	var err error

	cs := *NewClaimsMap()

	for key, a := range as {
		if cs[key], err = NewClaimFrom(a); err != nil {
			return nil, err
		}
	}

	return &cs, nil
}

func (c *ClaimsMap) String() string {
	var b strings.Builder
	b.WriteString("ClaimsMap(")

	for key, value := range *c {
		if b.Len() > 0 {
			b.WriteString("; ")
		}
		b.WriteByte('\'')
		b.WriteString(key)
		b.WriteString("': ")
		b.WriteString(value.String())
	}

	b.WriteByte(')')
	return b.String()
}
