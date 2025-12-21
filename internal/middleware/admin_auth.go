/*
Package middleware 管理后台认证中间件
Author: AliMPay Team
Description: 提供管理后台的认证和鉴权功能

功能:
  - Session管理
  - 登录验证
  - 访问控制
  - 操作日志
*/
package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"alimpay-go/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

/*
AdminAuthMiddleware 管理员认证中间件配置
字段:
  - merchantID: 商户ID
  - merchantKey: 商户密钥
  - sessions: session存储
  - mu: 读写锁
*/
type AdminAuthMiddleware struct {
	merchantID  string
	merchantKey string
	sessions    map[string]*Session
	mu          sync.RWMutex
}

/*
Session 会话信息
字段:
  - Token: 会话令牌
  - MerchantID: 商户ID
  - CreatedAt: 创建时间
  - LastAccess: 最后访问时间
  - IP: 客户端IP
*/
type Session struct {
	Token      string
	MerchantID string
	CreatedAt  time.Time
	LastAccess time.Time
	IP         string
}

/*
NewAdminAuthMiddleware 创建管理员认证中间件
参数:
  - merchantID: 商户ID
  - merchantKey: 商户密钥

返回:
  - *AdminAuthMiddleware: 认证中间件实例
*/
func NewAdminAuthMiddleware(merchantID, merchantKey string) *AdminAuthMiddleware {
	middleware := &AdminAuthMiddleware{
		merchantID:  merchantID,
		merchantKey: merchantKey,
		sessions:    make(map[string]*Session),
	}

	// 启动session清理任务
	go middleware.cleanupExpiredSessions()

	return middleware
}

/*
RequireAuth 要求认证的中间件
使用方法:

	router.GET("/admin/dashboard", authMiddleware.RequireAuth(), handler)
*/
func (m *AdminAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查session cookie
		token, err := c.Cookie("admin_session")
		if err != nil || token == "" {
			// 未登录，重定向到登录页
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		// 验证session
		session := m.getSession(token)
		if session == nil {
			// session无效
			c.SetCookie("admin_session", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		// 更新最后访问时间
		m.updateSessionAccess(token, c.ClientIP())

		// 设置上下文
		c.Set("admin_merchant_id", session.MerchantID)
		c.Set("admin_logged_in", true)

		c.Next()
	}
}

/*
HandleLogin 处理登录请求
POST /admin/login
参数:
  - pid: 商户ID
  - key: 商户密钥
*/
func (m *AdminAuthMiddleware) HandleLogin(c *gin.Context) {
	// 已登录用户跳转到后台
	if token, err := c.Cookie("admin_session"); err == nil && token != "" {
		if session := m.getSession(token); session != nil {
			c.Redirect(http.StatusFound, "/admin/dashboard")
			return
		}
	}

	// GET请求显示登录页面
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "admin_login.html", gin.H{
			"error": c.Query("error"),
		})
		return
	}

	// POST请求处理登录
	pid := c.PostForm("pid")
	key := c.PostForm("key")

	// 验证参数
	if pid == "" || key == "" {
		c.HTML(http.StatusOK, "admin_login.html", gin.H{
			"error": "请输入商户ID和密钥",
		})
		return
	}

	// 验证凭据
	if pid != m.merchantID || key != m.merchantKey {
		logger.Warn("Failed admin login attempt",
			zap.String("pid", pid),
			zap.String("ip", c.ClientIP()))

		c.HTML(http.StatusOK, "admin_login.html", gin.H{
			"error": "商户ID或密钥错误",
		})
		return
	}

	// 创建session
	token := m.createSession(pid, c.ClientIP())

	// 设置cookie（24小时有效）
	c.SetCookie("admin_session", token, 86400, "/", "", false, true)

	logger.Info("Admin logged in successfully",
		zap.String("pid", pid),
		zap.String("ip", c.ClientIP()))

	// 重定向到后台
	c.Redirect(http.StatusFound, "/admin/dashboard")
}

/*
HandleLogout 处理登出请求
GET /admin/logout
*/
func (m *AdminAuthMiddleware) HandleLogout(c *gin.Context) {
	// 获取并删除session
	token, err := c.Cookie("admin_session")
	if err == nil && token != "" {
		m.deleteSession(token)
	}

	// 清除cookie
	c.SetCookie("admin_session", "", -1, "/", "", false, true)

	logger.Info("Admin logged out",
		zap.String("ip", c.ClientIP()))

	// 重定向到登录页
	c.Redirect(http.StatusFound, "/admin/login")
}

/*
createSession 创建新session
参数:
  - merchantID: 商户ID
  - ip: 客户端IP

返回:
  - string: session令牌
*/
func (m *AdminAuthMiddleware) createSession(merchantID, ip string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 生成token
	token := m.generateToken(merchantID, ip)

	// 创建session
	session := &Session{
		Token:      token,
		MerchantID: merchantID,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
		IP:         ip,
	}

	m.sessions[token] = session

	return token
}

/*
getSession 获取session
参数:
  - token: session令牌

返回:
  - *Session: session信息
*/
func (m *AdminAuthMiddleware) getSession(token string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[token]
	if !exists {
		return nil
	}

	// 检查是否过期（24小时）
	if time.Since(session.LastAccess) > 24*time.Hour {
		return nil
	}

	return session
}

/*
updateSessionAccess 更新session最后访问时间
参数:
  - token: session令牌
  - ip: 客户端IP
*/
func (m *AdminAuthMiddleware) updateSessionAccess(token, ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[token]; exists {
		session.LastAccess = time.Now()
		// 检查IP变化
		if session.IP != ip {
			logger.Warn("Session IP changed",
				zap.String("token", token[:8]+"..."),
				zap.String("old_ip", session.IP),
				zap.String("new_ip", ip))
			session.IP = ip
		}
	}
}

/*
deleteSession 删除session
参数:
  - token: session令牌
*/
func (m *AdminAuthMiddleware) deleteSession(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, token)
}

/*
generateToken 生成session令牌
参数:
  - merchantID: 商户ID
  - ip: 客户端IP

返回:
  - string: 令牌
*/
func (m *AdminAuthMiddleware) generateToken(merchantID, ip string) string {
	data := merchantID + ip + time.Now().String()
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

/*
cleanupExpiredSessions 清理过期session
定时任务，每小时运行一次
*/
func (m *AdminAuthMiddleware) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		count := 0
		for token, session := range m.sessions {
			if time.Since(session.LastAccess) > 24*time.Hour {
				delete(m.sessions, token)
				count++
			}
		}
		m.mu.Unlock()

		if count > 0 {
			logger.Info("Cleaned up expired admin sessions", zap.Int("count", count))
		}
	}
}

/*
GetActiveSessions 获取活跃session数量
返回:
  - int: 活跃session数量
*/
func (m *AdminAuthMiddleware) GetActiveSessions() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, session := range m.sessions {
		if time.Since(session.LastAccess) <= 24*time.Hour {
			count++
		}
	}

	return count
}
