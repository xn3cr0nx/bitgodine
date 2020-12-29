package mailer

import (
	"github.com/sendgrid/rest"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// MockClient mocks sendgrid client
type MockClient struct {
	sendgrid.Client
}

// NewMockClient returns a mock instance of sendgrid client
func NewMockClient(key string) *MockClient {
	return &MockClient{}
}

// Send mock function that always return positive reponse for send action
func (sg *MockClient) Send(email *mail.SGMailV3) (resp *rest.Response, err error) {
	resp = &rest.Response{
		StatusCode: 200,
	}
	return
}
