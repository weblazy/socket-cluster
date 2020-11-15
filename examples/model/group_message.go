package model

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var GroupMessageHandler = GroupMessage{}

type GroupMessage struct {
	Id               int64      `json:"id" gorm:"primary_key;type:INT AUTO_INCREMENT"`
	GroupId          string     `json:"group_id" gorm:"column:group_id;NOT NULL;default:0;comment:'分组id';type:VARCHAR(255)"`
	GroupMessageType string     `json:"message_type" gorm:"column:message_type;NOT NULL;default:'';comment:'消息类型';type:VARCHAR(255)"`
	SendUid          string     `json:"send_uid" gorm:"column:send_uid;NOT NULL;default:'';comment:'发送者id';type:VARCHAR(255)"`
	Content          string     `json:"content" gorm:"column:content;NOT NULL;default:'';comment:'消息内容';type:VARCHAR(255)"`
	CreatedAt        time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt        *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*GroupMessage) TableName() string {
	return "group_message"
}

func (*GroupMessage) Insert(db *gormx.DB, data *GroupMessage) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*GroupMessage) GetOne(where string, args ...interface{}) (*GroupMessage, error) {
	var obj GroupMessage
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*GroupMessage) GetList(where string, args ...interface{}) ([]*GroupMessage, error) {
	var list []*GroupMessage
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*GroupMessage) GetCount(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&GroupMessage{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*GroupMessage) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&GroupMessage{}).Error
}

func (*GroupMessage) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&GroupMessage{}).Where(where, args...).Update(data).Error
}
