package conf

import (
	"github.com/bwmarrin/snowflake"
	"google.golang.org/protobuf/proto"
)

var Marshaler *proto.MarshalOptions
var MarshalerDefault *proto.MarshalOptions

var Unmarshaler *proto.UnmarshalOptions

var SnowlakeNode *snowflake.Node

func Init() {
	var err error
	SnowlakeNode, err = snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	Marshaler = &proto.MarshalOptions{}
	MarshalerDefault = &proto.MarshalOptions{}
	Unmarshaler = &proto.UnmarshalOptions{
		DiscardUnknown: true,
	}
}
