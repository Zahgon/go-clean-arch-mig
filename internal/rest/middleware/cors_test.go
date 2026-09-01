package middleware_test

import (
	"net/http"
	test "net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/bxcodec/go-clean-arch/internal/rest/middleware"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORS())
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := test.NewRequest(http.MethodGet, "/", nil)
	res := test.NewRecorder()
	r.ServeHTTP(res, req)

	assert.Equal(t, "*", res.Header().Get("Access-Control-Allow-Origin"))
}
