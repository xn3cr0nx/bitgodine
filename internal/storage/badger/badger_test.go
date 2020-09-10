package badger_test

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/storage/badger"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var _ = Describe("Badger", func() {
	var dir string

	BeforeSuite(func() {
		logger.Setup()
	})

	Context("when creating a new instance", func() {
		BeforeEach(func() {
			dir = "./test"
		})

		Context("when the path is permitted", func() {
			var (
				db *badger.Badger
			)

			It("the db is correctly initialized", func() {
				var err error
				db, err = badger.NewBadger(badger.Conf(dir), false)
				Expect(err).ToNot(HaveOccurred())
				Expect(dir).Should(BeADirectory())
			})

			It("the db is corectly cleaned up", func() {
				err := db.Empty()
				Expect(err).ToNot(HaveOccurred())
				Expect(dir).ToNot(BeADirectory())
			})
		})

		Context("when using default path", func() {
			var (
				db *badger.Badger
			)

			BeforeEach(func() {
				viper.SetDefault("db", dir)
			})

			It("the db is correctly initialized", func() {
				var err error
				db, err = badger.NewBadger(badger.Conf(""), false)
				Expect(err).ToNot(HaveOccurred())
				Expect(dir).Should(BeADirectory())
			})

			It("the db is corectly cleaned up", func() {
				err := db.Empty()
				Expect(err).ToNot(HaveOccurred())
				Expect(dir).ToNot(BeADirectory())
			})
		})

		Context("when the path is not permitted", func() {
			It("a permission error is returned", func() {
				dir = "/root/test"
				_, err := badger.NewBadger(badger.Conf(dir), false)
				Expect(err).To(HaveOccurred())
				Expect(strings.Contains(err.Error(), "permission denied")).To(BeTrue())
				Expect(dir).ShouldNot(BeADirectory())
			})
		})
	})

	Context("when inserting new elements in the db", func() {
		var (
			db *badger.Badger
		)

		BeforeEach(func() {
			dir = "./test"
			var err error
			db, err = badger.NewBadger(badger.Conf(dir), false)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := db.Empty()
			Expect(err).ToNot(HaveOccurred())
		})

		It("a generic key value element is correctly inserted", func() {
			key := "test"
			value := "insert"
			err := db.Store(key, []byte(value))
			Expect(err).ToNot(HaveOccurred())
		})

		It("a key value element with uuid key and json encoded value is correctly inserted", func() {
			key := uuid.New()
			value :=
				struct {
					Counter int
					Type    string
				}{10, "test"}

			encoded, err := json.Marshal(&value)
			Expect(err).ToNot(HaveOccurred())
			err = db.Store(key.String(), []byte(encoded))
			Expect(err).ToNot(HaveOccurred())
		})

		It("a batch map of key string and value []byte elements are correctly inserted", func() {
			batch := make(map[string][]byte)
			batch["test"] = []byte("insert")

			value :=
				struct {
					Counter int
					Type    string
				}{10, "test"}
			encoded, err := json.Marshal(&value)
			Expect(err).ToNot(HaveOccurred())

			batch[uuid.New().String()] = encoded
			batch[uuid.New().String()] = encoded

			err = db.StoreBatch(batch)
			Expect(err).ToNot(HaveOccurred())
		})

		It("a batch map of string int key value elements is not inserted and an error is thrown", func() {
			batch := make(map[string]int)
			batch[uuid.New().String()] = 12

			err := db.StoreBatch(batch)
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "type is not allowed")).To(BeTrue())
		})
	})

	Context("when reading from the db", func() {
		var (
			db            *badger.Badger
			UUID, element string
		)

		BeforeEach(func() {
			dir = "./test"
			var err error
			db, err = badger.NewBadger(badger.Conf(dir), false)
			Expect(err).ToNot(HaveOccurred())

			UUID = uuid.New().String()
			element = "test"
			err = db.Store(UUID, []byte(element))
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := db.Empty()
			Expect(err).ToNot(HaveOccurred())
		})

		It("an existing element is correctly retrieved by uuid key", func() {
			e, err := db.Read(UUID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(e)).To(Equal(element))
		})

		It("a not existing element is not retrieved by key", func() {
			_, err := db.Read(uuid.New().String())
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "Key not found")).To(BeTrue())
		})

		It("db keys are correctly retrieve", func() {
			keys, err := db.ReadKeys()
			Expect(err).ToNot(HaveOccurred())
			Expect(keys).To(HaveLen(1))
			Expect(keys).Should(ConsistOf([]string{UUID}))
		})

		It("db keys and values are correctly retrieve", func() {
			keys, err := db.ReadKeyValues()
			Expect(err).ToNot(HaveOccurred())
			Expect(keys).To(HaveLen(1))
			Expect(keys).Should(HaveKeyWithValue(UUID, []byte(element)))
		})

	})

	Context("when updating in the db", func() {
		var (
			db            *badger.Badger
			UUID, element string
		)

		BeforeEach(func() {
			dir = "./test"
			var err error
			db, err = badger.NewBadger(badger.Conf(dir), false)
			Expect(err).ToNot(HaveOccurred())

			UUID = uuid.New().String()
			element = "test"
			err = db.Store(UUID, []byte(element))
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := db.Empty()
			Expect(err).ToNot(HaveOccurred())
		})

		It("an element is correctly updated", func() {
			update := "updated"
			err := db.Store(UUID, []byte(element+update))
			Expect(err).ToNot(HaveOccurred())
			updated, err := db.Read(UUID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(updated)).To(Equal(element + update))
		})
	})

	Context("when deleting from the db", func() {
		var (
			db            *badger.Badger
			UUID, element string
		)

		BeforeEach(func() {
			dir = "./test"
			var err error
			db, err = badger.NewBadger(badger.Conf(dir), false)
			Expect(err).ToNot(HaveOccurred())

			UUID = uuid.New().String()
			element = "test"
			err = db.Store(UUID, []byte(element))
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := db.Empty()
			Expect(err).ToNot(HaveOccurred())
		})

		It("an element is correctly delete", func() {
			err := db.Delete(UUID)
			Expect(err).ToNot(HaveOccurred())
			_, err = db.Read(UUID)
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "Key not found")).To(BeTrue())
		})

		It("using a not existing key is just a NOP", func() {
			err := db.Delete(uuid.New().String())
			Expect(err).ToNot(HaveOccurred())
		})
	})

})
