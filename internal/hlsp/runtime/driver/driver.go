package driver

import (
	"context"
	"errors"
	"strings"

	"github.com/hephbuild/heph/internal/hlsp/runtime/builtin"
	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	"github.com/hephbuild/heph/lib/pluginsdk"
	pluginv1 "github.com/hephbuild/heph/plugin/gen/heph/plugin/v1"
	"google.golang.org/protobuf/types/descriptorpb"
)

// SymbolFromTargetDriver calls the driver's Config endpoint and builds
// a Symbol describing its target() schema, prepending the builtin target params.
func SymbolFromTargetDriver(ctx context.Context, drv pluginsdk.Driver) (*symbol.Symbol, error) {
	configReq := pluginv1.ConfigRequest_builder{}.Build()
	resp, err := drv.Config(ctx, configReq)
	if err != nil {
		return nil, err
	}

	return NewSymbolFromTargetSchema(resp)
}

// NewSymbolFromTargetSchema builds a Symbol describing a target() call
// for a specific driver, based on the driver's ConfigResponse schema.
// Builtin target params from GetTarget() are prepended to the driver-specific params.
func NewSymbolFromTargetSchema(resp *pluginv1.ConfigResponse) (*symbol.Symbol, error) {
	schema := resp.GetTargetSchema()
	if schema == nil {
		return nil, errors.New("invalid schema: missing target_schema")
	}

	sym := &symbol.Symbol{
		Name: builtin.TargetName,
		Kind: symbol.FunctionKind,
	}

	// Start with builtin target params (name, deps, out, etc.) if available.
	var builtinParams []*symbol.Parameter
	if targets := builtin.GetTarget(); len(targets) > 0 {
		builtinParams = targets[0].Parameters
	}

	fields := schema.GetField()
	params := make([]*symbol.Parameter, 0, len(builtinParams)+len(fields))
	params = append(params, builtinParams...)

	for _, f := range fields {
		params = append(params, &symbol.Parameter{
			Name: f.GetJsonName(),
			Type: fieldToType(f, schema),
		})
	}

	var sigBuilder strings.Builder
	sigBuilder.WriteString(sym.Name)
	sigBuilder.WriteString("(")
	for i, p := range params {
		if i > 0 {
			sigBuilder.WriteString(", ")
		}
		sigBuilder.WriteString(p.Name)
		if p.Type != nil {
			sigBuilder.WriteString(":" + p.Type.Name)
		}
	}
	sigBuilder.WriteString(")")
	sym.Parameters = params
	sym.Signature = sigBuilder.String()
	sym.Source = resp.GetName()

	return sym, nil
}

// fieldToType converts a FieldDescriptorProto into a *symbol.Symbol representing
// the field's type. For TYPE_MESSAGE it creates a ClassKind symbol with the message name.
func fieldToType(f *descriptorpb.FieldDescriptorProto, parent *descriptorpb.DescriptorProto) *symbol.Symbol {
	if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
		typeName := f.GetTypeName()
		parts := strings.Split(typeName, ".")
		simpleName := parts[len(parts)-1]

		// A proto map field is represented as a repeated message whose
		// descriptor has the map_entry option set.
		for _, nt := range parent.GetNestedType() {
			if nt.GetName() == simpleName && nt.GetOptions().GetMapEntry() {
				return symbol.DictType
			}
		}

		return &symbol.Symbol{Name: simpleName, Kind: symbol.ClassKind}
	}

	// Scalar / primitive types.
	switch f.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return symbol.PrimitiveBool
	case descriptorpb.FieldDescriptorProto_TYPE_STRING,
		descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return symbol.PrimitiveString
	case descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32, descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64, descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32, descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64, descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return symbol.PrimitiveInt
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return symbol.PrimitiveFloat
	default:
		return &symbol.Symbol{Name: f.GetType().String(), Kind: symbol.PrimitiveKind}
	}
}
