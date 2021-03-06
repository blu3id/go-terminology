// Code generated by protoc-gen-go. DO NOT EDIT.
// source: medicine.proto

package medicine

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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Units_PrescribingType int32

const (
	Units_UNDEFINED     Units_PrescribingType = 0
	Units_DOSE_BASED    Units_PrescribingType = 1
	Units_PRODUCT_BASED Units_PrescribingType = 2
)

var Units_PrescribingType_name = map[int32]string{
	0: "UNDEFINED",
	1: "DOSE_BASED",
	2: "PRODUCT_BASED",
}

var Units_PrescribingType_value = map[string]int32{
	"UNDEFINED":     0,
	"DOSE_BASED":    1,
	"PRODUCT_BASED": 2,
}

func (x Units_PrescribingType) String() string {
	return proto.EnumName(Units_PrescribingType_name, int32(x))
}

func (Units_PrescribingType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_c7f0045230e11d70, []int{1, 0}
}

type ParsedMedication struct {
	DrugName             string     `protobuf:"bytes,1,opt,name=drug_name,json=drugName,proto3" json:"drug_name,omitempty"`
	ConceptId            int64      `protobuf:"varint,2,opt,name=concept_id,json=conceptId,proto3" json:"concept_id,omitempty"`
	MappedDrugName       string     `protobuf:"bytes,3,opt,name=mapped_drug_name,json=mappedDrugName,proto3" json:"mapped_drug_name,omitempty"`
	Dose                 float64    `protobuf:"fixed64,4,opt,name=dose,proto3" json:"dose,omitempty"`
	Units                *Units     `protobuf:"bytes,5,opt,name=units,proto3" json:"units,omitempty"`
	Frequency            *Frequency `protobuf:"bytes,6,opt,name=frequency,proto3" json:"frequency,omitempty"`
	Route                *Route     `protobuf:"bytes,7,opt,name=route,proto3" json:"route,omitempty"`
	AsRequired           bool       `protobuf:"varint,8,opt,name=as_required,json=asRequired,proto3" json:"as_required,omitempty"`
	Notes                string     `protobuf:"bytes,9,opt,name=notes,proto3" json:"notes,omitempty"`
	String_              string     `protobuf:"bytes,10,opt,name=string,proto3" json:"string,omitempty"`
	EquivalentDose       float64    `protobuf:"fixed64,11,opt,name=equivalent_dose,json=equivalentDose,proto3" json:"equivalent_dose,omitempty"`
	DailyEquivalentDose  float64    `protobuf:"fixed64,12,opt,name=daily_equivalent_dose,json=dailyEquivalentDose,proto3" json:"daily_equivalent_dose,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *ParsedMedication) Reset()         { *m = ParsedMedication{} }
func (m *ParsedMedication) String() string { return proto.CompactTextString(m) }
func (*ParsedMedication) ProtoMessage()    {}
func (*ParsedMedication) Descriptor() ([]byte, []int) {
	return fileDescriptor_c7f0045230e11d70, []int{0}
}

func (m *ParsedMedication) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ParsedMedication.Unmarshal(m, b)
}
func (m *ParsedMedication) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ParsedMedication.Marshal(b, m, deterministic)
}
func (m *ParsedMedication) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ParsedMedication.Merge(m, src)
}
func (m *ParsedMedication) XXX_Size() int {
	return xxx_messageInfo_ParsedMedication.Size(m)
}
func (m *ParsedMedication) XXX_DiscardUnknown() {
	xxx_messageInfo_ParsedMedication.DiscardUnknown(m)
}

var xxx_messageInfo_ParsedMedication proto.InternalMessageInfo

func (m *ParsedMedication) GetDrugName() string {
	if m != nil {
		return m.DrugName
	}
	return ""
}

func (m *ParsedMedication) GetConceptId() int64 {
	if m != nil {
		return m.ConceptId
	}
	return 0
}

func (m *ParsedMedication) GetMappedDrugName() string {
	if m != nil {
		return m.MappedDrugName
	}
	return ""
}

func (m *ParsedMedication) GetDose() float64 {
	if m != nil {
		return m.Dose
	}
	return 0
}

func (m *ParsedMedication) GetUnits() *Units {
	if m != nil {
		return m.Units
	}
	return nil
}

func (m *ParsedMedication) GetFrequency() *Frequency {
	if m != nil {
		return m.Frequency
	}
	return nil
}

func (m *ParsedMedication) GetRoute() *Route {
	if m != nil {
		return m.Route
	}
	return nil
}

func (m *ParsedMedication) GetAsRequired() bool {
	if m != nil {
		return m.AsRequired
	}
	return false
}

func (m *ParsedMedication) GetNotes() string {
	if m != nil {
		return m.Notes
	}
	return ""
}

func (m *ParsedMedication) GetString_() string {
	if m != nil {
		return m.String_
	}
	return ""
}

func (m *ParsedMedication) GetEquivalentDose() float64 {
	if m != nil {
		return m.EquivalentDose
	}
	return 0
}

func (m *ParsedMedication) GetDailyEquivalentDose() float64 {
	if m != nil {
		return m.DailyEquivalentDose
	}
	return 0
}

type Units struct {
	ConceptId            int64                 `protobuf:"varint,1,opt,name=concept_id,json=conceptId,proto3" json:"concept_id,omitempty"`
	PrescribingType      Units_PrescribingType `protobuf:"varint,2,opt,name=prescribing_type,json=prescribingType,proto3,enum=medicine.Units_PrescribingType" json:"prescribing_type,omitempty"`
	Abbreviations        []string              `protobuf:"bytes,3,rep,name=abbreviations,proto3" json:"abbreviations,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *Units) Reset()         { *m = Units{} }
func (m *Units) String() string { return proto.CompactTextString(m) }
func (*Units) ProtoMessage()    {}
func (*Units) Descriptor() ([]byte, []int) {
	return fileDescriptor_c7f0045230e11d70, []int{1}
}

func (m *Units) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Units.Unmarshal(m, b)
}
func (m *Units) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Units.Marshal(b, m, deterministic)
}
func (m *Units) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Units.Merge(m, src)
}
func (m *Units) XXX_Size() int {
	return xxx_messageInfo_Units.Size(m)
}
func (m *Units) XXX_DiscardUnknown() {
	xxx_messageInfo_Units.DiscardUnknown(m)
}

