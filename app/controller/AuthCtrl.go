package controller

import (
	"net/http"
	"strconv"
	"vid/app/controller/exception"
	"vid/app/database"
	"vid/app/database/dao"
	"vid/app/middleware"
	"vid/app/model"
	"vid/app/model/dto"
	"vid/app/model/po"
	"vid/app/util"

	"github.com/gin-gonic/gin"
)

type authCtrl struct{}

var AuthCtrl = new(authCtrl)

// @Router 				/auth/login [POST]
// @Summary 			登录
// @Description 		用户登录
// @Param 				username formData string true "用户名"
// @Param 				password formData string true "用户密码"
// @Param 				expire formData integer false "登录有效期，默认为七天"
// @Accept 				multipart/form-data
// @ErrorCode 			400 request form data error
// @ErrorCode 			401 password error
// @ErrorCode 			404 user not found
// @ErrorCode 			500 login failed
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
									"authority": "normal"
								},
								"token": "Bearer xxx"
							}
 						} */
func (u *authCtrl) Login(c *gin.Context) {

	username := c.PostForm("username")
	password := c.PostForm("password")
	expireString := c.PostForm("expire")
	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest,
			dto.Result{}.Error(http.StatusBadRequest).SetMessage(exception.FormParamError.Error()))
		return
	}
	expire := util.JwtExpire
	if val, err := strconv.Atoi(expireString); err == nil {
		expire = int64(val)
	}

	passRecord := dao.PassDao.QueryByUsername(username)
	if passRecord == nil {
		c.JSON(http.StatusNotFound,
			dto.Result{}.Error(http.StatusNotFound).SetMessage(exception.UserNotFoundError.Error()))
		return
	}

	if !util.PassUtil.MD5Check(password, passRecord.EncryptedPass) {
		c.JSON(http.StatusUnauthorized,
			dto.Result{}.Error(http.StatusUnauthorized).SetMessage(exception.PasswordError.Error()))
		return
	}

	token, err := util.PassUtil.GenToken(passRecord.User.Uid, expire)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			dto.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.LoginError.Error()))
		return
	}

	c.JSON(http.StatusOK,
		dto.Result{}.Ok().PutData("user", passRecord.User).PutData("token", token))
}

// @Router 				/auth/register [POST]
// @Summary 			注册
// @Description 		用户注册
// @Param 				username formData string true "用户名" minLength(5) maxLength(30)
// @Param 				password formData string true "用户密码" minLength(8) maxLength(30)
// @Accept 				multipart/form-data
// @ErrorCode			400 request form data error
// @ErrorCode			400 request format error
// @ErrorCode			500 username duplicated
// @ErrorCode			500 register failed
/* @Success 200 		{
							"code": 200,
							"message": "Success",
							"data": {
								"uid": 10,
								"username": "aoihosizora",
								"sex": "unknown",
								"profile": "",
								"avatar_url": "",
								"birth_time": "2000-01-01",
								"authority": "normal"
							}
 						} */
func (u *authCtrl) Register(c *gin.Context) {

	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest,
			dto.Result{}.Error(http.StatusBadRequest).SetMessage(exception.FormParamError.Error()))
		return
	}
	if !model.FormatCheck.Username(username) || !model.FormatCheck.Password(password) {
		c.JSON(http.StatusBadRequest,
			dto.Result{}.Error(http.StatusBadRequest).SetMessage(exception.FormatError.Error()))
		return
	}

	passRecord := &po.PassRecord{
		EncryptedPass: util.PassUtil.MD5Encode(password),
		User: &po.User{
			Username: username, RegisterIP: c.ClientIP(),
		},
	}
	status := dao.PassDao.Insert(passRecord)
	if status == database.DbExisted {
		c.JSON(http.StatusInternalServerError,
			dto.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.UserNameUsedError.Error()))
		return
	} else if status == database.DbFailed {
		c.JSON(http.StatusInternalServerError,
			dto.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.RegisterError.Error()))
		return
	}

	c.JSON(http.StatusOK,
		dto.Result{}.Ok().SetData(passRecord.User))
}

// @Router 				/auth/pass [POST] [Auth]
// @Summary 			修改密码
// @Description 		用户修改密码
// @Param 				Authorization header string true "用户登录令牌"
// @Param 				password formData string true "用户密码" minLength(8) maxLength(30)
// @Accept 				multipart/form-data
// @ErrorCode			400 request form data error
// @ErrorCode			400 request format error
// @ErrorCode			401 authorization failed
// @ErrorCode			401 token has expired
// @ErrorCode			404 user not found
// @ErrorCode			500 update password failed
/* @Success 200 		{
							"code": 200,
							"message": "Success",
							"data": {
								"uid": 10,
								"username": "aoihosizora",
								"sex": "unknown",
								"profile": "",
								"avatar_url": "",
								"birth_time": "2000-01-01",
								"authority": "normal"
							}
 						} */
func (u *authCtrl) ModifyPass(c *gin.Context) {
	user := middleware.GetAuthUser(c)

	password := c.PostForm("password")
	if password == "" {
		c.JSON(http.StatusBadRequest,
			dto.Result{}.Error(http.StatusBadRequest).SetMessage(exception.FormParamError.Error()))
		return
	}
	if !model.FormatCheck.Password(password) {
		c.JSON(http.StatusBadRequest,
			dto.Result{}.Error(http.StatusBadRequest).SetMessage(exception.FormatError.Error()))
		return
	}

	passRecord := &po.PassRecord{
		EncryptedPass: util.PassUtil.MD5Encode(password),
		User:          user,
		Uid:           user.Uid,
	}
	status := dao.PassDao.Update(passRecord)
	if status == database.DbNotFound {
		c.JSON(http.StatusNotFound,
			dto.Result{}.Error(http.StatusNotFound).SetMessage(exception.UserNotFoundError.Error()))
		return
	} else if status == database.DbFailed {
		c.JSON(http.StatusInternalServerError,
			dto.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.UpdatePassError.Error()))
		return
	}

	c.JSON(http.StatusOK,
		dto.Result{}.Ok().SetData(passRecord.User))
}

// @Router 				/auth/ [GET] [Auth]
// @Summary 			查看当前登录用户
// @Description 		根据认证令牌，查看当前登录用户
// @Param 				Authorization header string true "用户登录令牌"
// @Accept 				multipart/form-data
// @ErrorCode			401 authorization failed
// @ErrorCode			401 token has expired
/* @Success 200 		{
							"code": 200,
							"message": "Success",
							"data": {
								"uid": 10,
								"username": "aoihosizora",
								"sex": "unknown",
								"profile": "",
								"avatar_url": "",
								"birth_time": "2000-01-01",
								"authority": "normal"
							}
 						} */
func (u *authCtrl) CurrentUser(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	c.JSON(http.StatusOK,
		dto.Result{}.Ok().SetData(user))
}
