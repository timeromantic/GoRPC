// Code generated by protoc-gen-go. DO NOT EDIT.
// source: iqiyi.proto

package Proto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ResultList_2_0 struct {
	AdvertiserId         *int64   `protobuf:"varint,1,opt,name=advertiser_id" json:"advertiser_id,omitempty"`
	Type                 *string  `protobuf:"bytes,2,opt,name=type" json:"type,omitempty"`
	IsDeliver            *bool    `protobuf:"varint,3,req,name=is_deliver" json:"is_deliver,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ResultList_2_0) Reset()         { *m = ResultList_2_0{} }
func (m *ResultList_2_0) String() string { return proto.CompactTextString(m) }
func (*ResultList_2_0) ProtoMessage()    {}
func (*ResultList_2_0) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb3748ae221b8b6c, []int{0}
}

func (m *ResultList_2_0) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ResultList_2_0.Unmarshal(m, b)
}
func (m *ResultList_2_0) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ResultList_2_0.Marshal(b, m, deterministic)
}
func (m *ResultList_2_0) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResultList_2_0.Merge(m, src)
}
func (m *ResultList_2_0) XXX_Size() int {
	return xxx_messageInfo_ResultList_2_0.Size(m)
}
func (m *ResultList_2_0) XXX_DiscardUnknown() {
	xxx_messageInfo_ResultList_2_0.DiscardUnknown(m)
}

var xxx_messageInfo_ResultList_2_0 proto.InternalMessageInfo

func (m *ResultList_2_0) GetAdvertiserId() int64 {
	if m != nil && m.AdvertiserId != nil {
		return *m.AdvertiserId
	}
	return 0
}

func (m *ResultList_2_0) GetType() string {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return ""
}

func (m *ResultList_2_0) GetIsDeliver() bool {
	if m != nil && m.IsDeliver != nil {
		return *m.IsDeliver
	}
	return false
}

type RTARequest_2_0 struct {
	Platform             *int32   `protobuf:"varint,1,req,name=platform" json:"platform,omitempty"`
	IdfaMd5              *string  `protobuf:"bytes,2,opt,name=idfa_md5" json:"idfa_md5,omitempty"`
	ImeiMd5              *string  `protobuf:"bytes,3,opt,name=imei_md5" json:"imei_md5,omitempty"`
	AdvertiserIds        []int64  `protobuf:"varint,4,rep,name=advertiser_ids" json:"advertiser_ids,omitempty"`
	Types                []string `protobuf:"bytes,5,rep,name=types" json:"types,omitempty"`
	Age                  *int32   `protobuf:"varint,6,opt,name=age" json:"age,omitempty"`
	Gender               *string  `protobuf:"bytes,7,opt,name=gender" json:"gender,omitempty"`
	City                 *int32   `protobuf:"varint,8,opt,name=city" json:"city,omitempty"`
	RequestIdMd5         *string  `protobuf:"bytes,9,opt,name=request_id_md5" json:"request_id_md5,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RTARequest_2_0) Reset()         { *m = RTARequest_2_0{} }
func (m *RTARequest_2_0) String() string { return proto.CompactTextString(m) }
func (*RTARequest_2_0) ProtoMessage()    {}
func (*RTARequest_2_0) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb3748ae221b8b6c, []int{1}
}

func (m *RTARequest_2_0) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RTARequest_2_0.Unmarshal(m, b)
}
func (m *RTARequest_2_0) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RTARequest_2_0.Marshal(b, m, deterministic)
}
func (m *RTARequest_2_0) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RTARequest_2_0.Merge(m, src)
}
func (m *RTARequest_2_0) XXX_Size() int {
	return xxx_messageInfo_RTARequest_2_0.Size(m)
}
func (m *RTARequest_2_0) XXX_DiscardUnknown() {
	xxx_messageInfo_RTARequest_2_0.DiscardUnknown(m)
}

var xxx_messageInfo_RTARequest_2_0 proto.InternalMessageInfo

func (m *RTARequest_2_0) GetPlatform() int32 {
	if m != nil && m.Platform != nil {
		return *m.Platform
	}
	return 0
}

func (m *RTARequest_2_0) GetIdfaMd5() string {
	if m != nil && m.IdfaMd5 != nil {
		return *m.IdfaMd5
	}
	return ""
}

func (m *RTARequest_2_0) GetImeiMd5() string {
	if m != nil && m.ImeiMd5 != nil {
		return *m.ImeiMd5
	}
	return ""
}

func (m *RTARequest_2_0) GetAdvertiserIds() []int64 {
	if m != nil {
		return m.AdvertiserIds
	}
	return nil
}

func (m *RTARequest_2_0) GetTypes() []string {
	if m != nil {
		return m.Types
	}
	return nil
}

func (m *RTARequest_2_0) GetAge() int32 {
	if m != nil && m.Age != nil {
		return *m.Age
	}
	return 0
}

func (m *RTARequest_2_0) GetGender() string {
	if m != nil && m.Gender != nil {
		return *m.Gender
	}
	return ""
}

func (m *RTARequest_2_0) GetCity() int32 {
	if m != nil && m.City != nil {
		return *m.City
	}
	return 0
}

func (m *RTARequest_2_0) GetRequestIdMd5() string {
	if m != nil && m.RequestIdMd5 != nil {
		return *m.RequestIdMd5
	}
	return ""
}

