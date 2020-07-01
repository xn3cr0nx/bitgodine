package dgraph_test

import (
	"time"

	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

var TxMock = models.Tx{
	TxID:     "b1fea52486ce0c62bb442b530a3f0132b826c74e473d1f2c220bfa78111c5082",
	Version:  1,
	Locktime: 0,
	Size:     0,
	Weight:   0,
	Fee:      0,
	Vin: []models.Input{
		models.Input{
			TxID:         "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:         4294967295,
			IsCoinbase:   true,
			Scriptsig:    "04ffff001d0102",
			ScriptsigAsm: "OP_PUSHBYTES_4 ffff001d OP_PUSHBYTES_1 02",
			Sequence:     4294967295,
		},
	},
	Vout: []models.Output{
		models.Output{
			Index:               0,
			Scriptpubkey:        "410411DB93E1DCDB8A016B49840F8C53BC1EB68A382E97B1482ECAD7B148A6909A5CB2E0EADDFB84CCF9744464F82E160BFA9B8B64F9D4C03F999B8643F656B412A3AC",
			ScriptpubkeyAddress: "12cbQLTFMXRnSzktFkuoG3eHoMeFtpTu3S",
			ScriptpubkeyAsm:     "0411db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b412a3 OP_CHECKSIG",
			ScriptpubkeyType:    "pubkey",
			Value:               5000000000,
		},
	},
	Status: []models.Status{
		models.Status{
			Confirmed:   true,
			BlockHeight: 170,
			BlockHash:   "00000000d1145790a8694403d4063f323d499e655c83426834d4ce2f8dd4a2ee",
			BlockTime:   time.Now(),
		},
	},
}
