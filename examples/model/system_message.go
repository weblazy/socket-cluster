package model

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var SystemMsgHandler = SystemMsg{}

type SystemMsg struct {
	Id         int64      `json:"id" gorm:"primary_key;type:INT AUTO_INCREMENT"`
	SendMsgId  int64      `json:"send_msg_id" gorm:"column:send_msg_id;NOT NULL;default:0;comment:'发送者消息id';type:INT"`
	NotifyUid  string     `json:"notify_uid" gorm:"column:notify_uid;NOT NULL;default:'';comment:'通知者id';type:VARCHAR(255)"`
	Username   string     `json:"username" gorm:"column:username;NOT NULL;default:'';comment:'用户名';type:VARCHAR(255)"`
	Avatar     string     `json:"avatar" gorm:"column:avatar;NOT NULL;default:'';comment:'头像';type:VARCHAR(255)"`
	ReceiveUid string     `json:"receive_uid" gorm:"column:receive_uid;NOT NULL;default:'';comment:'发送者id';type:VARCHAR(255)"`
	MsgType    string     `json:"msg_type" gorm:"column:msg_type;NOT NULL;default:'';comment:'消息类型:1添加好友2加群3退群';type:VARCHAR(255)"`
	SendUid    string     `json:"send_uid" gorm:"column:send_uid;NOT NULL;default:'';comment:'发送者id';type:VARCHAR(255)"`
	Content    string     `json:"content" gorm:"column:content;NOT NULL;default:'';comment:'消息内容';type:VARCHAR(255)"`
	Status     int64      `json:"status" gorm:"column:status;NOT NULL;default:0;comment:'0待验证,1同意,2已拒绝';type:TINYINT"`
	CreatedAt  time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt  *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*SystemMsg) TableName() string {
	return "system_msg"
}

func SystemMsgModel() *SystemMsg {
	return &SystemMsgHandler
}

func (*SystemMsg) Insert(db *gormx.DB, data *SystemMsg) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*SystemMsg) GetOne(where string, args ...interface{}) (*SystemMsg, error) {
	var obj SystemMsg
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*SystemMsg) GetList(where string, args ...interface{}) ([]*SystemMsg, error) {
	var list []*SystemMsg
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (this *SystemMsg) GetListPage(pageSize int64, order string, where string, args ...interface{}) ([]*SystemMsg, error) {
	var list []*SystemMsg
	db := Orm()
	return list, db.Table(this.TableName()).Where(where, args...).Order(order).Limit(pageSize).Find(&list).Error
}

func (*SystemMsg) Count(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&SystemMsg{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*SystemMsg) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&SystemMsg{}).Error
}

func (*SystemMsg) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&SystemMsg{}).Where(where, args...).Update(data).Error
}
