package middleware

import (
	"net/http"
	"sonnda-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

// RequireRole retorna um middleware que verifica se o usuário tem a role necessária
func RequireRole(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "usuário não autenticado",
			})
			c.Abort()
			return
		}

		userRole := user.Role

		// Verifica se a role do usuário está na lista permitida
		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "você não tem permissão para acessar este recurso",
		})
		c.Abort()
	}
}

// RequireProfessional middleware específico para rotas de médicos
func RequireProfessional() gin.HandlerFunc {
	return RequireRole(domain.RoleDoctor, domain.RoleAdmin)
}

// RequirePatient middleware específico para rotas de pacientes
func RequirePatient() gin.HandlerFunc {
	return RequireRole(domain.RolePatient, domain.RoleAdmin)
}

// RequireAdmin middleware específico para rotas de administradores
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(domain.RoleAdmin)
}
