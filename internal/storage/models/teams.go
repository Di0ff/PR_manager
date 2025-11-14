package models

type Teams struct {
	Name  string  `gorm:"column:team_name;primaryKey" json:"team_name"`
	Users []Users `gorm:"foreignKey:TeamName;references:Name" json:"members,omitempty"`
}
