package transports_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go"
	"github.com/stretchr/testify/assert"

	"github.com/cage1016/gokitistiok8s/pkg/addsvc/endpoints"
	"github.com/cage1016/gokitistiok8s/pkg/addsvc/service"
	"github.com/cage1016/gokitistiok8s/pkg/addsvc/transports"
	"github.com/cage1016/gokitistiok8s/test"
)

type httpTransportsTestSuite struct {
	ts *httptest.Server
}

func (s *httpTransportsTestSuite) SetupSubTest(t *testing.T) func(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stderr)
	zkt, _ := zipkin.NewTracer(nil, zipkin.WithNoopTracer(true))
	tracer := opentracing.GlobalTracer()

	svc := service.New(logger)
	eps := endpoints.New(svc, logger, tracer, zkt)
	handler := transports.NewHTTPHandler(eps, tracer, zkt, logger)
	s.ts = httptest.NewServer(handler)

	return func(t *testing.T) {
		defer s.ts.Close()
	}
}

func setupHTTPTransportTestCase(t *testing.T) (httpTransportsTestSuite, func(t *testing.T)) {
	return httpTransportsTestSuite{}, func(t *testing.T) {

	}
}

func TestHTTPServer_Sum(t *testing.T) {
	s, teardownTestCase := setupHTTPTransportTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		desc                                 string
		method, url, contentType, body, want string
		status                               int
	}{
		{
			desc:   "sum 3 + 4",
			method: http.MethodPost,
			url:    "/sum",
			body:   `{"a":3,"b":4}`,
			status: http.StatusOK,
			want:   `{"apiVersion":"","data":{"rs":7}}`,
		},
		{
			desc:   `sum 3 + 40`,
			method: http.MethodPost,
			url:    "/sum",
			status: http.StatusOK,
			body:   `{"a":3,"b":40}`,
			want:   `{"apiVersion":"","data":{"rs":43}}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			teardownSubTest := s.SetupSubTest(t)
			defer teardownSubTest(t)

			req := test.TestRequest{
				Client:      s.ts.Client(),
				Method:      tc.method,
				Url:         fmt.Sprintf("%s%s", s.ts.URL, tc.url),
				ContentType: tc.contentType,
				Body:        strings.NewReader(tc.body),
			}
			res, err := req.Make()
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			body, _ := ioutil.ReadAll(res.Body)

			assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
			assert.JSONEq(t, tc.want, strings.TrimSpace(string(body)), fmt.Sprintf("%s: expected body %v got %v", tc.desc, tc.want, strings.TrimSpace(string(body))))
		})
	}
}

func TestHTTPServer_Concat(t *testing.T) {
	s, teardownTestCase := setupHTTPTransportTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		desc                                 string
		method, url, contentType, body, want string
		status                               int
	}{
		{
			desc:   "concat 3 4",
			method: http.MethodPost,
			url:    "/concat",
			body:   `{"a":"3","b":"4"}`,
			status: http.StatusOK,
			want:   `{"apiVersion":"","data":{"rs":"34"}}`,
		},
		{
			desc:   `concat abc dd`,
			method: http.MethodPost,
			url:    "/concat",
			status: http.StatusOK,
			body:   `{"a":"abc","b":"dd"}`,
			want:   `{"apiVersion":"","data":{"rs":"abcdd"}}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			teardownSubTest := s.SetupSubTest(t)
			defer teardownSubTest(t)

			req := test.TestRequest{
				Client:      s.ts.Client(),
				Method:      tc.method,
				Url:         fmt.Sprintf("%s%s", s.ts.URL, tc.url),
				ContentType: tc.contentType,
				Body:        strings.NewReader(tc.body),
			}
			res, err := req.Make()
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			body, _ := ioutil.ReadAll(res.Body)
			assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
			assert.JSONEq(t, tc.want, strings.TrimSpace(string(body)), fmt.Sprintf("%s: expected body %v got %v", tc.desc, tc.want, strings.TrimSpace(string(body))))
		})
	}
}
