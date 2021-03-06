package model

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var AuthHandler = Auth{}

type Auth struct {
	Id                 int64      `json:"id" gorm:"primary_key;type:INT AUTO_INCREMENT"`
	Email              string     `json:"email" gorm:"column:email;NOT NULL;default:'';comment:'邮箱';type:VARCHAR(255)"`
	Username           string     `json:"username" gorm:"column:username;NOT NULL;default:'';comment:'用户名';type:VARCHAR(255)"`
	Avatar             string     `json:"avatar" gorm:"column:avatar;NOT NULL;default:'';comment:'头像';type:VARCHAR(255)"`
	Password           string     `json:"password" gorm:"column:password;NOT NULL;default:'';comment:'密码';type:VARCHAR(255)"`
	Status             int64      `json:"status" gorm:"column:status;NOT NULL;default:0;comment:'0禁用1正常';type:TINYINT"`
	MaxMsgId           int64      `json:"max_msg_id" gorm:"column:max_msg_id;NOT NULL;default:0;comment:'最大消息id';type:INT"`
	MaxPulledMsgId     int64      `json:"max_pulled_msg_id" gorm:"column:max_pulled_msg_id;NOT NULL;default:0;comment:'已拉取最大消息id';type:INT"`
	MaxReadSystemMsgId int64      `json:"max_read_system_msg_id" gorm:"column:max_read_system_msg_id;NOT NULL;default:0;comment:'已读最大系统消息id';type:INT"`
	CreatedAt          time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt          *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*Auth) TableName() string {
	return "auth"
}

func AuthModel() *Auth {
	return &AuthHandler
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

func (*Auth) Count(where string, args ...interface{}) (int, error) {
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

func (*Auth) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) (int64, error) {
	if db == nil {
		db = Orm()
	}
	db = db.Model(&Auth{}).Where(where, args...).Update(data)
	return db.RowsAffected, db.Error
}
