package fmutils

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Filter keeps the msg fields that are listed in the paths and clears all the rest.
//
// This is a handy wrapper for NestedMask.Filter method.
// If the same paths are used to process multiple proto messages use NestedMask.Filter method directly.
func Filter(msg proto.Message, paths []string) {
	NestedMaskFromPaths(paths).Filter(msg)
}

// Prune clears all the fields listed in paths from the given msg.
//
// This is a handy wrapper for NestedMask.Prune method.
// If the same paths are used to process multiple proto messages use NestedMask.Filter method directly.
func Prune(msg proto.Message, paths []string) {
	NestedMaskFromPaths(paths).Prune(msg)
}

// Overwrite overwrites all the fields listed in paths in the dest msg using values from src msg.
//
// This is a handy wrapper for NestedMask.Overwrite method.
// If the same paths are used to process multiple proto messages use NestedMask.Overwrite method directly.
func Overwrite(src, dest proto.Message, paths []string) {
	NestedMaskFromPaths(paths).Overwrite(src, dest)
}

// NestedMask represents a field mask as a recursive map.
type NestedMask map[string]NestedMask

// NestedMaskFromPaths creates an instance of NestedMask for the given paths.
//
// For example ["foo.bar", "foo.baz"] becomes {"foo": {"bar": nil, "baz": nil}}.
func NestedMaskFromPaths(paths []string) NestedMask {
	var add func(path string, fm NestedMask)
	add = func(path string, mask NestedMask) {
		if len(path) == 0 {
			// Invalid input.
			return
		}
		dotIdx := strings.IndexRune(path, '.')
		if dotIdx == -1 {
			mask[path] = nil
		} else {
			field := path[:dotIdx]
			if len(field) == 0 {
				// Invalid input.
				return
			}
			rest := path[dotIdx+1:]
			nested := mask[field]
			if nested == nil {
				nested = make(NestedMask)
				mask[field] = nested
			}
			add(rest, nested)
		}
	}

	mask := make(NestedMask)
	for _, p := range paths {
		add(p, mask)
	}

	return mask
}

// Filter keeps the msg fields that are listed in the paths and clears all the rest.
//
// If the mask is empty then all the fields are kept.
// Paths are assumed to be valid and normalized otherwise the function may panic.
// See google.golang.org/protobuf/types/known/fieldmaskpb for details.
func (mask NestedMask) Filter(msg proto.Message) {
	if len(mask) == 0 {
		return
	}

	rft := msg.ProtoReflect()
	rft.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		m, ok := mask[string(fd.Name())]
		if ok {
			if len(m) == 0 {
				return true
			}

			if fd.IsMap() {
				xmap := rft.Get(fd).Map()
				xmap.Range(func(mk protoreflect.MapKey, mv protoreflect.Value) bool {
					if mi, ok := m[mk.String()]; ok {
						if i, ok := mv.Interface().(protoreflect.Message); ok && len(mi) > 0 {
							mi.Filter(i.Interface())
						}
					} else {
						xmap.Clear(mk)
					}

					return true
				})
			} else if fd.IsList() {
				list := rft.Get(fd).List()
				for i := 0; i < list.Len(); i++ {
					m.Filter(list.Get(i).Message().Interface())
				}
			} else if fd.Kind() == protoreflect.MessageKind {
				m.Filter(rft.Get(fd).Message().Interface())
			}
		} else {
			rft.Clear(fd)
		}
		return true
	})
}

