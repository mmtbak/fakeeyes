package service

import (
	"time"

	"github.com/goodaye/fakeeyes/dao/rdb"
	"github.com/goodaye/fakeeyes/pkg/copy"
	"github.com/goodaye/fakeeyes/pkg/uuid"
	"github.com/goodaye/fakeeyes/protos/request"
	"github.com/gorilla/websocket"
)

type Device struct {
	Entity
	rdb.Device
	rdb.DeviceSession
}

// 登陆
func RegisterDevice(req request.DeviceInfo) (dev Device, err error) {

	var dbdev rdb.Device
	var updatedev = rdb.Device{
		LastLogin: time.Now(),
	}
	var has bool
	session := rdb.NewSession()
	defer session.Close()

	has, err = session.Where("sn = ?", req.SN).Cols("id").Get(&dbdev)
	if err != nil {
		return
	}

	copy.StructCopy(req, &updatedev)
	updatedev.Status = DeviceState.Online

	if !has {
		uuid := uuid.CreateUUID()
		updatedev.UUID = uuid
		_, err = session.Insert(&updatedev)
		if err != nil {
			return
		}
	} else {
		_, err = session.ID(dbdev.ID).Update(&updatedev)
		if err != nil {
			return
		}
	}
	dev.Device = dbdev
	dev.Session = session
	err = dev.CreateToken()
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
			ExpireTime: time.Now().Add(UserTokenExpireDuration),
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
			ExpireTime: time.Now().Add(UserTokenExpireDuration),
		}
		_, err = session.ID(dbsession.ID).Update(&updatedbss)
		if err != nil {
			session.Rollback()
			return
		}
	}
	var ndbsession rdb.DeviceSession
	_, err = session.Where("user_id = ? ", dev.Device.ID).Get(&ndbsession)
	if err != nil {
		return err
	}
	session.Commit()
	dev.DeviceSession = ndbsession
	return
}

func (dev *Device) SendHeartBeat(req request.DeviceInfo) error {

	session := rdb.NewSession()
	defer session.Close()
	var update rdb.Device
	copy.StructCopy(req, &update)
	update.Status = DeviceState.Online
	update.LastLogin = time.Now()
	_, err := session.ID(dev.Device.ID).Update(&update)
	if err != nil {
		return err
	}
	session.Commit()
	return nil
}

// 设备链接到系统，保存链接
func (dev *Device) Connect(conn *websocket.Conn) error {
	devid := dev.Device.UUID
	DeviceConns[devid] = conn
	return nil
}

// 通过token 验证设备
func DeviceLoginByToken(token string) (dev *Device, err error) {

	var dbuser rdb.Device
	var dbusersession rdb.DeviceSession
	var has bool

	dev = &Device{}
	session := rdb.NewSession()
	defer session.Close()

	has, err = session.Where("token = ?", token).Get(&dbusersession)
	if err != nil {
		return
	}
	if !has {
		err = ErrorUserNotFound
		return
	}
	has, err = session.ID(dbusersession.UserID).Get(&dbuser)
	if err != nil {
		return
	}
	if !has {
		err = ErrorUserNotFound
		return
	}
	dev.Device = dbuser
	dev.Session = session
	dev.DeviceSession = dbusersession
	return
}

func DescribeDevice(req request.DescribeDevice) (dev *Device, err error) {

	return nil, err
}
