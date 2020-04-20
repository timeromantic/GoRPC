package jmtextProto

import (
	"reflect"
	"testing"
)

func TestRequest(t *testing.T) {

	var giveClass = "classAAA"
	var giveMethod = "methodAAA"
	var giveParams = []interface{}{float64(1), "a"}
	var giveOwlContext = map[string]string{"a": "A", "b": "B"}

	b, err := MarshalReq("user1", "skey", giveClass, giveMethod, giveParams, nil)
	if err != nil {
		t.Fatal(err)
	}

	getClass, getMethod, getparams, getOwlContext, err := UnmarshalReq(b)
	if err != nil {
		t.Fatal(err)
	}

	if giveClass != getClass {
		t.Fatalf("class name not equal, give:%s, get:%s", giveClass, getClass)
	}

	if giveMethod != getMethod {
		t.Fatalf("method name not equal, give:%s, get:%s", giveMethod, getMethod)
	}

	if !reflect.DeepEqual(giveParams, getparams) {
		t.Fatalf("params not equal, give:%+v, get:%+v", giveParams, getparams)
	}

	if !reflect.DeepEqual(giveOwlContext, getOwlContext) {
		t.Fatalf("owlContext not equal, give:%+v, get:%+v", giveOwlContext, getOwlContext)
	}

}
