package fmutils

import (
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

// NestedMask represents a field mask as a recursive map.
type NestedMask map[string]NestedMask

// NestedMaskFromPaths creates an instance of NestedMask for the given paths.
func NestedMaskFromPaths(paths []string) NestedMask {
	mask := make(NestedMask)
	for _, path := range paths {
		curr := mask
		var letters []rune
		for _, letter := range path {
			if string(letter) == "." {
				if len(letters) == 0 {
					continue
				}

				key := string(letters)
				c, ok := curr[key]
				if ok {
					curr = c
				} else {
					curr[key] = make(NestedMask)
					curr = curr[key]
				}
				letters = nil
				continue
			}
			letters = append(letters, letter)
		}
		if len(letters) != 0 {
			key := string(letters)
			if _, ok := curr[key]; !ok {
				curr[key] = make(NestedMask)
			}
		}
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
			if fd.Kind() == protoreflect.MessageKind {
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
			if fd.Kind() == protoreflect.MessageKind && len(m) != 0 {
				m.Prune(rft.Get(fd).Message().Interface())
				return true
			}
			rft.Clear(fd)
		}
		return true
	})
}
