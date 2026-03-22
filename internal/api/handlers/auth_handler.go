package handlers

import (
	"net/http"

	"choice-matrix-backend/internal/models"
	"choice-matrix-backend/internal/repository"
	"choice-matrix-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
}

func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的请求参数: " + err.Error()})
		return
	}

	// Check if user exists
	if existingUser, _ := h.userRepo.FindByEmail(req.Email); existingUser != nil {
		c.JSON(http.StatusOK, gin.H{"code": 409, "message": "该邮箱已被注册"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "密码加密失败"})
		return
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Nickname:     req.Nickname,
	}

	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "注册成功", "data": nil})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的请求参数: " + err.Error()})
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "邮箱或密码错误"})
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "邮箱或密码错误"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "生成令牌失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "登录成功",
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"email":    user.Email,
				"nickname": user.Nickname,
				"pro":      user.ProStatus,
			},
		},
	})
}

// GetCurrentUser returns the profile of the currently logged-in user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "未授权的访问"})
		return
	}

	user, err := h.userRepo.FindByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取成功",
		"data": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"nickname": user.Nickname,
			"pro":      user.ProStatus,
			"credits":  user.AICredits,
		},
	})
}
