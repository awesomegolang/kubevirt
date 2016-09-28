package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.org/x/net/context"
	"net/http"

	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"strings"
)

func newValidRequest() *http.Request {
	request, _ := http.NewRequest("POST", "/api/v1/domain/raw", nil)
	request.Body = ioutil.NopCloser(strings.NewReader("<domain><name>testvm</name></domain>"))
	request.Header.Set("Content-Type", "application/xml")
	return request

}

type mockVmService struct {
	VmName string
	rawXml []byte
}

func (m *mockVmService) StartVmRaw(vmName string, rawXml []byte) error {
	m.VmName = vmName
	m.rawXml = rawXml
	return nil
}

func (m *mockVmService) Clear() {
	m.VmName = ""
	m.rawXml = nil
}

var _ = Describe("VirtController", func() {
	var recorder *httptest.ResponseRecorder
	var request *http.Request
	ctx := context.Background()
	svc := mockVmService{}
	endpoints := Handlers{
		RawDomainHandler: makeRawDomainHandler(ctx, makeRawDomainEndpoint(&svc)),
	}
	handler := http.Handler(defineRoutes(&endpoints))

	BeforeEach(func() {
		request = newValidRequest()
		recorder = httptest.NewRecorder()
		svc.Clear()
	})

	Describe("REST call", func() {
		Context("with invalid URL", func() {
			It("should return 404", func() {
				request.URL, _ = url.Parse("/api/v1/wrong/url")
				handler.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})
		Context("with missing Content-Type header", func() {
			It("should return 404", func() {
				request.Header.Del("Content-Type")
				handler.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})
		Context("with invalid XML", func() {
			It("should return 400", func() {
				request.Body = ioutil.NopCloser(strings.NewReader("test"))
				handler.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
		Context("with unexpected root XML element", func() {
			It("should return 400", func() {
				request.Body = ioutil.NopCloser(strings.NewReader("<test><name>myvm</name></test>"))
				handler.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
		Context("with missing VM name", func() {
			It("should return 400", func() {
				request.Body = ioutil.NopCloser(strings.NewReader("<domain><name></name></domain>"))
				handler.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
		Context("with valid XMl", func() {
			It("should return 200", func() {
				handler.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})
		})
		Context("with valid XMl", func() {
			It("should call vmService with the right arguments", func() {
				handler.ServeHTTP(recorder, request)
				Expect(svc.rawXml).To(Equal([]byte("<domain><name>testvm</name></domain>")))
				Expect(svc.VmName).To(Equal("testvm"))
			})
		})
	})

})
