package disk_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/xn3cr0nx/bitgodine/pkg/badger"
	. "github.com/xn3cr0nx/bitgodine/pkg/disjoint/disk"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var _ = Describe("Disk disjoint set", func() {
	var (
		set *DisjointSet
	)

	BeforeEach(func() {
		logger.Setup()

		conf := badger.Conf("test")
		d, err := NewDisjointSet(conf, true, true)
		Expect(err).ToNot(HaveOccurred())
		set = &d
	})

	AfterEach(func() {
		err := os.RemoveAll("test")
		Expect(err).ToNot(HaveOccurred())
	})

	It("should update size", func() {
		err := set.UpdateSize(1)
		Expect(err).ToNot(HaveOccurred())
		Expect(set.GetSize()).To(Equal(uint32(1)))

		size, err := set.GetStoredSize()
		Expect(err).ToNot(HaveOccurred())
		Expect(size).To(Equal(uint32(1)))
	})

	It("should update height", func() {
		err := set.UpdateHeight(1)
		Expect(err).ToNot(HaveOccurred())
		Expect(set.GetHeight()).To(Equal(int32(1)))

		height, err := set.GetStoredHeight()
		Expect(err).ToNot(HaveOccurred())
		Expect(height).To(Equal(int32(1)))
	})

	It("should update cluster", func() {
		set.MakeSet("1BoatSLRHtKNngkdXEeobR76b53LETtpyT")
		set.MakeSet("1BoatSLRHtKNngkdXEeobR76b53LETtpyF")

		Expect(set.GetSize()).To(Equal(uint32(2)))
		size, err := set.GetStoredSize()
		Expect(err).ToNot(HaveOccurred())
		Expect(size).To(Equal(uint32(2)))

		parents, err := set.GetStoredParents()
		Expect(err).ToNot(HaveOccurred())
		Expect(parents).To(HaveLen(2))

		ranks, err := set.GetStoredRanks()
		Expect(err).ToNot(HaveOccurred())
		Expect(ranks).To(HaveLen(2))

		clusters, err := set.GetStoredClusters()
		Expect(err).ToNot(HaveOccurred())
		Expect(clusters).To(HaveLen(2))
	})

})
