package model

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var UserGroupHandler = UserGroup{}

type UserGroup struct {
	Id             int64 `json:"id" gorm:"primary_key;type:INT AUTO_INCREMENT"`
	Uid            int64 `json:"uid" gorm:"column:uid;NOT NULL;default:0;comment:'用户id';type:INT"`
	GroupId        int64 `json:"group_id" gorm:"column:group_id;NOT NULL;default:0;comment:'分组id';type:INT"`
	MaxMsgId       int64 `json:"max_msg_id" gorm:"column:max_msg_id;NOT NULL;default:0;comment:'最大消息id';type:INT"`
	MaxReadMsgId   int64 `json:"max_read_msg_id" gorm:"column:max_read_msg_id;NOT NULL;default:0;comment:'已读最大消息id';type:INT"`
	MaxPulledMsgId int64 `json:"max_pulled_msg_id" gorm:"column:max_pulled_msg_id;NOT NULL;default:0;comment:'已拉取最大消息id';type:INT"`

	Status    int64      `json:"status" gorm:"column:status;NOT NULL;default:0;comment:'0免打扰1正常';type:TINYINT"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*UserGroup) TableName() string {
	return "user_group"
}

func UserGroupModel() *UserGroup {
	return &UserGroupHandler
}

func (this *UserGroup) Insert(db *gormx.DB, data *UserGroup) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (this *UserGroup) GetOne(where string, args ...interface{}) (*UserGroup, error) {
	var obj UserGroup
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (this *UserGroup) GetList(where string, args ...interface{}) ([]*UserGroup, error) {
	var list []*UserGroup
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (this *UserGroup) Count(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&UserGroup{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (this *UserGroup) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&UserGroup{}).Error
}

func (this *UserGroup) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&UserGroup{}).Where(where, args...).Update(data).Error
}
