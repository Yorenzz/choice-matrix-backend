package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"choice-matrix-backend/internal/auth"
	"choice-matrix-backend/internal/models"
	"choice-matrix-backend/internal/repository"
	"choice-matrix-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	userRepo        *repository.UserRepository
	refreshStore    *auth.RefreshStore
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	refreshStore *auth.RefreshStore,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *AuthHandler {
	return &AuthHandler{
		userRepo:        userRepo,
		refreshStore:    refreshStore,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type requestLocation struct {
	IP          string
	CountryCode string
	RegionName  string
	CityName    string
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func defaultNickname(email, nickname string) string {
	trimmedNickname := strings.TrimSpace(nickname)
	if trimmedNickname != "" {
		return trimmedNickname
	}

	localPart := strings.Split(strings.TrimSpace(email), "@")[0]
	if localPart == "" {
		return "New User"
	}
	return localPart
}

func extractRequestLocation(c *gin.Context) requestLocation {
	return requestLocation{
		IP:          c.ClientIP(),
		CountryCode: strings.ToUpper(strings.TrimSpace(c.GetHeader("X-Country-Code"))),
		RegionName:  strings.TrimSpace(c.GetHeader("X-Region-Name")),
		CityName:    strings.TrimSpace(c.GetHeader("X-City-Name")),
	}
}

func buildUserResponse(user *models.User) gin.H {
	var lastLoginAt any
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	return gin.H{
		"id":            user.ID,
		"email":         user.Email,
		"nickname":      user.Nickname,
		"avatar_url":    user.AvatarURL,
		"status":        user.Status,
		"plan":          user.Plan,
		"pro":           user.Plan == models.UserPlanPro,
		"credits":       user.AICredits,
		"last_login_at": lastLoginAt,
		"last_login_ip": user.LastLoginIP,
		"country_code":  user.CountryCode,
		"region_name":   user.RegionName,
		"city_name":     user.CityName,
	}
}

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", refreshToken, int(h.refreshTokenTTL.Seconds()), "/", "", false, true)
}

func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}

func (h *AuthHandler) issueTokenPair(ctx context.Context, userID uint) (gin.H, error) {
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		return nil, err
	}

	accessToken, err := utils.GenerateAccessToken(userID, h.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(userID, sessionID, h.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	if err := h.refreshStore.Save(ctx, sessionID, userID, h.refreshTokenTTL); err != nil {
		return nil, err
	}

	return gin.H{
		"token":         accessToken,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的请求参数: " + err.Error()})
		return
	}

	req.Email = normalizeEmail(req.Email)
	location := extractRequestLocation(c)

	if existingUser, err := h.userRepo.FindByEmail(req.Email); err == nil && existingUser != nil {
		c.JSON(http.StatusOK, gin.H{"code": 409, "message": "该邮箱已被注册"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "密码加密失败"})
		return
	}

	now := time.Now()
	resetAt := now.AddDate(0, 1, 0)
	user := &models.User{
		Email:            req.Email,
		PasswordHash:     string(hashedPassword),
		Nickname:         defaultNickname(req.Email, req.Nickname),
		Status:           models.UserStatusActive,
		Plan:             models.UserPlanFree,
		AICredits:        10,
		AICreditsResetAt: &resetAt,
		LastLoginAt:      &now,
		LastLoginIP:      location.IP,
		CountryCode:      location.CountryCode,
		RegionName:       location.RegionName,
		CityName:         location.CityName,
	}

	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建用户失败"})
		return
	}

	tokenData, err := h.issueTokenPair(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "生成令牌失败"})
		return
	}

	h.setRefreshTokenCookie(c, tokenData["refresh_token"].(string))
	tokenData["user"] = buildUserResponse(user)
	delete(tokenData, "refresh_token")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "注册成功",
		"data":    tokenData,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的请求参数: " + err.Error()})
		return
	}

	req.Email = normalizeEmail(req.Email)

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{"code": 401, "message": "邮箱或密码错误"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询用户失败"})
		return
	}

	if user.Status == models.UserStatusDisabled {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "账号已被禁用"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "邮箱或密码错误"})
		return
	}

	location := extractRequestLocation(c)
	loginAt := time.Now()
	if err := h.userRepo.UpdateLoginProfile(user.ID, loginAt, location.IP, location.CountryCode, location.RegionName, location.CityName); err == nil {
		user.LastLoginAt = &loginAt
		user.LastLoginIP = location.IP
		user.CountryCode = location.CountryCode
		user.RegionName = location.RegionName
		user.CityName = location.CityName
	}

	tokenData, err := h.issueTokenPair(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "生成令牌失败"})
		return
	}

	h.setRefreshTokenCookie(c, tokenData["refresh_token"].(string))
	tokenData["user"] = buildUserResponse(user)
	delete(tokenData, "refresh_token")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data":    tokenData,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || strings.TrimSpace(refreshToken) == "" {
		h.clearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "刷新令牌不存在"})
		return
	}

	claims, err := utils.ParseRefreshToken(refreshToken)
	if err != nil {
		h.clearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "刷新令牌无效或已过期"})
		return
	}

	userID, err := h.refreshStore.GetUserID(c.Request.Context(), claims.SessionID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			h.clearRefreshTokenCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "刷新会话已失效"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "刷新会话校验失败"})
		return
	}

	if userID != claims.UserID {
		_ = h.refreshStore.Delete(c.Request.Context(), claims.SessionID)
		h.clearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "刷新会话不匹配"})
		return
	}

	user, err := h.userRepo.FindByID(claims.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			_ = h.refreshStore.Delete(c.Request.Context(), claims.SessionID)
			h.clearRefreshTokenCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败"})
		return
	}

	if user.Status == models.UserStatusDisabled {
		_ = h.refreshStore.Delete(c.Request.Context(), claims.SessionID)
		h.clearRefreshTokenCookie(c)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "账号已被禁用"})
		return
	}

	tokenData, err := h.issueTokenPair(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "刷新令牌失败"})
		return
	}

	_ = h.refreshStore.Delete(c.Request.Context(), claims.SessionID)
	h.setRefreshTokenCookie(c, tokenData["refresh_token"].(string))
	tokenData["user"] = buildUserResponse(user)
	delete(tokenData, "refresh_token")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "刷新成功",
		"data":    tokenData,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if refreshToken, err := c.Cookie("refresh_token"); err == nil && strings.TrimSpace(refreshToken) != "" {
		claims, parseErr := utils.ParseRefreshToken(refreshToken)
		if parseErr == nil && claims.SessionID != "" {
			_ = h.refreshStore.Delete(c.Request.Context(), claims.SessionID)
		}
	}

	h.clearRefreshTokenCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "退出成功",
	})
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权的访问"})
		return
	}

	user, err := h.userRepo.FindByID(userID.(uint))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{"code": 404, "message": "用户不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "获取用户信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    buildUserResponse(user),
	})
}
