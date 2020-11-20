package model

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var FriendHandler = Friend{}

type Friend struct {
	Id        int64      `json:"id" gorm:"primary_key;type:INT AUTO_INCREMENT"`
	Uid       int64      `json:"uid" gorm:"column:uid;NOT NULL;default:0;comment:'发起请求用户id';type:INT"`
	OtherUid  int64      `json:"other_uid" gorm:"column:other_uid;NOT NULL;default:0;comment:'接收请求用户id';type:INT"`
	Status    int64      `json:"status" gorm:"column:status;NOT NULL;default:0;comment:'0未同意1已同意';type:TINYINT"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*Friend) TableName() string {
	return "friend"
}

func (*Friend) Insert(db *gormx.DB, data *Friend) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*Friend) GetOne(where string, args ...interface{}) (*Friend, error) {
	var obj Friend
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*Friend) GetList(where string, args ...interface{}) ([]*Friend, error) {
	var list []*Friend
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*Friend) GetCount(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&Friend{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*Friend) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&Friend{}).Error
}

func (*Friend) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&Friend{}).Where(where, args...).Update(data).Error
}
