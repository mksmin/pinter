package model

import "time"

type Host struct {
	ID              string     `json:"id"`
	Alias           string     `json:"alias"`
	Hostname        string     `json:"hostname"`
	Port            int        `json:"port"`
	Username        string     `json:"username"`
	IdentityFile    string     `json:"identityFile"`
	Notes           string     `json:"notes"`
	Favorite        bool       `json:"favorite"`
	LastConnectedAt *time.Time `json:"lastConnectedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type HostInput struct {
	Alias        string
	Hostname     string
	Port         int
	Username     string
	IdentityFile string
	Notes        string
	Favorite     bool
}

type ConnectionHistory struct {
	ID            string    `json:"id"`
	HostID        string    `json:"hostId"`
	AliasSnapshot string    `json:"aliasSnapshot"`
	Command       string    `json:"command"`
	StartedAt     time.Time `json:"startedAt"`
	ExitStatus    *int      `json:"exitStatus"`
	TerminalApp   string    `json:"terminalApp"`
}
