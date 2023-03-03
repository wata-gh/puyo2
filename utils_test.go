package puyo2

import "testing"

// func TestSimpleHands2Puyo(t *testing.T) {
// 	hands := ParseSimpleHands2("by22bb01rr43pb50yb10yr20yy43rp52py13pb42")
// 	fmt.Printf("%+v\n", hands)
// 	res := SimpleHands2Puyop(hands)
// 	fmt.Println(res)
// 	if res != "" {
// 		t.Fatal("res must be xxx")
// 	}
// }

func TestExpandMattulwanParam(t *testing.T) {
	exp := ExpandMattulwanParam("a58babcdbeb3cd2bc2de3")
	if exp != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaababcdbebbbcddbccdeee" {
		t.Fatalf("expand param must be aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaababcdbebbbcddbccdeee but %s", exp)
	}
}
