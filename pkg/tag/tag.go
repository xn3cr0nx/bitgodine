package tag

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/xn3cr0nx/bitgodine_server/internal/tag"
)

// Model structure to wrap checkbitcoinaddress tags
type Model struct {
	tag.Model
}
