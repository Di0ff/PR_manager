package models

type Teams struct {
	Name  string  `gorm:"column:team_name;primaryKey"`
	Users []Users `gorm:"foreignKey:TeamName;references:Name"`
}
