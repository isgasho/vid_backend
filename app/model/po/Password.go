package po

type Password struct {
	Uid           int    `gorm:"primary_key"`
	EncryptedPass string `gorm:"type:char(48);not null"`

	User *User `gorm:"foreignkey:Uid"`

	GormTime
}

func (Password) TableName() string {
	return "tbl_password"
}