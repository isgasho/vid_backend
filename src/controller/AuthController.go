package controller

import (
	"github.com/Aoi-hosizora/ahlib/xcondition"
	"github.com/Aoi-hosizora/ahlib/xdi"
	"github.com/Aoi-hosizora/ahlib/xmapper"
	"github.com/gin-gonic/gin"
	"github.com/vidorg/vid_backend/src/config"
	"github.com/vidorg/vid_backend/src/controller/exception"
	"github.com/vidorg/vid_backend/src/database"
	"github.com/vidorg/vid_backend/src/database/dao"
	"github.com/vidorg/vid_backend/src/middleware"
	"github.com/vidorg/vid_backend/src/model/common"
	"github.com/vidorg/vid_backend/src/model/dto"
	"github.com/vidorg/vid_backend/src/model/dto/param"
	"github.com/vidorg/vid_backend/src/model/po"
	"github.com/vidorg/vid_backend/src/util"
	"net/http"
)

type AuthController struct {
	Config     *config.ServerConfig   `di:"~"`
	JwtService *middleware.JwtService `di:"~"`
	PassDao    *dao.PassDao           `di:"~"`
	TokenDao   *dao.TokenDao          `di:"~"`
	Mapper     *xmapper.EntityMapper  `di:"~"`
}

func NewAuthController(dic *xdi.DiContainer) *AuthController {
	ctrl := &AuthController{}
	dic.Inject(ctrl)
	if xdi.HasNilDi(ctrl) {
		panic("Has nil di field")
	}

	return ctrl
}

// @Router				/v1/auth/login [POST]
// @Summary				登录
// @Description			用户登录
// @Tag					Authorization
// @Param				username formData string true "用户名"
// @Param				password formData string true "用户密码"
// @Param				expire formData integer false "登录有效期，默认为七天"
// @Accept				multipart/form-data
// @ErrorCode			400 request param error
// @ErrorCode			401 password error
// @ErrorCode			404 user not found
// @ErrorCode			500 login failed
/* @Success 200			{
							"code": 200,
							"message": "success",
							"data": {
								"user": ${user},
								"token": "Bearer xxx",
								"expire": 604800
							}
 						} */
func (a *AuthController) Login(c *gin.Context) {
	loginParam := &param.LoginParam{}
	if err := c.ShouldBind(loginParam); err != nil {
		common.Result{}.Error(http.StatusBadRequest).SetMessage(exception.RequestParamError.Error()).JSON(c) // Login only use param error
		return
	}
	if loginParam.Expire <= 0 {
		loginParam.Expire = a.Config.JwtConfig.Expire
	}

	passRecord := a.PassDao.QueryByUsername(loginParam.Username)
	if passRecord == nil {
		common.Result{}.Error(http.StatusNotFound).SetMessage(exception.UserNotFoundError.Error()).JSON(c)
		return
	}

	if !util.AuthUtil.CheckPassword(loginParam.Password, passRecord.EncryptedPass) {
		common.Result{}.Error(http.StatusUnauthorized).SetMessage(exception.PasswordError.Error()).JSON(c)
		return
	}

	token, err := util.AuthUtil.GenerateToken(passRecord.User.Uid, loginParam.Expire, a.Config.JwtConfig)
	if err != nil {
		common.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.LoginError.Error()).JSON(c)
		return
	}

	ok := a.TokenDao.Insert(token, passRecord.Uid, loginParam.Expire)
	if !ok {
		common.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.LoginError.Error()).JSON(c)
		return
	}

	retDto := xcondition.First(a.Mapper.Map(&dto.UserDto{}, passRecord.User)).(*dto.UserDto)
	common.Result{}.Ok().
		PutData("user", retDto).
		PutData("token", token).
		PutData("expire", loginParam.Expire).JSON(c)
}

// @Router				/v1/auth/register [POST]
// @Summary				注册
// @Description			用户注册
// @Tag					Authorization
// @Param				username formData string true "用户名，长度在 [5, 30] 之间"
// @Param				password formData string true "用户密码，长度在 [8, 30] 之间"
// @Accept				multipart/form-data
// @ErrorCode			400 request param error
// @ErrorCode			400 request format error
// @ErrorCode			500 username has been used
// @ErrorCode			500 register failed
/* @Success 200			{
							"code": 200,
							"message": "success",
							"data": ${user}
 						} */
