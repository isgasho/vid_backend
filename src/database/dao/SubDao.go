package dao

import (
	"github.com/Aoi-hosizora/ahlib/xdi"
	"github.com/jinzhu/gorm"
	"github.com/vidorg/vid_backend/src/config"
	"github.com/vidorg/vid_backend/src/database"
	"github.com/vidorg/vid_backend/src/model/po"
)

type SubDao struct {
	Config  *config.ServerConfig `di:"~"`
	Db      *gorm.DB             `di:"~"`
	UserDao *UserDao             `di:"~"`

	PageSize        int32  `di:"-"`
	ColSubscribers  string `di:"-"`
	ColSubscribings string `di:"-"`
}

func NewSubDao(dic *xdi.DiContainer) *SubDao {
	repo := &SubDao{
		ColSubscribers:  "Subscribers",
		ColSubscribings: "Subscribings",
	}
	dic.Inject(repo)
	if xdi.HasNilDi(repo) {
		panic("Has nil di field")
	}

	repo.PageSize = repo.Config.MySqlConfig.PageSize
	return repo
}

func (s *SubDao) QuerySubscriberUsers(uid int32, page int32) (users []*po.User, count int32, status database.DbStatus) {
	// https://gorm.io/docs/many_to_many.html
	user := &po.User{Uid: uid}
	if !s.UserDao.Exist(uid) {
		return nil, 0, database.DbNotFound
	}
	count = int32(s.Db.Model(user).Association(s.ColSubscribers).Count()) // 开始关联模式
	s.Db.Limit(s.PageSize).Offset((page-1)*s.PageSize).Model(user).Related(&users, s.ColSubscribers)
	return users, count, database.DbSuccess
}

func (s *SubDao) QuerySubscribingUsers(uid int32, page int32) (users []*po.User, count int32, status database.DbStatus) {
	user := &po.User{Uid: uid}
	if !s.UserDao.Exist(uid) {
		return nil, 0, database.DbNotFound
	}
	count = int32(s.Db.Model(user).Association(s.ColSubscribings).Count())
	s.Db.Limit(s.PageSize).Offset((page-1)*s.PageSize).Model(user).Related(&users, s.ColSubscribings)
	return users, count, database.DbSuccess
}

func (s *SubDao) QuerySubCnt(uid int32) (subscribingCnt int32, subscriberCnt int32, status database.DbStatus) {
	if !s.UserDao.Exist(uid) {
		return 0, 0, database.DbNotFound
	}
	user := &po.User{Uid: uid}
	subscribingCnt = int32(s.Db.Model(user).Association(s.ColSubscribings).Count())
	subscriberCnt = int32(s.Db.Model(user).Association(s.ColSubscribers).Count())
	return subscribingCnt, subscriberCnt, database.DbSuccess
}

func (s *SubDao) SubscribeUser(meUid int32, toUid int32) database.DbStatus {
	if !s.UserDao.Exist(toUid) || !s.UserDao.Exist(meUid) {
		return database.DbNotFound
	}
	ass := s.Db.Model(&po.User{Uid: toUid}).Association(s.ColSubscribers).Append(&po.User{Uid: meUid})
	if ass.Error != nil {
		return database.DbFailed
	}
	return database.DbSuccess
}

func (s *SubDao) UnSubscribeUser(meUid int32, toUid int32) database.DbStatus {
	if !s.UserDao.Exist(toUid) || !s.UserDao.Exist(meUid) {
		return database.DbNotFound
	}
	ass := s.Db.Model(&po.User{Uid: toUid}).Association(s.ColSubscribers).Delete(&po.User{Uid: meUid})
	if ass.Error != nil {
		return database.DbFailed
	}
	return database.DbSuccess
}
