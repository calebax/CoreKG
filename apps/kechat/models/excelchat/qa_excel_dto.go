package excelchat

type TableDDL struct {
	Table       string `gorm:"column:Table"`
	CreateTable string `gorm:"column:Create Table"`
	View        string `gorm:"column:View"`
	CreateView  string `gorm:"column:Create View"`
}
