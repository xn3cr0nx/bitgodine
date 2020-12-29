package mailer

import (
	"errors"

	"github.com/spf13/viper"

	"github.com/sendgrid/rest"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var c *Client

// SGClient interface to define mailer client
type SGClient interface {
	Send(*mail.SGMailV3) (*rest.Response, error)
}

// Config sendgrid configuration options
type Config struct {
	Key   string
	Email string
}

// Client instance wrapper
type Client struct {
	Client SGClient
	Email  string
}

// Conf default sendgrid configuration options
func Conf() *Config {
	return &Config{
		Key:   viper.GetString("mailer.key"),
		Email: viper.GetString("mailer.service"),
	}
}

// NewClient returns a new instance of sendgrid client
func NewClient(conf *Config) (*Client, error) {
	if c != nil {
		return c, nil
	}
	if conf == nil {
		return nil, errors.New("Sendgrid client error")
	}
	c = &Client{
		Client: sendgrid.NewSendClient(conf.Key),
		Email:  conf.Email,
	}
	return c, nil
}

// func Resend(c *echo.Context) {
// 	action := "Mailer.Resend"
// 	log.LogRequest(c, action, nil)
// 	var body struct {
// 		EmailNonce string `json:"email_nonce" binding:"required"`
// 	}
// 	err := c.ShouldBindJSON(&body)
// 	if err != nil {
// 		log.LogError(c, action, err, nil)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	// val, err := db.Redis().HGetAll(fmt.Sprintf("email_nonce:%s", body.EmailNonce)).Result()
// 	val, err := db.Redis().HGetAll(strings.Join([]string{"email_nonce", body.EmailNonce}, ":")).Result()
// 	if err != nil {
// 		log.LogError(c, action, err, nil)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	if len(val) == 0 {
// 		err := errors.New("No such email")
// 		log.LogError(c, action, err, nil)
// 		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 		return
// 	}
// 	switch val["MAILER_FUNCTION"] {
// 	case "UserVerification":
// 		{
// 			_, err := UserVerification(val["name"], val["email"], val["url"], c)
// 			if err != nil {
// 				log.LogError(c, action, err, nil)
// 				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 				return
// 			}
// 		}
// 	case "ResetPassword":
// 		{
// 			_, err := ResetPasswordReq(val["name"], val["surname"], val["email"], val["url"], val["token"], c)
// 			if err != nil {
// 				log.LogError(c, action, err, nil)
// 				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 				return
// 			}
// 		}
// 	}
// 	_, err = db.Redis().Del(strings.Join([]string{"email_nonce", body.EmailNonce}, ":")).Result()
// 	if err != nil {
// 		log.LogError(c, action, err, nil)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"success": true,
// 		"message": "Email resent",
// 	})
// }

// func UserVerification(name, email, url string, c *echo.Context) (*rest.Response, error) {
// 	from := mail.NewEmail("BUULL EXCHANGE", sg.Email)
// 	subject := "Verify your email address"
// 	to := mail.NewEmail(name, email)
// 	request := mail.NewV3Mail()
// 	request.SetFrom(from)
// 	request.Subject = subject
// 	request.SetTemplateID("942897ef-8124-44a1-a906-517c05718774")
// 	subs := mail.NewPersonalization()
// 	subs.AddTos(to)
// 	subs.Subject = subject
// 	subs.SetSubstitution("<%url%>", url)
// 	request.AddPersonalizations(subs)
// 	return sg.Client.Send(request)
// }

// func ResetPasswordReq(name, surname, email, url, token string, c *echo.Context) (*rest.Response, error) {
// 	from := mail.NewEmail("BUULL EXCHANGE", sg.Email)
// 	subject := "Password Recovery Request"
// 	to := mail.NewEmail(name, email)
// 	request := mail.NewV3Mail()
// 	request.SetFrom(from)
// 	request.Subject = subject
// 	request.SetTemplateID("14df6eb3-5d3c-4f16-99ac-7c2cb0575f48")
// 	subs := mail.NewPersonalization()
// 	subs.AddTos(to)
// 	subs.Subject = subject
// 	subs.SetSubstitution("<%url%>", url)
// 	subs.SetSubstitution("<%name%>", name)
// 	subs.SetSubstitution("<%surname%>", surname)
// 	request.AddPersonalizations(subs)
// 	return sg.Client.Send(request)
// }

// func ResetPassword(email string, c *echo.Context) (*rest.Response, error) {
// 	from := mail.NewEmail("BUULL EXCHANGE", sg.Email)
// 	subject := "Password Recovery Successful"
// 	to := mail.NewEmail("", email)
// 	request := mail.NewV3Mail()
// 	request.SetFrom(from)
// 	request.Subject = subject
// 	request.SetTemplateID("3559a6c3-3013-40cc-a84c-0f714b1dd055")
// 	subs := mail.NewPersonalization()
// 	subs.AddTos(to)
// 	subs.Subject = subject
// 	request.AddPersonalizations(subs)
// 	return sg.Client.Send(request)
// }