var xxx_messageInfo_Units proto.InternalMessageInfo

func (m *Units) GetConceptId() int64 {
	if m != nil {
		return m.ConceptId
	}
	return 0
}

func (m *Units) GetPrescribingType() Units_PrescribingType {
	if m != nil {
		return m.PrescribingType
	}
	return Units_UNDEFINED
}

func (m *Units) GetAbbreviations() []string {
	if m != nil {
		return m.Abbreviations
	}
	return nil
}

type Frequency struct {
	ConceptId            int64    `protobuf:"varint,1,opt,name=concept_id,json=conceptId,proto3" json:"concept_id,omitempty"`
	Names                []string `protobuf:"bytes,2,rep,name=names,proto3" json:"names,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Frequency) Reset()         { *m = Frequency{} }
func (m *Frequency) String() string { return proto.CompactTextString(m) }
func (*Frequency) ProtoMessage()    {}
func (*Frequency) Descriptor() ([]byte, []int) {
	return fileDescriptor_c7f0045230e11d70, []int{2}
}

func (m *Frequency) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Frequency.Unmarshal(m, b)
}
func (m *Frequency) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Frequency.Marshal(b, m, deterministic)
}
func (m *Frequency) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Frequency.Merge(m, src)
}
func (m *Frequency) XXX_Size() int {
	return xxx_messageInfo_Frequency.Size(m)
}
func (m *Frequency) XXX_DiscardUnknown() {
	xxx_messageInfo_Frequency.DiscardUnknown(m)
}

var xxx_messageInfo_Frequency proto.InternalMessageInfo

func (m *Frequency) GetConceptId() int64 {
	if m != nil {
		return m.ConceptId
	}
	return 0
}

func (m *Frequency) GetNames() []string {
	if m != nil {
		return m.Names
	}
	return nil
}

type Route struct {
	ConceptId            int64    `protobuf:"varint,1,opt,name=concept_id,json=conceptId,proto3" json:"concept_id,omitempty"`
	Abbreviation         string   `protobuf:"bytes,2,opt,name=abbreviation,proto3" json:"abbreviation,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Route) Reset()         { *m = Route{} }
func (m *Route) String() string { return proto.CompactTextString(m) }
func (*Route) ProtoMessage()    {}
func (*Route) Descriptor() ([]byte, []int) {
	return fileDescriptor_c7f0045230e11d70, []int{3}
}

func (m *Route) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Route.Unmarshal(m, b)
}
func (m *Route) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Route.Marshal(b, m, deterministic)
}
func (m *Route) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Route.Merge(m, src)
}
func (m *Route) XXX_Size() int {
	return xxx_messageInfo_Route.Size(m)
}
func (m *Route) XXX_DiscardUnknown() {
	xxx_messageInfo_Route.DiscardUnknown(m)
}

