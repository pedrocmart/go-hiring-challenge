package models

// Category represents a product category.
type Category struct {
	ID       uint      `gorm:"primaryKey"`
	Code     string    `gorm:"uniqueIndex;not null"`
	Name     string    `gorm:"not null"`
	Products []Product `gorm:"foreignKey:CategoryID"`
}

func (c *Category) TableName() string {
	return "categories"
}