// Prune clears all the fields listed in paths from the given msg.
//
// All other fields are kept untouched. If the mask is empty no fields are cleared.
// This operation is the opposite of NestedMask.Filter.
// Paths are assumed to be valid and normalized otherwise the function may panic.
// See google.golang.org/protobuf/types/known/fieldmaskpb for details.
func (mask NestedMask) Prune(msg proto.Message) {
	if len(mask) == 0 {
		return
	}

	rft := msg.ProtoReflect()
	rft.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		m, ok := mask[string(fd.Name())]
		if ok {
			if len(m) == 0 {
				rft.Clear(fd)
				return true
			}

			if fd.IsMap() {
				xmap := rft.Get(fd).Map()
				xmap.Range(func(mk protoreflect.MapKey, mv protoreflect.Value) bool {
					if mi, ok := m[mk.String()]; ok {
						if i, ok := mv.Interface().(protoreflect.Message); ok && len(mi) > 0 {
							mi.Prune(i.Interface())
						} else {
							xmap.Clear(mk)
						}
					}

					return true
				})
			} else if fd.IsList() {
				list := rft.Get(fd).List()
				for i := 0; i < list.Len(); i++ {
					m.Prune(list.Get(i).Message().Interface())
				}
			} else if fd.Kind() == protoreflect.MessageKind {
				m.Prune(rft.Get(fd).Message().Interface())
			}
		}
		return true
	})
}

// Overwrite overwrites all the fields listed in paths in the dest msg using values from src msg.
//
// All other fields are kept untouched. If the mask is empty, no fields are overwritten.
// Supports scalars, messages, repeated fields, and maps.
// If the parent of the field is nil message, the parent is initiated before overwriting the field
// If the field in src is empty value, the field in dest is cleared.
// Paths are assumed to be valid and normalized otherwise the function may panic.
func (mask NestedMask) Overwrite(src, dest proto.Message) {
	mask.overwrite(src.ProtoReflect(), dest.ProtoReflect())
}

func (mask NestedMask) overwrite(srcRft, destRft protoreflect.Message) {
	for srcFDName, submask := range mask {
		srcFD := srcRft.Descriptor().Fields().ByName(protoreflect.Name(srcFDName))
		srcVal := srcRft.Get(srcFD)
		if len(submask) == 0 {
			if isValid(srcFD, srcVal) && !srcVal.Equal(srcFD.Default()) {
				destRft.Set(srcFD, srcVal)
			} else {
				destRft.Clear(srcFD)
			}
		} else if srcFD.IsMap() && srcFD.Kind() == protoreflect.MessageKind {
			srcMap := srcRft.Get(srcFD).Map()
			destMap := destRft.Get(srcFD).Map()
			if !destMap.IsValid() {
				destRft.Set(srcFD, protoreflect.ValueOf(srcMap))
				destMap = destRft.Get(srcFD).Map()
			}
			srcMap.Range(func(mk protoreflect.MapKey, mv protoreflect.Value) bool {
				if mi, ok := submask[mk.String()]; ok {
					if i, ok := mv.Interface().(protoreflect.Message); ok && len(mi) > 0 {
						newVal := protoreflect.ValueOf(i.New())
						destMap.Set(mk, newVal)
						mi.overwrite(mv.Message(), newVal.Message())
					} else {

						destMap.Set(mk, mv)
					}
				} else {
					destMap.Clear(mk)
				}
				return true
			})
		} else if srcFD.IsList() && srcFD.Kind() == protoreflect.MessageKind {
			srcList := srcRft.Get(srcFD).List()
			destList := destRft.Mutable(srcFD).List()
			// Truncate anything in dest that exceeds the length of src
			if srcList.Len() < destList.Len() {
				destList.Truncate(srcList.Len())
			}
			for i := 0; i < srcList.Len(); i++ {
				srcListItem := srcList.Get(i)
				var destListItem protoreflect.Message
				if destList.Len() > i {
					// Overwrite existing items.
					destListItem = destList.Get(i).Message()
				} else {
					// Append new items to overwrite.
					destListItem = destList.AppendMutable().Message()
				}
				submask.overwrite(srcListItem.Message(), destListItem)
			}

		} else if srcFD.Kind() == protoreflect.MessageKind {
			// If the dest field is nil
			if !destRft.Get(srcFD).Message().IsValid() {
				destRft.Set(srcFD, protoreflect.ValueOf(destRft.Get(srcFD).Message().New()))
			}
			submask.overwrite(srcRft.Get(srcFD).Message(), destRft.Get(srcFD).Message())
		}
	}
}

func isValid(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
	if fd.IsMap() {
		return val.Map().IsValid()
	} else if fd.IsList() {
		return val.List().IsValid()
	} else if fd.Message() != nil {
		return val.Message().IsValid()
	}
	return true
}
