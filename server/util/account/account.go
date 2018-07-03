package account

type Account struct {
	PublicProjectsEnabled bool `json:"PublicProjectsEnabled"`
	HoursPerDay uint8 `json:"hoursPerDay"`
	DaysPerWeek uint8 `json:"daysPerWeek"`
}
