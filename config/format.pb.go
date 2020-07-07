// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.3
// source: format.proto

package config

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	General   *General   `protobuf:"bytes,1,opt,name=general,proto3" json:"general,omitempty"`
	Bluetooth *Bluetooth `protobuf:"bytes,2,opt,name=bluetooth,proto3" json:"bluetooth,omitempty"`
	Devices   *Devices   `protobuf:"bytes,3,opt,name=devices,proto3" json:"devices,omitempty"`
	InfluxDb  *InfluxDB  `protobuf:"bytes,100,opt,name=influx_db,json=influxDb,proto3" json:"influx_db,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_format_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_format_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_format_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetGeneral() *General {
	if x != nil {
		return x.General
	}
	return nil
}

func (x *Config) GetBluetooth() *Bluetooth {
	if x != nil {
		return x.Bluetooth
	}
	return nil
}

func (x *Config) GetDevices() *Devices {
	if x != nil {
		return x.Devices
	}
	return nil
}

func (x *Config) GetInfluxDb() *InfluxDB {
	if x != nil {
		return x.InfluxDb
	}
	return nil
}

type General struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EnableDebugLogs   bool  `protobuf:"varint,1,opt,name=enable_debug_logs,json=enableDebugLogs,proto3" json:"enable_debug_logs,omitempty"`
	MaxRefreshRateSec int64 `protobuf:"varint,2,opt,name=max_refresh_rate_sec,json=maxRefreshRateSec,proto3" json:"max_refresh_rate_sec,omitempty"`
	BufferSize        int64 `protobuf:"varint,3,opt,name=buffer_size,json=bufferSize,proto3" json:"buffer_size,omitempty"`
	LogToStdout       bool  `protobuf:"varint,100,opt,name=log_to_stdout,json=logToStdout,proto3" json:"log_to_stdout,omitempty"`
	LogToInflux       bool  `protobuf:"varint,101,opt,name=log_to_influx,json=logToInflux,proto3" json:"log_to_influx,omitempty"`
}

func (x *General) Reset() {
	*x = General{}
	if protoimpl.UnsafeEnabled {
		mi := &file_format_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *General) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*General) ProtoMessage() {}

func (x *General) ProtoReflect() protoreflect.Message {
	mi := &file_format_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use General.ProtoReflect.Descriptor instead.
func (*General) Descriptor() ([]byte, []int) {
	return file_format_proto_rawDescGZIP(), []int{1}
}

func (x *General) GetEnableDebugLogs() bool {
	if x != nil {
		return x.EnableDebugLogs
	}
	return false
}

func (x *General) GetMaxRefreshRateSec() int64 {
	if x != nil {
		return x.MaxRefreshRateSec
	}
	return 0
}

func (x *General) GetBufferSize() int64 {
	if x != nil {
		return x.BufferSize
	}
	return 0
}

func (x *General) GetLogToStdout() bool {
	if x != nil {
		return x.LogToStdout
	}
	return false
}

func (x *General) GetLogToInflux() bool {
	if x != nil {
		return x.LogToInflux
	}
	return false
}

type Bluetooth struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HciId int64 `protobuf:"varint,1,opt,name=hci_id,json=hciId,proto3" json:"hci_id,omitempty"`
}

func (x *Bluetooth) Reset() {
	*x = Bluetooth{}
	if protoimpl.UnsafeEnabled {
		mi := &file_format_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Bluetooth) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Bluetooth) ProtoMessage() {}

func (x *Bluetooth) ProtoReflect() protoreflect.Message {
	mi := &file_format_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Bluetooth.ProtoReflect.Descriptor instead.
func (*Bluetooth) Descriptor() ([]byte, []int) {
	return file_format_proto_rawDescGZIP(), []int{2}
}

func (x *Bluetooth) GetHciId() int64 {
	if x != nil {
		return x.HciId
	}
	return 0
}

type Devices struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RuuviTag []*Devices_RuuviTag `protobuf:"bytes,1,rep,name=ruuvi_tag,json=ruuviTag,proto3" json:"ruuvi_tag,omitempty"`
}

func (x *Devices) Reset() {
	*x = Devices{}
	if protoimpl.UnsafeEnabled {
		mi := &file_format_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Devices) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Devices) ProtoMessage() {}

func (x *Devices) ProtoReflect() protoreflect.Message {
	mi := &file_format_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Devices.ProtoReflect.Descriptor instead.
func (*Devices) Descriptor() ([]byte, []int) {
	return file_format_proto_rawDescGZIP(), []int{3}
}

func (x *Devices) GetRuuviTag() []*Devices_RuuviTag {
	if x != nil {
		return x.RuuviTag
	}
	return nil
}

type InfluxDB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Connection       string `protobuf:"bytes,1,opt,name=connection,proto3" json:"connection,omitempty"`
	Database         string `protobuf:"bytes,2,opt,name=database,proto3" json:"database,omitempty"`
	Table            string `protobuf:"bytes,3,opt,name=table,proto3" json:"table,omitempty"`
	Username         string `protobuf:"bytes,4,opt,name=username,proto3" json:"username,omitempty"`
	Password         string `protobuf:"bytes,5,opt,name=password,proto3" json:"password,omitempty"`
	Precision        string `protobuf:"bytes,6,opt,name=precision,proto3" json:"precision,omitempty"`
	RetentionPolicy  string `protobuf:"bytes,7,opt,name=retention_policy,json=retentionPolicy,proto3" json:"retention_policy,omitempty"`
	WriteConsistency string `protobuf:"bytes,8,opt,name=write_consistency,json=writeConsistency,proto3" json:"write_consistency,omitempty"`
}

func (x *InfluxDB) Reset() {
	*x = InfluxDB{}
	if protoimpl.UnsafeEnabled {
		mi := &file_format_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InfluxDB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InfluxDB) ProtoMessage() {}

func (x *InfluxDB) ProtoReflect() protoreflect.Message {
	mi := &file_format_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InfluxDB.ProtoReflect.Descriptor instead.
func (*InfluxDB) Descriptor() ([]byte, []int) {
	return file_format_proto_rawDescGZIP(), []int{4}
}

func (x *InfluxDB) GetConnection() string {
	if x != nil {
		return x.Connection
	}
	return ""
}

func (x *InfluxDB) GetDatabase() string {
	if x != nil {
		return x.Database
	}
	return ""
}

func (x *InfluxDB) GetTable() string {
	if x != nil {
		return x.Table
	}
	return ""
}

func (x *InfluxDB) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *InfluxDB) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *InfluxDB) GetPrecision() string {
	if x != nil {
		return x.Precision
	}
	return ""
}

func (x *InfluxDB) GetRetentionPolicy() string {
	if x != nil {
		return x.RetentionPolicy
	}
	return ""
}

func (x *InfluxDB) GetWriteConsistency() string {
	if x != nil {
		return x.WriteConsistency
	}
	return ""
}

type Devices_RuuviTag struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Mac       string `protobuf:"bytes,1,opt,name=mac,proto3" json:"mac,omitempty"`
	HumanName string `protobuf:"bytes,2,opt,name=human_name,json=humanName,proto3" json:"human_name,omitempty"`
}

func (x *Devices_RuuviTag) Reset() {
	*x = Devices_RuuviTag{}
	if protoimpl.UnsafeEnabled {
		mi := &file_format_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Devices_RuuviTag) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Devices_RuuviTag) ProtoMessage() {}

func (x *Devices_RuuviTag) ProtoReflect() protoreflect.Message {
	mi := &file_format_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Devices_RuuviTag.ProtoReflect.Descriptor instead.
func (*Devices_RuuviTag) Descriptor() ([]byte, []int) {
	return file_format_proto_rawDescGZIP(), []int{3, 0}
}

func (x *Devices_RuuviTag) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

func (x *Devices_RuuviTag) GetHumanName() string {
	if x != nil {
		return x.HumanName
	}
	return ""
}

var File_format_proto protoreflect.FileDescriptor

var file_format_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa2,
	0x01, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x22, 0x0a, 0x07, 0x67, 0x65, 0x6e,
	0x65, 0x72, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x47, 0x65, 0x6e,
	0x65, 0x72, 0x61, 0x6c, 0x52, 0x07, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x6c, 0x12, 0x28, 0x0a,
	0x09, 0x62, 0x6c, 0x75, 0x65, 0x74, 0x6f, 0x6f, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0a, 0x2e, 0x42, 0x6c, 0x75, 0x65, 0x74, 0x6f, 0x6f, 0x74, 0x68, 0x52, 0x09, 0x62, 0x6c,
	0x75, 0x65, 0x74, 0x6f, 0x6f, 0x74, 0x68, 0x12, 0x22, 0x0a, 0x07, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x73, 0x52, 0x07, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x12, 0x26, 0x0a, 0x09, 0x69,
	0x6e, 0x66, 0x6c, 0x75, 0x78, 0x5f, 0x64, 0x62, 0x18, 0x64, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09,
	0x2e, 0x49, 0x6e, 0x66, 0x6c, 0x75, 0x78, 0x44, 0x42, 0x52, 0x08, 0x69, 0x6e, 0x66, 0x6c, 0x75,
	0x78, 0x44, 0x62, 0x22, 0xcf, 0x01, 0x0a, 0x07, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x6c, 0x12,
	0x2a, 0x0a, 0x11, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x5f,
	0x6c, 0x6f, 0x67, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x65, 0x6e, 0x61, 0x62,
	0x6c, 0x65, 0x44, 0x65, 0x62, 0x75, 0x67, 0x4c, 0x6f, 0x67, 0x73, 0x12, 0x2f, 0x0a, 0x14, 0x6d,
	0x61, 0x78, 0x5f, 0x72, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x5f, 0x72, 0x61, 0x74, 0x65, 0x5f,
	0x73, 0x65, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x11, 0x6d, 0x61, 0x78, 0x52, 0x65,
	0x66, 0x72, 0x65, 0x73, 0x68, 0x52, 0x61, 0x74, 0x65, 0x53, 0x65, 0x63, 0x12, 0x1f, 0x0a, 0x0b,
	0x62, 0x75, 0x66, 0x66, 0x65, 0x72, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x0a, 0x62, 0x75, 0x66, 0x66, 0x65, 0x72, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x22, 0x0a,
	0x0d, 0x6c, 0x6f, 0x67, 0x5f, 0x74, 0x6f, 0x5f, 0x73, 0x74, 0x64, 0x6f, 0x75, 0x74, 0x18, 0x64,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x6c, 0x6f, 0x67, 0x54, 0x6f, 0x53, 0x74, 0x64, 0x6f, 0x75,
	0x74, 0x12, 0x22, 0x0a, 0x0d, 0x6c, 0x6f, 0x67, 0x5f, 0x74, 0x6f, 0x5f, 0x69, 0x6e, 0x66, 0x6c,
	0x75, 0x78, 0x18, 0x65, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x6c, 0x6f, 0x67, 0x54, 0x6f, 0x49,
	0x6e, 0x66, 0x6c, 0x75, 0x78, 0x22, 0x22, 0x0a, 0x09, 0x42, 0x6c, 0x75, 0x65, 0x74, 0x6f, 0x6f,
	0x74, 0x68, 0x12, 0x15, 0x0a, 0x06, 0x68, 0x63, 0x69, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x05, 0x68, 0x63, 0x69, 0x49, 0x64, 0x22, 0x76, 0x0a, 0x07, 0x44, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x12, 0x2e, 0x0a, 0x09, 0x72, 0x75, 0x75, 0x76, 0x69, 0x5f, 0x74, 0x61,
	0x67, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x73, 0x2e, 0x52, 0x75, 0x75, 0x76, 0x69, 0x54, 0x61, 0x67, 0x52, 0x08, 0x72, 0x75, 0x75, 0x76,
	0x69, 0x54, 0x61, 0x67, 0x1a, 0x3b, 0x0a, 0x08, 0x52, 0x75, 0x75, 0x76, 0x69, 0x54, 0x61, 0x67,
	0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d,
	0x61, 0x63, 0x12, 0x1d, 0x0a, 0x0a, 0x68, 0x75, 0x6d, 0x61, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x68, 0x75, 0x6d, 0x61, 0x6e, 0x4e, 0x61, 0x6d,
	0x65, 0x22, 0x8a, 0x02, 0x0a, 0x08, 0x49, 0x6e, 0x66, 0x6c, 0x75, 0x78, 0x44, 0x42, 0x12, 0x1e,
	0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1a,
	0x0a, 0x08, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x61,
	0x62, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x61, 0x62, 0x6c, 0x65,
	0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x72, 0x65, 0x63,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x65,
	0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x29, 0x0a, 0x10, 0x72, 0x65, 0x74, 0x65, 0x6e, 0x74,
	0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0f, 0x72, 0x65, 0x74, 0x65, 0x6e, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x6c, 0x69, 0x63,
	0x79, 0x12, 0x2b, 0x0a, 0x11, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x73, 0x69,
	0x73, 0x74, 0x65, 0x6e, 0x63, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x77, 0x72,
	0x69, 0x74, 0x65, 0x43, 0x6f, 0x6e, 0x73, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x63, 0x79, 0x42, 0x20,
	0x5a, 0x1e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x35, 0x69,
	0x2f, 0x72, 0x75, 0x75, 0x76, 0x69, 0x32, 0x64, 0x62, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_format_proto_rawDescOnce sync.Once
	file_format_proto_rawDescData = file_format_proto_rawDesc
)

func file_format_proto_rawDescGZIP() []byte {
	file_format_proto_rawDescOnce.Do(func() {
		file_format_proto_rawDescData = protoimpl.X.CompressGZIP(file_format_proto_rawDescData)
	})
	return file_format_proto_rawDescData
}

var file_format_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_format_proto_goTypes = []interface{}{
	(*Config)(nil),           // 0: Config
	(*General)(nil),          // 1: General
	(*Bluetooth)(nil),        // 2: Bluetooth
	(*Devices)(nil),          // 3: Devices
	(*InfluxDB)(nil),         // 4: InfluxDB
	(*Devices_RuuviTag)(nil), // 5: Devices.RuuviTag
}
var file_format_proto_depIdxs = []int32{
	1, // 0: Config.general:type_name -> General
	2, // 1: Config.bluetooth:type_name -> Bluetooth
	3, // 2: Config.devices:type_name -> Devices
	4, // 3: Config.influx_db:type_name -> InfluxDB
	5, // 4: Devices.ruuvi_tag:type_name -> Devices.RuuviTag
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_format_proto_init() }
func file_format_proto_init() {
	if File_format_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_format_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_format_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*General); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_format_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Bluetooth); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_format_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Devices); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_format_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InfluxDB); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_format_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Devices_RuuviTag); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_format_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_format_proto_goTypes,
		DependencyIndexes: file_format_proto_depIdxs,
		MessageInfos:      file_format_proto_msgTypes,
	}.Build()
	File_format_proto = out.File
	file_format_proto_rawDesc = nil
	file_format_proto_goTypes = nil
	file_format_proto_depIdxs = nil
}