package common

import "github.com/gin-gonic/gin"

// RespondAppError é o ponto único para traduzir erros da aplicação/domínio para HTTP.
//
// Por enquanto, este helper apenas delega para o mapeamento por sentinelas em
// RespondDomainError.
func RespondAppError(c *gin.Context, err error) {
	RespondDomainError(c, err)
}
