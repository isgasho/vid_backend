package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"vid/app/controller/exception"
	"vid/app/database"
	"vid/app/database/dao"
	"vid/app/middleware"
	"vid/app/model"
	"vid/app/model/dto"
	"vid/app/model/enum"
	"vid/app/model/vo"
)

type userCtrl struct{}

var UserCtrl = new(userCtrl)

// @Router 				/user?page [GET] [Auth]
// @Summary 			查询所有用户
// @Description 		管理员查询所有用户，返回分页数据，Admin
// @Param 				Authorization header string true "用户登录令牌"
// @Param 				page query integer false "分页"
// @Accept 				multipart/form-data
// @ErrorCode			401 authorization failed
// @ErrorCode			401 token has expired
// @ErrorCode			401 need admin authority
/* @Success 200 		{
							"code": 200,
							"message": "Success",
							"data": {
								"count": 1,
								"page": 1,
								"data": [
									{
										"uid": 1,
										"username": "User1",
										"sex": "male",
										"profile": "",
										"avatar_url": "",
										"birth_time": "2000-01-01",
										"authority": "admin"
									}
								]
							}
 						} */
func (u *userCtrl) QueryAllUsers(c *gin.Context) {
	pageString := c.Query("page")
	page, err := strconv.Atoi(pageString)
	if err != nil || page == 0 {
		page = 1
	}
	users, count := dao.UserDao.QueryAll(page)
	c.JSON(http.StatusOK,
		dto.Result{}.Ok().SetPage(count, page, users))
}

// @Router 				/user/{uid} [GET]
// @Summary 			查询用户
// @Description 		查询用户信息
// @Param 				uid path integer true "用户id"
// @Accept 				multipart/form-data
// @ErrorCode			400 request route param error
// @ErrorCode			404 user not found
/* @Success 200 		{
							"code": 200,
							"message": "Success",
							"data": {
								"user": {
									"uid": 10,
									"username": "aoihosizora",
									"sex": "unknown",
									"profile": "",
									"avatar_url": "",
									"birth_time": "2000-01-01",
									"authority": "admin"
								},
								"extra": {
									"subscribing_cnt": 1,
									"subscriber_cnt": 2,
									"video_cnt": 0,
									"playlist_cnt": 0
								}
							}
 						} */
func (u *userCtrl) QueryUser(c *gin.Context) {
	uidString := c.Param("uid")
	uid, err := strconv.Atoi(uidString)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			dto.Result{}.Error(http.StatusBadRequest).SetMessage(exception.RouteParamError.Error()))
		return
	}

	user := dao.UserDao.QueryByUid(uid)
	if user == nil {
		c.JSON(http.StatusNotFound,
			dto.Result{}.Error(http.StatusNotFound).SetMessage(exception.UserNotFoundError.Error()))
		return
	}

	isSelfOrAdmin := middleware.GetAuthUser(c) == nil || user.Authority == enum.AuthAdmin
	if !isSelfOrAdmin {
		user.PhoneNumber = ""
	}
	subscribingCnt, subscriberCnt, _ := dao.SubDao.QuerySubCnt(user.Uid)
	extraInfo := &dto.UserExtraInfo{
		PhoneNumber:      user.PhoneNumber,
		SubscribingCount: subscribingCnt,
		SubscriberCount:  subscriberCnt,
		VideoCount:       0,
		PlaylistCount:    0,
	}

	c.JSON(http.StatusOK,
		dto.Result{}.Ok().PutData("user", user).PutData("extra", extraInfo))
}

// @Router 				/user/ [PUT] [Auth]
// @Summary 			更新用户
// @Description 		更新用户信息
// @Param 				Authorization header string true "用户登录令牌"
// @Param 				username formData string false "用户名" minLength(8) maxLength(30)
// @Param 				sex formData string false "用户性别" enum(male, female, unknown)
// @Param 				profile formData string false "用户简介" minLength(0) maxLength(255)
// @Param 				birth_time formData string false "用户生日，固定格式为2000-01-01"
// @Param 				phone_number formData string false "用户手机号码"
// @Accept 				multipart/form-data
// @ErrorCode 			400 request format error
// @ErrorCode 			401 authorization failed
// @ErrorCode 			401 token has expired
// @ErrorCode 			404 user not found
// @ErrorCode 			500 username duplicated
// @ErrorCode 			500 user update failed
/* @Success 200 		{
							"code": 200,
							"message": "Success",
							"data": {
								"uid": 10,
								"username": "aoihosizora",
								"sex": "male",
								"profile": "Demo Profile",
								"avatar_url": "",
								"birth_time": "2019-12-18",
								"authority": "admin"
							}
 						} */
func (u *userCtrl) UpdateUser(c *gin.Context) {
	user := middleware.GetAuthUser(c)

	username := c.DefaultPostForm("username", user.Username)
	profile := c.DefaultPostForm("profile", user.Profile)
	if !model.FormatCheck.Username(username) || !model.FormatCheck.UserProfile(profile) {
		c.JSON(http.StatusBadRequest,
			dto.Result{}.Error(http.StatusBadRequest).SetMessage(exception.FormatError.Error()))
		return
	}
	user.Username = username
	user.Profile = profile
	user.Sex = enum.StringToSex(c.DefaultPostForm("sex", string(user.Sex)))
	user.BirthTime = vo.JsonDate{}.Parse(c.DefaultPostForm("birth_time", user.BirthTime.String()), user.BirthTime)
	user.PhoneNumber = c.DefaultPostForm("phone_number", user.PhoneNumber)

	status := dao.UserDao.Update(user)
	if status == database.DbNotFound {
		c.JSON(http.StatusNotFound,
			dto.Result{}.Error(http.StatusNotFound).SetMessage(exception.UserNotFoundError.Error()))
		return
	} else if status == database.DbExisted {
		c.JSON(http.StatusInternalServerError,
			dto.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.UserNameUsedError.Error()))
		return
	} else if status == database.DbFailed {
		c.JSON(http.StatusInternalServerError,
			dto.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.UserUpdateError.Error()))
		return
	}

	c.JSON(http.StatusOK,
		dto.Result{}.Ok().SetData(user))
}

// @Router 				/user/ [DELETE] [Auth]
// @Summary 			删除用户
// @Description 		删除用户所有信息
// @Param 				Authorization header string true "用户登录令牌"
// @Accept 				multipart/form-data
// @ErrorCode 			401 authorization failed
// @ErrorCode 			401 token has expired
// @ErrorCode 			404 user not found
// @ErrorCode 			404 user delete failed
/* @Success 200 		{
							"code": 200,
							"message": "Success"
 						} */
func (u *userCtrl) DeleteUser(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	status := dao.UserDao.Delete(user.Uid)
	if status == database.DbNotFound {
		c.JSON(http.StatusNotFound,
			dto.Result{}.Error(http.StatusNotFound).SetMessage(exception.UserNotFoundError.Error()))
		return
	} else if status == database.DbFailed {
		c.JSON(http.StatusInternalServerError,
			dto.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.UserDeleteError.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Result{}.Ok())
}
