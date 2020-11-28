package model

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

type UserMsg struct {
	Id         int64      `json:"id" gorm:"primary_key;type:INT AUTO_INCREMENT"`
	NotifyUid  string     `json:"notify_uid" gorm:"column:notify_uid;NOT NULL;default:'';comment:'通知者id';type:VARCHAR(255)"`
	Username   string     `json:"username" gorm:"column:username;NOT NULL;default:'';comment:'用户名';type:VARCHAR(255)"`
	Avatar     string     `json:"avatar" gorm:"column:avatar;NOT NULL;default:'';comment:'头像';type:VARCHAR(255)"`
	ReceiveUid string     `json:"receive_uid" gorm:"column:receive_uid;NOT NULL;default:'';comment:'发送者id';type:VARCHAR(255)"`
	MsgType    string     `json:"message_type" gorm:"column:message_type;NOT NULL;default:'';comment:'消息类型';type:VARCHAR(255)"`
	SendUid    string     `json:"send_uid" gorm:"column:send_uid;NOT NULL;default:'';comment:'发送者id';type:VARCHAR(255)"`
	Content    string     `json:"content" gorm:"column:content;NOT NULL;default:'';comment:'消息内容';type:VARCHAR(255)"`
	Status     int64      `json:"status" gorm:"column:status;NOT NULL;default:0;comment:'0未已查看1已经查看';type:TINYINT"`
	CreatedAt  time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt  *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
	table      string     `json:"-"`
}

var userMsgMap map[int64]*UserMsg = make(map[int64]*UserMsg)

// @desc
// @auth liuguoqiang 2020-11-26
// @param
// @return
func UserMsgModel(index int64) *UserMsg {
	return userMsgMap[index]
}

func (this *UserMsg) TableName() string {
	return this.table
}
func (*UserMsg) Insert(db *gormx.DB, data *UserMsg) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*UserMsg) GetOne(where string, args ...interface{}) (*UserMsg, error) {
	var obj UserMsg
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*UserMsg) GetList(where string, args ...interface{}) ([]*UserMsg, error) {
	var list []*UserMsg
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*UserMsg) GetListPage(pageSize int64, where string, args ...interface{}) ([]*UserMsg, error) {
	var list []*UserMsg
	db := Orm()
	return list, db.Where(where, args...).Limit(pageSize).Find(&list).Error
}

func (*UserMsg) GetCount(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&UserMsg{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*UserMsg) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&UserMsg{}).Error
}

func (*UserMsg) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&UserMsg{}).Where(where, args...).Update(data).Error
}
