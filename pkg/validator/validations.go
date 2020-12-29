package validator

import (
	"github.com/xn3cr0nx/bitgodine/internal/jwt"

	"github.com/go-playground/validator/v10"
)

// type SignupInfo struct {
// 	Email             string `json:"email" binding:"required"`
// 	Password          string `json:"password" binding:"required"`
// 	FirstName         string `json:"first_name" binding:"required"`
// 	LastName          string `json:"last_name" binding:"required"`
// 	RecaptchaResponse string `json:"g_recaptcha_response" binding:"required"`
// }

// func Recaptcha(c *gin.Context) {
// 	action := "Validation.Recaptcha"
// 	var signup SignupInfo
// 	err := c.ShouldBindJSON(&signup)
// 	if err != nil {
// 		log.LogError(c, action, err, nil)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		c.Abort()
// 		return
// 	}
// 	c.Set("signup", signup)

// 	// if viper.GetBool("captcha.mock") {
// 	c.Next()
// 	return
// 	// }

// 	url := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s&remoteip=%s", viper.GetString("captcha"), signup.RecaptchaResponse, c.ClientIP())
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		log.LogError(c, action, err, nil)
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
// 		c.Abort()
// 		return
// 	}
// 	defer resp.Body.Close()
// 	body, err := ioutil.ReadAll(resp.Body)
// 	var data interface{}
// 	json.Unmarshal(body, &data)
// 	if data.(map[string]interface{})["success"] == false {
// 		err := errors.New(data.(map[string]interface{})["error-codes"].([]interface{})[0].(string))
// 		log.LogError(c, action, err, nil)
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
// 		c.Abort()
// 		return
// 	}

// 	c.Next()
// }

// TestingValidator to test custom validators
func TestingValidator(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) > 5
}

// JWTValidator validates that the field contains a valid jwt token
func JWTValidator(fl validator.FieldLevel) bool {
	token := fl.Field().String()
	if err := jwt.Validate(token); err != nil {
		return false
	}
	return true
}

// LimitValidator to test custom validators
func LimitValidator(fl validator.FieldLevel) bool {
	return fl.Field().Int() < 500 && fl.Field().Int()%5 == 0
}

// PasswordValidator to test custom validators
func PasswordValidator(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) > 6
}
