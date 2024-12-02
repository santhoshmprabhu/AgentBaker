// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: aksnodeconfig/v1/network_config.proto

package aksnodeconfigv1

import (
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

type NetworkPlugin int32

const (
	NetworkPlugin_NETWORK_PLUGIN_UNSPECIFIED NetworkPlugin = 0
	NetworkPlugin_NETWORK_PLUGIN_NONE        NetworkPlugin = 1
	NetworkPlugin_NETWORK_PLUGIN_AZURE       NetworkPlugin = 2
	NetworkPlugin_NETWORK_PLUGIN_KUBENET     NetworkPlugin = 3
)

// Enum value maps for NetworkPlugin.
var (
	NetworkPlugin_name = map[int32]string{
		0: "NETWORK_PLUGIN_UNSPECIFIED",
		1: "NETWORK_PLUGIN_NONE",
		2: "NETWORK_PLUGIN_AZURE",
		3: "NETWORK_PLUGIN_KUBENET",
	}
	NetworkPlugin_value = map[string]int32{
		"NETWORK_PLUGIN_UNSPECIFIED": 0,
		"NETWORK_PLUGIN_NONE":        1,
		"NETWORK_PLUGIN_AZURE":       2,
		"NETWORK_PLUGIN_KUBENET":     3,
	}
)

func (x NetworkPlugin) Enum() *NetworkPlugin {
	p := new(NetworkPlugin)
	*p = x
	return p
}

func (x NetworkPlugin) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NetworkPlugin) Descriptor() protoreflect.EnumDescriptor {
	return file_aksnodeconfig_v1_network_config_proto_enumTypes[0].Descriptor()
}

func (NetworkPlugin) Type() protoreflect.EnumType {
	return &file_aksnodeconfig_v1_network_config_proto_enumTypes[0]
}

func (x NetworkPlugin) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NetworkPlugin.Descriptor instead.
func (NetworkPlugin) EnumDescriptor() ([]byte, []int) {
	return file_aksnodeconfig_v1_network_config_proto_rawDescGZIP(), []int{0}
}

type NetworkPolicy int32

const (
	NetworkPolicy_NETWORK_POLICY_UNSPECIFIED NetworkPolicy = 0
	NetworkPolicy_NETWORK_POLICY_NONE        NetworkPolicy = 1
	NetworkPolicy_NETWORK_POLICY_AZURE       NetworkPolicy = 2
	NetworkPolicy_NETWORK_POLICY_CALICO      NetworkPolicy = 3
)

// Enum value maps for NetworkPolicy.
var (
	NetworkPolicy_name = map[int32]string{
		0: "NETWORK_POLICY_UNSPECIFIED",
		1: "NETWORK_POLICY_NONE",
		2: "NETWORK_POLICY_AZURE",
		3: "NETWORK_POLICY_CALICO",
	}
	NetworkPolicy_value = map[string]int32{
		"NETWORK_POLICY_UNSPECIFIED": 0,
		"NETWORK_POLICY_NONE":        1,
		"NETWORK_POLICY_AZURE":       2,
		"NETWORK_POLICY_CALICO":      3,
	}
)

func (x NetworkPolicy) Enum() *NetworkPolicy {
	p := new(NetworkPolicy)
	*p = x
	return p
}

func (x NetworkPolicy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NetworkPolicy) Descriptor() protoreflect.EnumDescriptor {
	return file_aksnodeconfig_v1_network_config_proto_enumTypes[1].Descriptor()
}

func (NetworkPolicy) Type() protoreflect.EnumType {
	return &file_aksnodeconfig_v1_network_config_proto_enumTypes[1]
}

func (x NetworkPolicy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NetworkPolicy.Descriptor instead.
func (NetworkPolicy) EnumDescriptor() ([]byte, []int) {
	return file_aksnodeconfig_v1_network_config_proto_rawDescGZIP(), []int{1}
}

type NetworkConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Network plugin to be used by the cluster. Options are NONE, AZURE, KUBENET.
	NetworkPlugin NetworkPlugin `protobuf:"varint,1,opt,name=network_plugin,json=networkPlugin,proto3,enum=aksnodeconfig.v1.NetworkPlugin" json:"network_plugin,omitempty"`
	// Network policy to be used by the cluster.
	// This is still needed to compute ENSURE_NO_DUPE_PROMISCUOUS_BRIDGE.
	// Other than that, it is not used by others. See the discussions here https://github.com/Azure/AgentBaker/pull/4241#discussion_r1554283228
	NetworkPolicy NetworkPolicy `protobuf:"varint,2,opt,name=network_policy,json=networkPolicy,proto3,enum=aksnodeconfig.v1.NetworkPolicy" json:"network_policy,omitempty"`
	// URL to the vnet cni plugins tarball.
	VnetCniPluginsUrl string `protobuf:"bytes,3,opt,name=vnet_cni_plugins_url,json=vnetCniPluginsUrl,proto3" json:"vnet_cni_plugins_url,omitempty"`
	// URL to the cni plugins tarball.
	CniPluginsUrl string `protobuf:"bytes,4,opt,name=cni_plugins_url,json=cniPluginsUrl,proto3" json:"cni_plugins_url,omitempty"`
}

