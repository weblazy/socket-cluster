package model

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var AuthHandler = Auth{}

type Auth struct {
	Id        int64      `json:"id" gorm:"primary_key;type:INT AUTO_INCREMENT"`
	Username  string     `json:"username" gorm:"column:username;NOT NULL;default:'';comment:'用户名';type:VARCHAR(255)"`
	Avatar    string     `json:"avatar" gorm:"column:avatar;NOT NULL;default:'';comment:'头像';type:VARCHAR(255)"`
	Password  string     `json:"password" gorm:"column:password;NOT NULL;default:'';comment:'密码';type:VARCHAR(255)"`
	Status    int64      `json:"status" gorm:"column:status;NOT NULL;default:0;comment:'0禁用1正常';type:TINYINT"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*Auth) TableName() string {
	return "auth"
}

func (*Auth) Insert(db *gormx.DB, data *Auth) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*Auth) GetOne(where string, args ...interface{}) (*Auth, error) {
	var obj Auth
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*Auth) GetList(where string, args ...interface{}) ([]*Auth, error) {
	var list []*Auth
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*Auth) GetCount(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&Auth{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*Auth) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&Auth{}).Error
}

func (*Auth) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&Auth{}).Where(where, args...).Update(data).Error
}
