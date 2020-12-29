package jwt_test

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	. "github.com/xn3cr0nx/bitgodine/internal/jwt"
)

var _ = Describe("Jwt", func() {
	var (
		config CustomClaims
		token  string
	)

	BeforeEach(func() {
		config = CustomClaims{
			ID:    "1234",
			Email: "dev@bqtx.com",
		}
		var err error
		token, err = NewToken(config.ID, config.Email, 72*time.Hour)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("Generating new jwt token", func() {
		It("should provide default token configuration model", func() {
			conf := Config()
			Expect(conf).ToNot(BeNil())
			Expect(conf.ContextKey).To(Equal("x-token"))
			Expect(conf.SigningKey).To(Equal([]byte(viper.GetString("auth.secret"))))
		})

		It("should generate a new jwt token", func() {
			t, err := NewToken("1234", "dev@bqtx.com", 72*time.Hour)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(t).To(ContainSubstring("."))
		})
	})

	Describe("Validating jwt token", func() {
		It("Should validate a correct jwt token", func() {
			err := Validate(token)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("Should fail validation of invalid jwt token", func() {
			err := Validate("Testing.validate.token")
			Expect(err).To(HaveOccurred())
		})
		It("Should fail validation of malformed jwt token", func() {
			err := Validate(token + "testing")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("signature is invalid"))
		})
	})

	Describe("Decoding jwt token", func() {
		It("Should correctly decode jwt token", func() {
			tk, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(*jwt.Token) (interface{}, error) {
				return []byte(viper.GetString("auth.secret")), nil
			})
			res, err := Decode(tk)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.ID).To(Equal(config.ID))
			Expect(res.Email).To(Equal(config.Email))
		})
	})
})
