package ip2region

import (
	"bytes"
	"context"
	"fmt"
)

type Provider interface {
	Search(ctx context.Context, ip string, langs ...string) (*Result, error)
	Update(ctx context.Context) error
	Close() error
}

type Result struct {
	IP          string `json:"ip,omitempty"`
	Continent   Name   `json:"continent,omitempty"`
	Country     Name   `json:"country,omitempty"`
	Subdivision Name   `json:"subdivision,omitempty"`
	City        Name   `json:"city,omitempty"`
	ISP         string `json:"isp,omitempty"`
}

type Name struct {
	Name string `json:"name,omitempty"`
	Code string `json:"code,omitempty"`
	ID   uint   `json:"id,omitempty"`
}

func NewName(name, code string, id uint) Name {
	return Name{Name: name, Code: code, ID: id}
}

func (n Name) String() string {
	if n.Name != "" {
		if n.Code != "" {
			return fmt.Sprintf("%s(%s)", n.Name, n.Code)
		} else if n.ID > 0 {
			return fmt.Sprintf("%s(%d)", n.Name, n.ID)
		} else {
			return n.Name
		}
	} else {
		if n.Code != "" {
			return n.Code
		} else if n.ID > 0 {
			return fmt.Sprintf("%d", n.ID)
		} else {
			return ""
		}
	}
}

func (r Result) InfoText() string {
	var w bytes.Buffer

	if s := r.Continent.String(); s != "" {
		w.WriteString(s)
		w.WriteString(", ")
	}

	if s := r.Country.String(); s != "" {
		w.WriteString(s)
		w.WriteString(", ")
	}

	if s := r.Subdivision.String(); s != "" {
		w.WriteString(s)
		w.WriteString(", ")
	}

	if s := r.City.String(); s != "" {
		w.WriteString(s)
		w.WriteString(", ")
	}

	if r.ISP != "" {
		w.WriteString(r.ISP)
		w.WriteString(", ")
	}

	if w.Len() > 2 {
		w.Truncate(w.Len() - 2)
	}

	return w.String()
}

func (r Result) String() string {
	return fmt.Sprintf("%15s: %s", r.IP, r.InfoText())
}
