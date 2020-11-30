package model

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var SmsCodeHandler = SmsCode{}

type SmsCode struct {
	Id        int64      `json:"id" gorm:"primary_key;type:INT AUTO_INCREMENT"`
	Email     string     `json:"email" gorm:"column:email;NOT NULL;default:'';comment:'邮箱';type:VARCHAR(255)"`
	Code      string     `json:"code" gorm:"column:code;NOT NULL;default:'';comment:'验证码';type:VARCHAR(255)"`
	Status    int64      `json:"status" gorm:"column:status;NOT NULL;default:0;comment:'0未发送1已发送2已使用';type:TINYINT"`
	MsgType   int64      `json:"msg_type" gorm:"column:msg_type;NOT NULL;default:0;comment:'消息类型1注册消息2修改密码';type:TINYINT"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*SmsCode) TableName() string {
	return "sms_code"
}

func SmsCodeModel() *SmsCode {
	return &SmsCodeHandler
}

func (*SmsCode) Insert(db *gormx.DB, data *SmsCode) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*SmsCode) GetOne(where string, args ...interface{}) (*SmsCode, error) {
	var obj SmsCode
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*SmsCode) GetList(where string, args ...interface{}) ([]*SmsCode, error) {
	var list []*SmsCode
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*SmsCode) Count(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&SmsCode{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*SmsCode) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&SmsCode{}).Error
}

func (*SmsCode) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&SmsCode{}).Where(where, args...).Update(data).Error
}
