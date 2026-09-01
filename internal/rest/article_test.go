package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	faker "github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/bxcodec/go-clean-arch/internal/rest"
	"github.com/bxcodec/go-clean-arch/internal/rest/mocks"
)

func newRouter(svc rest.ArticleService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	rest.NewArticleHandler(r, svc)
	return r
}

func TestFetch(t *testing.T) {
	var mockArticle domain.Article
	err := faker.FakeData(&mockArticle)
	assert.NoError(t, err)
	mockUCase := new(mocks.ArticleService)
	mockListArticle := make([]domain.Article, 0)
	mockListArticle = append(mockListArticle, mockArticle)
	num := 1
	cursor := "2"
	mockUCase.On("Fetch", mock.Anything, cursor, int64(num)).Return(mockListArticle, "10", nil)

	r := newRouter(mockUCase)
	req, err := http.NewRequestWithContext(context.TODO(),
		http.MethodGet, "/articles?num=1&cursor="+cursor, strings.NewReader(""))
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	responseCursor := rec.Header().Get("X-Cursor")
	assert.Equal(t, "10", responseCursor)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUCase.AssertExpectations(t)
}

func TestFetchError(t *testing.T) {
	mockUCase := new(mocks.ArticleService)
	num := 1
	cursor := "2"
	mockUCase.On("Fetch", mock.Anything, cursor, int64(num)).Return(nil, "", domain.ErrInternalServerError)

	r := newRouter(mockUCase)
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/articles?num=1&cursor="+cursor, strings.NewReader(""))
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	responseCursor := rec.Header().Get("X-Cursor")
	assert.Equal(t, "", responseCursor)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUCase.AssertExpectations(t)
}

func TestGetByID(t *testing.T) {
	var mockArticle domain.Article
	err := faker.FakeData(&mockArticle)
	assert.NoError(t, err)

	mockUCase := new(mocks.ArticleService)

	num := int(mockArticle.ID)

	mockUCase.On("GetByID", mock.Anything, int64(num)).Return(mockArticle, nil)

	r := newRouter(mockUCase)
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/articles/"+strconv.Itoa(num), strings.NewReader(""))
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockUCase.AssertExpectations(t)
}

func TestStore(t *testing.T) {
	mockArticle := domain.Article{
		Title:     "Title",
		Content:   "Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tempMockArticle := mockArticle
	tempMockArticle.ID = 0
	mockUCase := new(mocks.ArticleService)

	j, err := json.Marshal(tempMockArticle)
	assert.NoError(t, err)

	mockUCase.On("Store", mock.Anything, mock.AnythingOfType("*domain.Article")).Return(nil)

	r := newRouter(mockUCase)
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/articles", strings.NewReader(string(j)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockUCase.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	var mockArticle domain.Article
	err := faker.FakeData(&mockArticle)
	assert.NoError(t, err)

	mockUCase := new(mocks.ArticleService)

	num := int(mockArticle.ID)

	mockUCase.On("Delete", mock.Anything, int64(num)).Return(nil)

	r := newRouter(mockUCase)
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodDelete, "/articles/"+strconv.Itoa(num), strings.NewReader(""))
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockUCase.AssertExpectations(t)
}
