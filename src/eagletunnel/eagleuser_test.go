package eagletunnel

import (
	"errors"
	"testing"
	"time"
)

func Test_EagleUser_Parse(t *testing.T) {
	compareEagleUser(t, "ABC:P1:100:private", nil, "ABC", "P1", 100, PrivateUser)
	compareEagleUser(t, ":P1:100:private", errors.New("null username"), "", "", 0, 0)
	compareEagleUser(t, "ABC::100:private", errors.New("null password"), "", "", 0, 0)
	compareEagleUser(t, "ABC:P1::private", nil, "ABC", "P1", 0, PrivateUser)
	compareEagleUser(t, "ABC:P1:100:", nil, "ABC", "P1", 100, PrivateUser)
	compareEagleUser(t, "ABC:P1::", nil, "ABC", "P1", 0, PrivateUser)
	compareEagleUser(t, "ABC:P1::share", nil, "ABC", "P1", 0, SharedUser)
	compareEagleUser(t, "ABC:P1:100", nil, "ABC", "P1", 100, PrivateUser)
	compareEagleUser(t, "ABC:P1", nil, "ABC", "P1", 0, PrivateUser)
	compareEagleUser(t, "账户名:密码", nil, "账户名", "密码", 0, PrivateUser)
	compareEagleUser(t, "ABC:DEF:::kasjdflkjafl", nil, "ABC", "DEF", 0, PrivateUser)
	compareEagleUser(t, "ABC:P1:ABC", errors.New("when parse EagleUser: strconv.ParseInt: parsing \"ABC\": invalid syntax"), "", "", 0, 0)
	compareEagleUser(t, "ABC:P1:-1", errors.New("speed limit must be bigger than or equal to 0"), "", "", 0, 0)
	compareEagleUser(t, "ABC:P1::haha", errors.New("unknown user type"), "", "", 0, 0)
}

func compareEagleUser(t *testing.T, input string, theErr error, id string, password string, speedlimit int64, theType int) {
	t.Log("开始检查：", input)
	ip := "127.0.0.1"
	user, err := ParseEagleUser(input, ip)
	if err == nil {
		if theErr != nil {
			t.Error("本该有报错：", theErr.Error())
		} else {
			if user.ID != id {
				t.Error("ID为", id)
			}
			if user.Password != password {
				t.Error("密码为", password)
			}
			if user.pause == nil {
				t.Error("pause为nil")
			} else {
				if *user.pause == true {
					t.Error("pause被初始化为true")
				}
			}
			if user.speed != 0 {
				t.Error("速度应该被初始化为0")
			}
			if user.speedLimit != speedlimit {
				t.Error("限速不匹配")
			}
			output := user.toString()
			if output != id+":"+password {
				t.Error("错误的输出：", output)
			}
			if user.bytes != 0 {
				t.Error("bytes应该初始化为0")
			}
			if user.lastIP != ip {
				t.Error("错误的IP")
			}
			if user.typeOfUser != theType {
				t.Error("错误的类型")
			}
		}
	} else {
		if theErr == nil {
			t.Error("不应该有的报错：", err.Error())
		} else {
			if err.Error() != theErr.Error() {
				t.Error("错误的报错内容：", err.Error(), "正确的报错内容：", theErr.Error())
			}
		}
	}
}

func Test_EagleUser_Check(t *testing.T) {
	validUser, _ := ParseEagleUser("abc:jsl*", "")
	user2Check, _ := ParseEagleUser("abc:jsl", "127.0.0.1")
	err := validUser.CheckAuth(user2Check)
	if err == nil {
		t.Error("未发现错误密码")
	}
	user2Check.Password = "jsl*"
	err = validUser.CheckAuth(user2Check)
	if err != nil {
		t.Error("不必要的报错：", err.Error())
	}
	user2Check.lastIP = "192.168.0.1"
	err = validUser.CheckAuth(user2Check)
	if err == nil {
		t.Error("未识别已改变的IP")
	}
	user2Check.lastIP = "192.168.0.2"
	user2Check.lastTime = user2Check.lastTime.Add(4 * time.Minute)
	err = validUser.CheckAuth(user2Check)
	if err != nil {
		t.Error("未识别已过3分钟")
	}
	user2Check.lastIP = "192.168.0.3"
	user2Check.lastTime = user2Check.lastTime.Add(2 * time.Minute)
	err = validUser.CheckAuth(user2Check)
	if err == nil {
		t.Error("时间只过了2分钟")
	}
	user2Check.lastIP = "192.168.0.4"
	err = validUser.CheckAuth(user2Check)
	if err == nil {
		t.Error("未识别已改变的IP")
	}
	err = validUser.CheckAuth(user2Check)
	if err == nil {
		t.Error("未识别已改变的IP")
	}
	validUser, _ = ParseEagleUser("abc:jsl*::share", "")
	err = validUser.CheckAuth(user2Check)
	if err != nil {
		t.Error("不该有的报错：", err.Error())
	}
	user2Check.lastIP = "192.168.0.5"
	err = validUser.CheckAuth(user2Check)
	if err != nil {
		t.Error("shared账户不应该限制多地同时登录")
	}
	validUser, _ = ParseEagleUser("abc:jsl*::shared", "")
	err = validUser.CheckAuth(user2Check)
	if err != nil {
		t.Error("不该有的报错：", err.Error())
	}
	user2Check.lastIP = "192.168.0.6"
	err = validUser.CheckAuth(user2Check)
	if err != nil {
		t.Error("shared账户不应该限制多地同时登录")
	}
}
