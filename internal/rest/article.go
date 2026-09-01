package rest

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/bxcodec/go-clean-arch/domain"
)

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

// ArticleService represent the article's usecases
//
//go:generate mockery --name ArticleService
type ArticleService interface {
	Fetch(ctx context.Context, cursor string, num int64) ([]domain.Article, string, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)
	Update(ctx context.Context, ar *domain.Article) error
	GetByTitle(ctx context.Context, title string) (domain.Article, error)
	Store(context.Context, *domain.Article) error
	Delete(ctx context.Context, id int64) error
}

// ArticleHandler  represent the httphandler for article
type ArticleHandler struct {
	Service ArticleService
}

const defaultNum = 10

// NewArticleHandler will initialize the articles/ resources endpoint
func NewArticleHandler(r *gin.Engine, svc ArticleService) {
	handler := &ArticleHandler{
		Service: svc,
	}
	r.GET("/articles", handler.FetchArticle)
	r.POST("/articles", handler.Store)
	r.GET("/articles/:id", handler.GetByID)
	r.DELETE("/articles/:id", handler.Delete)
	r.OPTIONS("/articles", preflight("OPTIONS, GET, POST"))
	r.OPTIONS("/articles/:id", preflight("OPTIONS, GET, DELETE"))
}

// preflight replicates the automatic OPTIONS handling of the previous router.
func preflight(allow string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Allow", allow)
		c.Status(http.StatusNoContent)
	}
}

// FetchArticle will fetch the article based on given params
func (a *ArticleHandler) FetchArticle(c *gin.Context) {

	numS := c.Query("num")
	num, err := strconv.Atoi(numS)
	if err != nil || num == 0 {
		num = defaultNum
	}

	cursor := c.Query("cursor")
	ctx := c.Request.Context()

	listAr, nextCursor, err := a.Service.Fetch(ctx, cursor, int64(num))
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}

	// Use the writer header directly: gin's c.Header deletes the header when
	// the value is empty, but the baseline emits an empty X-Cursor header.
	c.Writer.Header().Set(`X-Cursor`, nextCursor)
	c.JSON(http.StatusOK, listAr)
}

// GetByID will get article by given id
func (a *ArticleHandler) GetByID(c *gin.Context) {
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrNotFound.Error())
		return
	}

	id := int64(idP)
	ctx := c.Request.Context()

	art, err := a.Service.GetByID(ctx, id)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, art)
}

func isRequestValid(m *domain.Article) (bool, error) {
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Store will store the article by given request body
func (a *ArticleHandler) Store(c *gin.Context) {
	var article domain.Article
	err := c.ShouldBind(&article)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	var ok bool
	if ok, err = isRequestValid(&article); !ok {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx := c.Request.Context()
	err = a.Service.Store(ctx, &article)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, article)
}

// Delete will delete article by given param
func (a *ArticleHandler) Delete(c *gin.Context) {
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrNotFound.Error())
		return
	}

	id := int64(idP)
	ctx := c.Request.Context()

	err = a.Service.Delete(ctx, id)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	logrus.Error(err)
	switch err {
	case domain.ErrInternalServerError:
		return http.StatusInternalServerError
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