func (a *AuthController) Register(c *gin.Context) {
	registerParam := &param.RegisterParam{}
	if err := c.ShouldBind(registerParam); err != nil {
		common.Result{}.Error(http.StatusBadRequest).SetMessage(exception.WrapValidationError(err).Error()).JSON(c) // Register use wrap error
		return
	}

	encrypted, err := util.AuthUtil.EncryptPassword(registerParam.Password)
	if err != nil {
		common.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.RegisterError.Error()).JSON(c)
		return
	}
	passRecord := &po.PassRecord{
		EncryptedPass: encrypted,
		User: &po.User{
			Username:   registerParam.Username,
			RegisterIP: c.ClientIP(),
		},
	}
	status := a.PassDao.Insert(passRecord)
	if status == database.DbExisted {
		common.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.UsernameUsedError.Error()).JSON(c)
		return
	} else if status == database.DbFailed {
		common.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.RegisterError.Error()).JSON(c)
		return
	}

	retDto := xcondition.First(a.Mapper.Map(&dto.UserDto{}, passRecord.User)).(*dto.UserDto)
	common.Result{}.Ok().SetData(retDto).JSON(c)
}

// @Router				/v1/auth/ [GET] [Auth]
// @Summary				当前登录用户
// @Description			根据认证信息，查看当前登录用户
// @Tag					Authorization
// @Accept				multipart/form-data
/* @Success 200			{
							"code": 200,
							"message": "success",
							"data": ${user}
 						} */
func (a *AuthController) CurrentUser(c *gin.Context) {
	authUser := a.JwtService.GetAuthUser(c)
	retDto := xcondition.First(a.Mapper.Map(&dto.UserDto{}, authUser)).(*dto.UserDto)
	common.Result{}.Ok().SetData(retDto).JSON(c)
}

// @Router				/v1/auth/logout [POST] [Auth]
// @Summary				注销
// @Description			用户注销，删除认证信息
// @Tag					Authorization
// @Accept				multipart/form-data
// @ErrorCode			500 logout failed
/* @Success 200			{
							"code": 200,
							"message": "success"
 						} */
func (a *AuthController) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	ok := a.TokenDao.Delete(authHeader)
	if !ok {
		common.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.LogoutError.Error()).JSON(c)
		return
	}

	common.Result{}.Ok().JSON(c)
}

// @Router				/v1/auth/password [PUT] [Auth]
// @Summary				修改密码
// @Description			用户修改密码
// @Tag					Authorization
// @Param				password formData string true "用户密码，长度在 [8, 30] 之间"
// @Accept				multipart/form-data
// @ErrorCode			400 request param error
// @ErrorCode			400 request format error
// @ErrorCode			404 user not found
// @ErrorCode			500 update password failed
/* @Success 200			{
							"code": 200,
							"message": "success"
 						} */
func (a *AuthController) UpdatePassword(c *gin.Context) {
	authUser := a.JwtService.GetAuthUser(c)
	authHeader := c.GetHeader("Authorization")
	passParam := &param.PassParam{}
	if err := c.ShouldBind(passParam); err != nil {
		common.Result{}.Error(http.StatusBadRequest).SetMessage(exception.WrapValidationError(err).Error()).JSON(c)
		return
	}

	encrypted, err := util.AuthUtil.EncryptPassword(passParam.Password)
	if err != nil {
		common.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.UpdatePassError.Error()).JSON(c)
		return
	}
	passRecord := &po.PassRecord{
		EncryptedPass: encrypted,
		Uid:           authUser.Uid,
	}
	status := a.PassDao.Update(passRecord)
	if status == database.DbNotFound {
		common.Result{}.Error(http.StatusNotFound).SetMessage(exception.UserNotFoundError.Error()).JSON(c)
		return
	} else if status == database.DbFailed {
		common.Result{}.Error(http.StatusInternalServerError).SetMessage(exception.UpdatePassError.Error()).JSON(c)
		return
	}
	// Delete token but ignore result
	_ = a.TokenDao.Delete(authHeader)

	common.Result{}.Ok().JSON(c)
}
