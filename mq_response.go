package xipc

import (
	"github.com/joe-at-startupmedia/xipc/protos"
)

type MqResponse protos.Response

// AsProtobuf used to convert the local type equivalent (MqResponse)
// back to its protobuf instance
func (mqr *MqResponse) AsProtobuf() *protos.Response {
	return (*protos.Response)(mqr)
}

func (mqr *MqResponse) PrepareFromRequest(mqs *MqRequest) *MqResponse {
	mqr.RequestId = mqs.Id
	return mqr
}

// ProtoResponseToMqResponse used to convert the protobuf to the local
// type equivalent (MqResponse) for leveraging instance methods
func ProtoResponseToMqResponse(mqr *protos.Response) *MqResponse {
	return (*MqResponse)(mqr)
}
