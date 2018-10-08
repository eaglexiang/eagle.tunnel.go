package eagletunnel

import (
	"testing"
)

func Test_Cipher(t *testing.T) {
	c := Cipher{}
	c.SetPassword("password")

	ori := "hello world!"
	bytes := []byte(ori)
	c.Encrypt(bytes)
	dst := string(bytes)
	if ori == dst {
		t.Error(ori, " == ", dst)
	}
	c.Decrypt(bytes)
	dst = string(bytes)
	if ori != dst {
		t.Error(ori, " != ", dst)
	}

	ori = "这是原数据"
	bytes = []byte(ori)
	c.Encrypt(bytes)
	dst = string(bytes)
	if ori == dst {
		t.Error(ori, " == ", dst)
	}
	c.Decrypt(bytes)
	dst = string(bytes)
	if ori != dst {
		t.Error(ori, " != ", dst)
	}

	ori0 := "数据一"
	ori1 := "数据二"
	ori = ori0 + ori1
	bytes0 := []byte(ori0)
	bytes1 := []byte(ori1)
	c.Encrypt(bytes0)
	c.Encrypt(bytes1)
	bytes = append(bytes0, bytes1...)
	dst = string(bytes)
	if ori == dst {
		t.Error(ori, " == ", dst)
	}
	c.Decrypt(bytes)
	dst = string(bytes)
	if ori != dst {
		t.Error(ori, " != ", dst)
	}
}
