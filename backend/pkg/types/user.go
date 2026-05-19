package types

import "time"

type PfpPreset string

const (
	PfpBlue   PfpPreset = "blue"
	PfpGreen  PfpPreset = "green"
	PfpRed    PfpPreset = "red"
	PfpYellow PfpPreset = "yellow"
	PfpPurple PfpPreset = "purple"
	PfpOrange PfpPreset = "orange"
)

var validPfpPresets = map[PfpPreset]bool{
	PfpBlue: true, PfpGreen: true, PfpRed: true,
	PfpYellow: true, PfpPurple: true, PfpOrange: true,
}

func (p PfpPreset) Valid() bool {
	return validPfpPresets[p]
}

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Pfp       PfpPreset `json:"pfp"`
	CreatedAt time.Time `json:"createdAt"`
}
