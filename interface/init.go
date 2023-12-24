package intf

import (
	"github.com/smallnest/rpcx/client"
)

var Factory InterfaceFactory
var Router InterfaceRouter

func init() {
	Factory.registry = make(map[string]InterfaceCreator)
	Router.locals = make(map[string]interface{})
	Router.remotes = make(map[string]client.XClient)
}
