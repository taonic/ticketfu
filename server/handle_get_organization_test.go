package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/taonic/ticketfu/config"
	"github.com/taonic/ticketfu/worker/org"
	"go.temporal.io/sdk/mocks"
	"go.temporal.io/server/common/log"
)

func TestHandleGetOrganization(t *testing.T) {
	validSummary := `{
		"overview": "Test Organization overview",
		"main_topics": ["Topic 1", "Topic 2"],
		"key_people": ["Person 1", "Person 2"],
		"key_insights": "Key insights about the organization",
		"trending_topics": [
			{
				"topic": "Topic A",
				"frequency": 5,
				"importance": "high"
			}
		],
		"recommended_actions": ["Action 1", "Action 2"]
	}`

	testCases := []struct {
		name           string
		orgID          string
		setupMock      func(*mocks.Client)
		expectedStatus int
		expectedResp   map[string]interface{}
		expectedError  string
	}{
		{
			name:  "Success",
			orgID: "123",
			setupMock: func(m *mocks.Client) {
				mockFuture := &mocks.Value{}
				mockFuture.On("Get", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					resp := args.Get(0).(*org.QueryOrganizationOutput)
					resp.Summary = validSummary
				}).Return(nil)

				m.On("QueryWorkflow", mock.Anything, "organization-workflow-123", "", org.QueryOrganizationSummary, "").
					Return(mockFuture, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: map[string]interface{}{
				"summary": map[string]interface{}{
					"overview":     "Test Organization overview",
					"main_topics":  []interface{}{"Topic 1", "Topic 2"},
					"key_people":   []interface{}{"Person 1", "Person 2"},
					"key_insights": "Key insights about the organization",
					"trending_topics": []interface{}{
						map[string]interface{}{
							"topic":      "Topic A",
							"frequency":  float64(5),
							"importance": "high",
						},
					},
					"recommended_actions": []interface{}{"Action 1", "Action 2"},
				},
			},
		},
		{
			name:  "Workflow Query Error",
			orgID: "456",
			setupMock: func(m *mocks.Client) {
				m.On("QueryWorkflow", mock.Anything, "organization-workflow-456", "", org.QueryOrganizationSummary, "").
					Return(nil, errors.New("workflow not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Failed to query workflow",
		},
		{
			name:  "Invalid JSON in Summary",
			orgID: "789",
			setupMock: func(m *mocks.Client) {
				mockFuture := &mocks.Value{}
				mockFuture.On("Get", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					resp := args.Get(0).(*org.QueryOrganizationOutput)
					resp.Summary = "invalid json"
				}).Return(nil)

				m.On("QueryWorkflow", mock.Anything, "organization-workflow-789", "", org.QueryOrganizationSummary, "").
					Return(mockFuture, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to parse summary JSON",
		},
		{
			name:  "Response Decode Error",
			orgID: "999",
			setupMock: func(m *mocks.Client) {
				mockFuture := &mocks.Value{}
				mockFuture.On("Get", mock.Anything, mock.Anything).Return(errors.New("decode error"))

				m.On("QueryWorkflow", mock.Anything, "organization-workflow-999", "", org.QueryOrganizationSummary, "").
					Return(mockFuture, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to decode workflow response",
		},
		{
			name:           "Missing Organization ID",
			orgID:          "",
			setupMock:      func(m *mocks.Client) {},
			expectedStatus: http.StatusMovedPermanently,
			expectedError:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mocks.Client{}
			tc.setupMock(mockClient)

			// Create server
			server := NewHTTPServer(config.ServerConfig{
				APIToken: "test-api-key",
			}, mockClient, log.NewTestLogger())

			// Create request
			req := httptest.NewRequest("GET", "/api/v1/organization/"+tc.orgID+"/summary", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(APIKeyHeader, "test-api-key")

			// Create the response recorder and router
			w := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/api/v1/organization/{orgId}/summary", server.handleGetOrganization)

			// Serve the request
			router.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				assert.Contains(t, w.Body.String(), tc.expectedError)
			} else if tc.expectedResp != nil {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// Check structure of response
				summaryObj, ok := resp["summary"].(map[string]interface{})
				assert.True(t, ok)

				expectedSummary := tc.expectedResp["summary"].(map[string]interface{})
				assert.Equal(t, expectedSummary["overview"], summaryObj["overview"])
				assert.Equal(t, expectedSummary["key_insights"], summaryObj["key_insights"])

				// Check arrays
				assert.ElementsMatch(t, expectedSummary["main_topics"], summaryObj["main_topics"])
				assert.ElementsMatch(t, expectedSummary["key_people"], summaryObj["key_people"])
				assert.ElementsMatch(t, expectedSummary["recommended_actions"], summaryObj["recommended_actions"])

				// Check trending topics
				expectedTopics := expectedSummary["trending_topics"].([]interface{})
				actualTopics := summaryObj["trending_topics"].([]interface{})
				assert.Equal(t, len(expectedTopics), len(actualTopics))

				expectedTopic := expectedTopics[0].(map[string]interface{})
				actualTopic := actualTopics[0].(map[string]interface{})
				assert.Equal(t, expectedTopic["topic"], actualTopic["topic"])
				assert.Equal(t, expectedTopic["frequency"], actualTopic["frequency"])
				assert.Equal(t, expectedTopic["importance"], actualTopic["importance"])
			}

			mockClient.AssertExpectations(t)
		})
	}
}
