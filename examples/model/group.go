package model

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var GroupHandler = Group{}

type Group struct {
	Id        int64      `json:"id" gorm:"primary_key;type:INT AUTO_INCREMENT"`
	GroupName string     `json:"group_name" gorm:"column:group_name;NOT NULL;default:'';comment:'分组名';type:VARCHAR(255)"`
	Avatar    string     `json:"avatar" gorm:"column:avatar;NOT NULL;default:'';comment:'头像';type:VARCHAR(255)"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*Group) TableName() string {
	return "group"
}

func GroupModel() *Group {
	return &GroupHandler
}

func (*Group) Insert(db *gormx.DB, data *Group) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*Group) GetOne(where string, args ...interface{}) (*Group, error) {
	var obj Group
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*Group) GetList(where string, args ...interface{}) ([]*Group, error) {
	var list []*Group
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*Group) Count(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&Group{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*Group) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&Group{}).Error
}

func (*Group) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&Group{}).Where(where, args...).Update(data).Error
}
