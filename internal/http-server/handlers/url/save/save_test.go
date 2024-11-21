package save

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foreground-eclipse/url-shortener/internal/http-server/handlers/url/save/mocks"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestSaveHandler(t *testing.T) {
	testCases := []struct {
		name               string
		inputURL           string
		inputAlias         string
		mockCheckAlias     func(m *mocks.URLSaver)
		mockSaveURL        func(m *mocks.URLSaver)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:       "Successful Save with Custom Alias",
			inputURL:   "https://example.com",
			inputAlias: "custom",
			mockCheckAlias: func(m *mocks.URLSaver) {
				m.On("CheckIfAliasExists", "custom").Return(false, nil)
			},
			mockSaveURL: func(m *mocks.URLSaver) {
				m.On("SaveURL", "https://example.com", "custom").Return(int64(1), nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Alias Already Exists",
			inputURL:   "https://example.com",
			inputAlias: "existing",
			mockCheckAlias: func(m *mocks.URLSaver) {
				m.On("CheckIfAliasExists", "existing").Return(true, nil)
			},
			mockSaveURL:        func(m *mocks.URLSaver) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"error":"alias already exists"}`,
		},
		{
			name:       "Invalid URL",
			inputURL:   "not-a-url",
			inputAlias: "",
			mockCheckAlias: func(m *mocks.URLSaver) {
				// No calls expected
			},
			mockSaveURL:        func(m *mocks.URLSaver) {},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create logger
			logger := slog.Default()

			// Create mock URLSaver
			mockURLSaver := mocks.NewURLSaver(t)

			// Setup mock expectations
			tc.mockCheckAlias(mockURLSaver)
			tc.mockSaveURL(mockURLSaver)

			// Create handler
			handler := New(logger, mockURLSaver)

			// Prepare request body
			reqBody, err := json.Marshal(Request{
				URL:   tc.inputURL,
				Alias: tc.inputAlias,
			})
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewBuffer(reqBody))
			req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "test-request-id"))

			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler(w, req)

			// Check response
			resp := w.Result()
			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			// If expected response is not empty, check response body
			if tc.expectedResponse != "" {
				var respBody map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&respBody)
				assert.NoError(t, err)
				assert.Contains(t, respBody["error"], tc.expectedResponse)
			}

			// Verify mock expectations
			mockURLSaver.AssertExpectations(t)
		})
	}
}

// Additional test for random alias generation
func TestSaveHandlerRandomAlias(t *testing.T) {
	logger := slog.Default()
	mockURLSaver := mocks.NewURLSaver(t)

	// Expect first generated alias to exist, then generate a new one
	mockURLSaver.On("CheckIfAliasExists", mock.AnythingOfType("string")).Return(true, nil).Once()
	mockURLSaver.On("CheckIfAliasExists", mock.AnythingOfType("string")).Return(false, nil)
	mockURLSaver.On("SaveURL", "https://example.com", mock.AnythingOfType("string")).Return(int64(1), nil)

	handler := New(logger, mockURLSaver)

	reqBody, err := json.Marshal(Request{
		URL: "https://example.com",
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "test-request-id"))

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mockURLSaver.AssertExpectations(t)
}
