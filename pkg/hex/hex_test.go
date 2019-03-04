package hex

import (
	"fmt"
	"testing"
)

func TestHexToStream(t *testing.T) {
	str := "410411db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b412a3ac"
	res := HexToStream(str)
	fmt.Println("res", res)
}
func TestHexRotation(t *testing.T) {
	str := "4104b10dd882c04204481116bd4b41510e98c05a869af51376807341fc7e3892c9034835954782295784bfc763d9736ed4122c8bb13d6e02c0882cb7502ce1ae8287ac"
	res := HexRotation(str)
	fmt.Println("res", res)
}
