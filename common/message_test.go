package common

import (
	"encoding/gob"
	"fmt"
	"reflect"
	"testing"
)

func TestDeepCopy(t *testing.T) {
	msg := &Message{
		ID:   "1",
		Data: []byte("hello"),
	}

	copyMsg := msg.DeepCopy()

	copyMsg.Data.([]byte)[0] = 'T'

	fmt.Println(string(msg.Data.([]byte)))     // hello
	fmt.Println(string(copyMsg.Data.([]byte))) // Hello
}

type MessageInner struct {
	Name string
}

func TestEncodeDecode(t *testing.T) {
	m := Message{
		From: "exec",
		// Msgid:  121212,
		Name:   "message",
		Height: 12,
		// Round:  1,
		Data: "string",
	}
	data := GobEncode(m)

	var m1 Message
	if err := GobDecode(data, &m1); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(m, m1) {
		t.Error("m not equals m1")
	}

	fmt.Printf("m1:%v\n", m1)
	//----------------------------------------------------------
	data = GobEncode(&m)

	var m2 Message
	if err := GobDecode(data, &m2); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(m, m2) {
		t.Error("m not equals m2")
	}
	fmt.Printf("m2:%v\n", m2)
	//----------------------------------------------------------
	gob.Register(&MessageInner{})

	m.Data = MessageInner{"ok"}
	data = GobEncode(m)

	var m3 Message
	if err := GobDecode(data, &m3); err != nil {
		t.Error(err)
	}
	// if !reflect.DeepEqual(m, m3) {
	// 	t.Error("m not equals m3")
	// }
	//Due to the registration above, the decoded result will always be a reference type,so not equals
	fmt.Printf("m:%v,m.data:%v<---->m3:%v,m3.data:%v\n", m, m.Data, m3, m3.Data)

	//----------------------------------------------------------

	m.Data = &MessageInner{"ok"}
	data = GobEncode(m)

	var m4 Message
	if err := GobDecode(data, &m4); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(m, m4) {
		t.Error("m not equals m4")
	}
	fmt.Printf("m4:%v,m4.data:%v\n", m4, m4.Data)
}