func (x *NetworkConfig) Reset() {
	*x = NetworkConfig{}
	mi := &file_aksnodeconfig_v1_network_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *NetworkConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkConfig) ProtoMessage() {}

func (x *NetworkConfig) ProtoReflect() protoreflect.Message {
	mi := &file_aksnodeconfig_v1_network_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkConfig.ProtoReflect.Descriptor instead.
func (*NetworkConfig) Descriptor() ([]byte, []int) {
	return file_aksnodeconfig_v1_network_config_proto_rawDescGZIP(), []int{0}
}

func (x *NetworkConfig) GetNetworkPlugin() NetworkPlugin {
	if x != nil {
		return x.NetworkPlugin
	}
	return NetworkPlugin_NETWORK_PLUGIN_UNSPECIFIED
}

func (x *NetworkConfig) GetNetworkPolicy() NetworkPolicy {
	if x != nil {
		return x.NetworkPolicy
	}
	return NetworkPolicy_NETWORK_POLICY_UNSPECIFIED
}

func (x *NetworkConfig) GetVnetCniPluginsUrl() string {
	if x != nil {
		return x.VnetCniPluginsUrl
	}
	return ""
}

func (x *NetworkConfig) GetCniPluginsUrl() string {
	if x != nil {
		return x.CniPluginsUrl
	}
	return ""
}

var File_aksnodeconfig_v1_network_config_proto protoreflect.FileDescriptor

var file_aksnodeconfig_v1_network_config_proto_rawDesc = []byte{
	0x0a, 0x25, 0x61, 0x6b, 0x73, 0x6e, 0x6f, 0x64, 0x65, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f,
	0x76, 0x31, 0x2f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x61, 0x6b, 0x73, 0x6e, 0x6f, 0x64, 0x65,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x76, 0x31, 0x22, 0xf8, 0x01, 0x0a, 0x0d, 0x4e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x46, 0x0a, 0x0e, 0x6e,
	0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1f, 0x2e, 0x61, 0x6b, 0x73, 0x6e, 0x6f, 0x64, 0x65, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x50, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x52, 0x0d, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x50, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x12, 0x46, 0x0a, 0x0e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x70,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1f, 0x2e, 0x61, 0x6b,
	0x73, 0x6e, 0x6f, 0x64, 0x65, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x4e,
	0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x0d, 0x6e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x2f, 0x0a, 0x14, 0x76,
	0x6e, 0x65, 0x74, 0x5f, 0x63, 0x6e, 0x69, 0x5f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x5f,
	0x75, 0x72, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x76, 0x6e, 0x65, 0x74, 0x43,
	0x6e, 0x69, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x55, 0x72, 0x6c, 0x12, 0x26, 0x0a, 0x0f,
	0x63, 0x6e, 0x69, 0x5f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x5f, 0x75, 0x72, 0x6c, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x63, 0x6e, 0x69, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x73, 0x55, 0x72, 0x6c, 0x2a, 0x7e, 0x0a, 0x0d, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x1e, 0x0a, 0x1a, 0x4e, 0x45, 0x54, 0x57, 0x4f, 0x52, 0x4b,
	0x5f, 0x50, 0x4c, 0x55, 0x47, 0x49, 0x4e, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46,
	0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x17, 0x0a, 0x13, 0x4e, 0x45, 0x54, 0x57, 0x4f, 0x52, 0x4b,
	0x5f, 0x50, 0x4c, 0x55, 0x47, 0x49, 0x4e, 0x5f, 0x4e, 0x4f, 0x4e, 0x45, 0x10, 0x01, 0x12, 0x18,
	0x0a, 0x14, 0x4e, 0x45, 0x54, 0x57, 0x4f, 0x52, 0x4b, 0x5f, 0x50, 0x4c, 0x55, 0x47, 0x49, 0x4e,
	0x5f, 0x41, 0x5a, 0x55, 0x52, 0x45, 0x10, 0x02, 0x12, 0x1a, 0x0a, 0x16, 0x4e, 0x45, 0x54, 0x57,
	0x4f, 0x52, 0x4b, 0x5f, 0x50, 0x4c, 0x55, 0x47, 0x49, 0x4e, 0x5f, 0x4b, 0x55, 0x42, 0x45, 0x4e,
	0x45, 0x54, 0x10, 0x03, 0x2a, 0x7d, 0x0a, 0x0d, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x50,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x1e, 0x0a, 0x1a, 0x4e, 0x45, 0x54, 0x57, 0x4f, 0x52, 0x4b,
	0x5f, 0x50, 0x4f, 0x4c, 0x49, 0x43, 0x59, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46,
	0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x17, 0x0a, 0x13, 0x4e, 0x45, 0x54, 0x57, 0x4f, 0x52, 0x4b,
	0x5f, 0x50, 0x4f, 0x4c, 0x49, 0x43, 0x59, 0x5f, 0x4e, 0x4f, 0x4e, 0x45, 0x10, 0x01, 0x12, 0x18,
	0x0a, 0x14, 0x4e, 0x45, 0x54, 0x57, 0x4f, 0x52, 0x4b, 0x5f, 0x50, 0x4f, 0x4c, 0x49, 0x43, 0x59,
	0x5f, 0x41, 0x5a, 0x55, 0x52, 0x45, 0x10, 0x02, 0x12, 0x19, 0x0a, 0x15, 0x4e, 0x45, 0x54, 0x57,
	0x4f, 0x52, 0x4b, 0x5f, 0x50, 0x4f, 0x4c, 0x49, 0x43, 0x59, 0x5f, 0x43, 0x41, 0x4c, 0x49, 0x43,
	0x4f, 0x10, 0x03, 0x42, 0x5a, 0x5a, 0x58, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x41, 0x7a, 0x75, 0x72, 0x65, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x62, 0x61, 0x6b,
	0x65, 0x72, 0x2f, 0x61, 0x6b, 0x73, 0x2d, 0x6e, 0x6f, 0x64, 0x65, 0x2d, 0x63, 0x6f, 0x6e, 0x74,
	0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x61,
	0x6b, 0x73, 0x6e, 0x6f, 0x64, 0x65, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x76, 0x31, 0x3b,
	0x61, 0x6b, 0x73, 0x6e, 0x6f, 0x64, 0x65, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x76, 0x31, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_aksnodeconfig_v1_network_config_proto_rawDescOnce sync.Once
	file_aksnodeconfig_v1_network_config_proto_rawDescData = file_aksnodeconfig_v1_network_config_proto_rawDesc
)

func file_aksnodeconfig_v1_network_config_proto_rawDescGZIP() []byte {
	file_aksnodeconfig_v1_network_config_proto_rawDescOnce.Do(func() {
		file_aksnodeconfig_v1_network_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_aksnodeconfig_v1_network_config_proto_rawDescData)
	})
	return file_aksnodeconfig_v1_network_config_proto_rawDescData
}

