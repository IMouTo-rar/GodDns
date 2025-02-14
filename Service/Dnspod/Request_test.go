package Dnspod

import (
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestRequest_GetRecordId(t *testing.T) {

	p := Parameters{
		LoginToken:   "TOKEN",
		Format:       "json",
		Lang:         "cn",
		ErrorOnEmpty: "no",
		Domain:       "domain.com",
		RecordId:     0,
		Subdomain:    "",
		RecordLine:   "默认",
		Value:        "",
		TTL:          600,
		Type:         "A",
	}

	r := Request{
		parameters: p,
	}

	done := make(chan bool)
	status, err := r.GetRecordId(done)
	<-done
	if err != nil {
		t.Error(err)
	}

	t.Log(status)

}

func TestRequest_MakeRequest(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)

	p := Parameters{
		LoginToken:   "TOKEN",
		Format:       "json",
		Lang:         "en",
		ErrorOnEmpty: "no",
		Domain:       "domain.com",
		RecordId:     0,
		Subdomain:    "",
		RecordLine:   "默认",
		Value:        "",
		TTL:          600,
		Type:         "AAAA",
	}

	r := Request{
		parameters: p,
	}
	done := make(chan bool)
	status, err := r.GetRecordId(done)
	select {
	case <-done:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(10 * time.Second):
		t.Failed()
	}

	t.Log(status)

	err = r.MakeRequest()
	if err != nil {
		t.Error(err)
	}

	t.Log(status)
}
