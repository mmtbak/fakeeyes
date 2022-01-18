package device

import (
	"time"

	"github.com/goodaye/fakeeyes/dao/rdb"
	"github.com/goodaye/fakeeyes/pkg/uuid"
	"github.com/goodaye/fakeeyes/protos/request"
	"github.com/goodaye/fakeeyes/service"
)

type Device struct {
	service.Entity
	rdb.Device
	rdb.DeviceSession
}

// 登陆
func Register(req request.DeviceInfo) (user Device, err error) {

	var dbdev rdb.Device

	var has bool
	session := rdb.NewSession()
	defer session.Close()

	has, err = session.Where("sn = ?", req.SN).Get(&dbdev)
	if err != nil {
		return
	}
	if !has {
		err = service.ErrorUserNotFound
		return
	}
	user.Device = dbdev
	user.Session = session
	err = user.CreateToken()
	if err != nil {
		return
	}
	return
}

// 创建token
func (dev *Device) CreateToken() (err error) {

	var dbsession rdb.DeviceSession
	session := dev.WithSession()

	has, err := session.Where("user_id = ? ", dev.Device.ID).Get(&dbsession)
	if err != nil {
		return err
	}
	token := uuid.CreateUUID()
	if !has {
		// 创建新的dbsssion
		newdbss := rdb.DeviceSession{
			UserID:     dev.Device.ID,
			Token:      token,
			ExpireTime: time.Now().Add(service.UserTokenExpireDuration),
		}
		_, err = session.Insert(newdbss)
		if err != nil {
			session.Rollback()
			return
		}
	} else {
		// 更新现有的
		updatedbss := rdb.DeviceSession{
			Token:      token,
			ExpireTime: time.Now().Add(service.UserTokenExpireDuration),
		}
		_, err = session.ID(dbsession.ID).Update(&updatedbss)
		if err != nil {
			session.Rollback()
			return
		}
	}
	_, err = session.Where("user_id = ? ", dev.Device.ID).Get(&dbsession)
	if err != nil {
		return err
	}
	session.Commit()
	dev.DeviceSession = dbsession
	return
}

// 检查登陆状态
func (dev *Device) CheckLoginStatus() (err error) {

	return nil
}