type RTAResponse_2_0 struct {
	StatusCode           *int32            `protobuf:"varint,1,req,name=status_code" json:"status_code,omitempty"`
	Result               []*ResultList_2_0 `protobuf:"bytes,2,rep,name=result" json:"result,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *RTAResponse_2_0) Reset()         { *m = RTAResponse_2_0{} }
func (m *RTAResponse_2_0) String() string { return proto.CompactTextString(m) }
func (*RTAResponse_2_0) ProtoMessage()    {}
func (*RTAResponse_2_0) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb3748ae221b8b6c, []int{2}
}

func (m *RTAResponse_2_0) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RTAResponse_2_0.Unmarshal(m, b)
}
func (m *RTAResponse_2_0) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RTAResponse_2_0.Marshal(b, m, deterministic)
}
func (m *RTAResponse_2_0) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RTAResponse_2_0.Merge(m, src)
}
func (m *RTAResponse_2_0) XXX_Size() int {
	return xxx_messageInfo_RTAResponse_2_0.Size(m)
}
func (m *RTAResponse_2_0) XXX_DiscardUnknown() {
	xxx_messageInfo_RTAResponse_2_0.DiscardUnknown(m)
}

var xxx_messageInfo_RTAResponse_2_0 proto.InternalMessageInfo

func (m *RTAResponse_2_0) GetStatusCode() int32 {
	if m != nil && m.StatusCode != nil {
		return *m.StatusCode
	}
	return 0
}

func (m *RTAResponse_2_0) GetResult() []*ResultList_2_0 {
	if m != nil {
		return m.Result
	}
	return nil
}

func init() {
	proto.RegisterType((*ResultList_2_0)(nil), "Proto.ResultList_2_0")
	proto.RegisterType((*RTARequest_2_0)(nil), "Proto.RTARequest_2_0")
	proto.RegisterType((*RTAResponse_2_0)(nil), "Proto.RTAResponse_2_0")
}

func init() {
	proto.RegisterFile("iqiyi.proto", fileDescriptor_bb3748ae221b8b6c)
}

var fileDescriptor_bb3748ae221b8b6c = []byte{
	// 272 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0xc1, 0x6a, 0xeb, 0x30,
	0x10, 0x45, 0x91, 0x15, 0xe7, 0xc5, 0xe3, 0xc4, 0xaf, 0xa8, 0x24, 0x68, 0x29, 0x0c, 0x05, 0xad,
	0x4c, 0x09, 0xf4, 0x03, 0xba, 0x2c, 0xb4, 0x50, 0x4c, 0xf7, 0xc2, 0x44, 0x93, 0x30, 0x60, 0xc7,
	0x8e, 0x24, 0x07, 0xfc, 0x57, 0xfd, 0xc4, 0x62, 0xe1, 0x4d, 0x56, 0x03, 0x67, 0x86, 0x3b, 0x87,
	0x0b, 0x39, 0xdd, 0x68, 0xa2, 0x6a, 0x70, 0x7d, 0xe8, 0x45, 0xfa, 0x3d, 0x8f, 0xf2, 0x03, 0x8a,
	0x1a, 0xfd, 0xd8, 0x86, 0x4f, 0xf2, 0xc1, 0x1c, 0xcd, 0xab, 0xd8, 0xc3, 0xae, 0xb1, 0x77, 0x74,
	0x81, 0x3c, 0x3a, 0x43, 0x56, 0x32, 0xc5, 0x34, 0x17, 0x5b, 0x58, 0x85, 0x69, 0x40, 0x99, 0x28,
	0xa6, 0x33, 0x21, 0x00, 0xc8, 0x1b, 0x8b, 0x2d, 0xdd, 0xd1, 0x49, 0xae, 0x12, 0xbd, 0x29, 0x7f,
	0x19, 0x14, 0xf5, 0xcf, 0x7b, 0x8d, 0xb7, 0x11, 0x97, 0xac, 0x27, 0xd8, 0x0c, 0x6d, 0x13, 0xce,
	0xbd, 0xeb, 0x24, 0x53, 0x89, 0x4e, 0x67, 0x42, 0xf6, 0xdc, 0x98, 0xce, 0xbe, 0x2d, 0x51, 0x33,
	0xe9, 0x90, 0x22, 0xe1, 0x91, 0x1c, 0xa0, 0x78, 0x30, 0xf0, 0x72, 0xa5, 0xb8, 0xe6, 0x62, 0x07,
	0xe9, 0xac, 0xe0, 0x65, 0xaa, 0xb8, 0xce, 0x44, 0x0e, 0xbc, 0xb9, 0xa0, 0x5c, 0x2b, 0xa6, 0x53,
	0x51, 0xc0, 0xfa, 0x82, 0x57, 0x8b, 0x4e, 0xfe, 0x8b, 0x19, 0x5b, 0x58, 0x9d, 0x28, 0x4c, 0x72,
	0x13, 0xb7, 0x07, 0x28, 0xdc, 0xa2, 0x45, 0x36, 0x7e, 0xca, 0xe6, 0xab, 0xf2, 0x0b, 0xfe, 0x47,
	0x63, 0x3f, 0xf4, 0x57, 0x8f, 0x51, 0xf9, 0x19, 0x72, 0x1f, 0x9a, 0x30, 0x7a, 0x73, 0xea, 0x2d,
	0x2e, 0xd6, 0x2f, 0xb0, 0x76, 0xb1, 0x25, 0x99, 0x28, 0xae, 0xf3, 0xe3, 0xbe, 0x8a, 0xed, 0x55,
	0x8f, 0xd5, 0xfd, 0x05, 0x00, 0x00, 0xff, 0xff, 0x96, 0x56, 0x32, 0x98, 0x61, 0x01, 0x00, 0x00,
}