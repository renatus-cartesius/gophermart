package handlers

// func TestUpdateHandler(t *testing.T) {
// 	type want struct {
// 		code        int
// 		contentType string
// 	}
// 	tests := []struct {
// 		name    string
// 		path    string
// 		pattern string
// 		metric  string
// 		value   string
// 		want    want
// 	}{
// 		{
// 			name:    "gauge test",
// 			path:    "http://localhost:8080/update/gauge/MCacheInuse/14400.000000",
// 			pattern: "POST /update/{type}/{name}/{value}",
// 			metric:  "Alloc",
// 			value:   "1233.1233",
// 			want: want{
// 				code:        200,
// 				contentType: "text/plain",
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			router := http.NewServeMux()
// 			router.HandleFunc(tt.pattern, UpdateHandler)

// 			server := httptest.NewServer(router)
// 			defer server.Close()

// 			request, _ := http.NewRequest(http.MethodPost, tt.path, nil)
// 			request.Header.Set("Content-Type", "text/plain")

// 			response, err := http.DefaultClient.Do(request)
// 			if err != nil {
// 				t.Fatalf("Failed to send request \"%s\": %v", tt.path, err)
// 			}
// 			defer response.Body.Close()

// 			assert.Equal(t, tt.want.code, response.StatusCode)

// 			assert.Equal(t, tt.want.contentType, response.Header.Get("Content-Type"))
// 		})
// 	}
// }