var xxx_messageInfo_Route proto.InternalMessageInfo

func (m *Route) GetConceptId() int64 {
	if m != nil {
		return m.ConceptId
	}
	return 0
}

func (m *Route) GetAbbreviation() string {
	if m != nil {
		return m.Abbreviation
	}
	return ""
}

func init() {
	proto.RegisterEnum("medicine.Units_PrescribingType", Units_PrescribingType_name, Units_PrescribingType_value)
	proto.RegisterType((*ParsedMedication)(nil), "medicine.ParsedMedication")
	proto.RegisterType((*Units)(nil), "medicine.Units")
	proto.RegisterType((*Frequency)(nil), "medicine.Frequency")
	proto.RegisterType((*Route)(nil), "medicine.Route")
}

func init() { proto.RegisterFile("medicine.proto", fileDescriptor_c7f0045230e11d70) }

var fileDescriptor_c7f0045230e11d70 = []byte{
	// 463 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x93, 0xd1, 0x6a, 0xdb, 0x3e,
	0x14, 0xc6, 0xff, 0x4a, 0x9a, 0xfc, 0xe3, 0x93, 0xc6, 0xf1, 0xd4, 0x6d, 0x08, 0xc6, 0xa8, 0x31,
	0x1b, 0xf3, 0x55, 0x60, 0xd9, 0x0b, 0x6c, 0xad, 0x5d, 0x68, 0x61, 0x69, 0x50, 0x9b, 0x9b, 0xdd,
	0x18, 0xc5, 0x3e, 0x0b, 0x82, 0x5a, 0xf6, 0x24, 0xa7, 0x90, 0x47, 0xdb, 0xc3, 0xec, 0x5d, 0x86,
	0xe5, 0x24, 0x5e, 0x7c, 0xd3, 0x3b, 0x9f, 0xdf, 0xf9, 0xce, 0x87, 0x8e, 0x3e, 0x0b, 0xdc, 0x1c,
	0x33, 0x99, 0x4a, 0x85, 0xb3, 0x52, 0x17, 0x55, 0x41, 0x47, 0x87, 0x3a, 0xf8, 0xdd, 0x07, 0x6f,
	0x29, 0xb4, 0xc1, 0xec, 0x7b, 0x8d, 0x44, 0x25, 0x0b, 0x45, 0xdf, 0x81, 0x93, 0xe9, 0xed, 0x26,
	0x51, 0x22, 0x47, 0x46, 0x7c, 0x12, 0x3a, 0x7c, 0x54, 0x83, 0x85, 0xc8, 0x91, 0xbe, 0x07, 0x48,
	0x0b, 0x95, 0x62, 0x59, 0x25, 0x32, 0x63, 0x3d, 0x9f, 0x84, 0x7d, 0xee, 0xec, 0xc9, 0x6d, 0x46,
	0x43, 0xf0, 0x72, 0x51, 0x96, 0x98, 0x25, 0xad, 0x45, 0xdf, 0x5a, 0xb8, 0x0d, 0x8f, 0x0e, 0x46,
	0x14, 0xce, 0xb2, 0xc2, 0x20, 0x3b, 0xf3, 0x49, 0x48, 0xb8, 0xfd, 0xa6, 0x1f, 0x61, 0xb0, 0x55,
	0xb2, 0x32, 0x6c, 0xe0, 0x93, 0x70, 0x3c, 0x9f, 0xce, 0x8e, 0x07, 0x5f, 0xd5, 0x98, 0x37, 0x5d,
	0xfa, 0x19, 0x9c, 0x9f, 0x1a, 0x7f, 0x6d, 0x51, 0xa5, 0x3b, 0x36, 0xb4, 0xd2, 0x8b, 0x56, 0x7a,
	0x73, 0x68, 0xf1, 0x56, 0x55, 0x3b, 0xeb, 0x62, 0x5b, 0x21, 0xfb, 0xbf, 0xeb, 0xcc, 0x6b, 0xcc,
	0x9b, 0x2e, 0xbd, 0x84, 0xb1, 0x30, 0x49, 0x3d, 0x25, 0x35, 0x66, 0x6c, 0xe4, 0x93, 0x70, 0xc4,
	0x41, 0x18, 0xbe, 0x27, 0xf4, 0x35, 0x0c, 0x54, 0x51, 0xa1, 0x61, 0x8e, 0x5d, 0xaa, 0x29, 0xe8,
	0x5b, 0x18, 0x9a, 0x4a, 0x4b, 0xb5, 0x61, 0x60, 0xf1, 0xbe, 0xa2, 0x9f, 0x60, 0x5a, 0x0f, 0x3e,
	0x8b, 0x27, 0x54, 0x55, 0x62, 0xd7, 0x1d, 0xdb, 0x75, 0xdd, 0x16, 0x47, 0xf5, 0xe2, 0x73, 0x78,
	0x93, 0x09, 0xf9, 0xb4, 0x4b, 0xba, 0xf2, 0x73, 0x2b, 0xbf, 0xb0, 0xcd, 0xf8, 0x64, 0x26, 0xf8,
	0x43, 0x60, 0x60, 0xaf, 0xa5, 0x93, 0x09, 0xe9, 0x66, 0x72, 0x07, 0x5e, 0xa9, 0xd1, 0xa4, 0x5a,
	0xae, 0xa5, 0xda, 0x24, 0xd5, 0xae, 0x44, 0x1b, 0x9c, 0x3b, 0xbf, 0xec, 0x5c, 0xf0, 0x6c, 0xd9,
	0xea, 0x1e, 0x77, 0x25, 0xf2, 0x69, 0x79, 0x0a, 0xe8, 0x07, 0x98, 0x88, 0xf5, 0x5a, 0xe3, 0xb3,
	0xb4, 0xff, 0x8a, 0x61, 0x7d, 0xbf, 0x1f, 0x3a, 0xfc, 0x14, 0x06, 0xd7, 0x30, 0xed, 0x38, 0xd1,
	0x09, 0x38, 0xab, 0x45, 0x14, 0xdf, 0xdc, 0x2e, 0xe2, 0xc8, 0xfb, 0x8f, 0xba, 0x00, 0xd1, 0xfd,
	0x43, 0x9c, 0x5c, 0x7d, 0x7b, 0x88, 0x23, 0x8f, 0xd0, 0x57, 0x30, 0x59, 0xf2, 0xfb, 0x68, 0x75,
	0xfd, 0xb8, 0x47, 0xbd, 0xe0, 0x2b, 0x38, 0xc7, 0x28, 0x5f, 0x5a, 0xb1, 0x8e, 0x45, 0xe4, 0x68,
	0x58, 0xcf, 0x1e, 0xa7, 0x29, 0x82, 0x3b, 0x18, 0xd8, 0x74, 0x5f, 0x9a, 0x0e, 0xe0, 0xfc, 0xdf,
	0xf3, 0xdb, 0xcb, 0x71, 0xf8, 0x09, 0xbb, 0x82, 0x1f, 0xc7, 0x57, 0xb3, 0x1e, 0xda, 0x67, 0xf4,
	0xe5, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xb2, 0xbe, 0x74, 0x1e, 0x58, 0x03, 0x00, 0x00,
}