var file_aksnodeconfig_v1_network_config_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_aksnodeconfig_v1_network_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_aksnodeconfig_v1_network_config_proto_goTypes = []any{
	(NetworkPlugin)(0),    // 0: aksnodeconfig.v1.NetworkPlugin
	(NetworkPolicy)(0),    // 1: aksnodeconfig.v1.NetworkPolicy
	(*NetworkConfig)(nil), // 2: aksnodeconfig.v1.NetworkConfig
}
var file_aksnodeconfig_v1_network_config_proto_depIdxs = []int32{
	0, // 0: aksnodeconfig.v1.NetworkConfig.network_plugin:type_name -> aksnodeconfig.v1.NetworkPlugin
	1, // 1: aksnodeconfig.v1.NetworkConfig.network_policy:type_name -> aksnodeconfig.v1.NetworkPolicy
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_aksnodeconfig_v1_network_config_proto_init() }
func file_aksnodeconfig_v1_network_config_proto_init() {
	if File_aksnodeconfig_v1_network_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_aksnodeconfig_v1_network_config_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_aksnodeconfig_v1_network_config_proto_goTypes,
		DependencyIndexes: file_aksnodeconfig_v1_network_config_proto_depIdxs,
		EnumInfos:         file_aksnodeconfig_v1_network_config_proto_enumTypes,
		MessageInfos:      file_aksnodeconfig_v1_network_config_proto_msgTypes,
	}.Build()
	File_aksnodeconfig_v1_network_config_proto = out.File
	file_aksnodeconfig_v1_network_config_proto_rawDesc = nil
	file_aksnodeconfig_v1_network_config_proto_goTypes = nil
	file_aksnodeconfig_v1_network_config_proto_depIdxs = nil
}
