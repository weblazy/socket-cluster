package model

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var UserGroupMessageHandler = UserGroupMessage{}

type UserGroupMessage struct {
	Id             int64      `json:"id" gorm:"primary_key;type:INT AUTO_INCREMENT"`
	GroupMessageId int64      `json:"group_message_id" gorm:"column:group_message_id;NOT NULL;default:0;comment:'分组消息id';type:INT"`
	ReceiveUid     string     `json:"receive_uid" gorm:"column:receive_uid;NOT NULL;default:'';comment:'接收id';type:VARCHAR(255)"`
	Status         int64      `json:"status" gorm:"column:status;NOT NULL;default:0;comment:'0未已查看1已经查看';type:TINYINT"`
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt      *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*UserGroupMessage) TableName() string {
	return "group_message"
}

func (*UserGroupMessage) Insert(db *gormx.DB, data *UserGroupMessage) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*UserGroupMessage) GetOne(where string, args ...interface{}) (*UserGroupMessage, error) {
	var obj UserGroupMessage
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*UserGroupMessage) GetList(where string, args ...interface{}) ([]*UserGroupMessage, error) {
	var list []*UserGroupMessage
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*UserGroupMessage) GetCount(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&UserGroupMessage{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*UserGroupMessage) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&UserGroupMessage{}).Error
}

func (*UserGroupMessage) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&UserGroupMessage{}).Where(where, args...).Update(data).Error
}
