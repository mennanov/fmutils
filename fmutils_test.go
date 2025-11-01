package fmutils

import (
	"reflect"
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/mennanov/fmutils/testproto"
)

func Test_NestedMaskFromPaths(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		args args
		want NestedMask
	}{
		{
			name: "no nested fields",
			args: args{paths: []string{"a", "b", "c"}},
			want: NestedMask{"a": nil, "b": nil, "c": nil},
		},
		{
			name: "with nested fields",
			args: args{paths: []string{"aaa.bb.c", "dd.e", "f"}},
			want: NestedMask{
				"aaa": NestedMask{"bb": NestedMask{"c": nil}},
				"dd":  NestedMask{"e": nil},
				"f":   nil},
		},
		{
			name: "single field",
			args: args{paths: []string{"a"}},
			want: NestedMask{"a": nil},
		},
		{
			name: "empty fields",
			args: args{paths: []string{}},
			want: NestedMask{},
		},
		{
			name: "invalid input",
			args: args{paths: []string{".", "..", "...", ".a.", ""}},
			want: NestedMask{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NestedMaskFromPaths(tt.args.paths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NestedMaskFromPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createAny(m proto.Message) *anypb.Any {
	any, err := anypb.New(m)
	if err != nil {
		panic(err)
	}
	return any
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		msg   proto.Message
		want  proto.Message
	}{
		{
			name:  "empty mask keeps all the fields",
			paths: []string{},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
				LoginTimestamps: []int64{1, 2},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
				LoginTimestamps: []int64{1, 2},
			},
		},
		{
			name:  "path with empty string is ignored",
			paths: []string{""},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
			},
		},
		{
			name:  "mask with all root fields keeps all root fields",
			paths: []string{"user", "photo"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
		},
		{
			name:  "mask with single root field keeps that field only",
			paths: []string{"user"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
			},
		},
		{
			name:  "mask with nested fields keeps the listed fields only",
			paths: []string{"user.name", "photo.path", "photo.dimensions.width"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					Name: "user name",
				},
				Photo: &testproto.Photo{
					Path: "photo path",
					Dimensions: &testproto.Dimensions{
						Width: 100,
					},
				},
			},
		},
		{
			name:  "mask with oneof field keeps the entire field",
			paths: []string{"user"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_User{User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				}},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_User{User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				}},
			},
		},
		{
			name:  "mask with nested oneof fields keeps listed fields only",
			paths: []string{"profile.photo.dimensions", "profile.user.user_id", "profile.login_timestamps"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{Profile: &testproto.Profile{
					User: &testproto.User{
						UserId: 1,
						Name:   "user name",
					},
					Photo: &testproto.Photo{
						PhotoId: 1,
						Path:    "photo path",
						Dimensions: &testproto.Dimensions{
							Width:  100,
							Height: 120,
						},
					},
					LoginTimestamps: []int64{1, 2, 3},
				}},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{Profile: &testproto.Profile{
					User: &testproto.User{
						UserId: 1,
					},
					Photo: &testproto.Photo{
						Dimensions: &testproto.Dimensions{
							Width:  100,
							Height: 120,
						},
					},
					LoginTimestamps: []int64{1, 2, 3},
				}},
			},
		},
		{
			name:  "mask with Any field in oneof field keeps the entire Any field",
			paths: []string{"details"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Details{Details: createAny(&testproto.Result{
					Data:      []byte("bytes"),
					NextToken: 1,
				})},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Details{Details: createAny(&testproto.Result{
					Data:      []byte("bytes"),
					NextToken: 1,
				})},
			},
		},
		{
			name:  "mask with repeated nested fields keeps the listed fields",
			paths: []string{"profile.gallery.photo_id", "profile.gallery.dimensions.height"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Dimensions: &testproto.Dimensions{
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Dimensions: &testproto.Dimensions{
									Height: 400,
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mask with repeated field keeps the listed field only",
			paths: []string{"profile.gallery"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mask with map field keeps the listed field only",
			paths: []string{"profile.attributes.a1", "profile.attributes.a2.tags.t2", "profile.attributes.aNonExistant"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t2": "2",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Filter(tt.msg, tt.paths)
			if !proto.Equal(tt.msg, tt.want) {
				t.Errorf("msg %v, want %v", tt.msg, tt.want)
			}
		})
	}
}

func TestPrune(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		msg   proto.Message
		want  proto.Message
	}{
		{
			name:  "empty mask keeps all the fields",
			paths: []string{},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
		},
		{
			name:  "mask all root fields clears all fields",
			paths: []string{"user", "photo"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{},
		},
		{
			name:  "mask with single root field clears that field only",
			paths: []string{"user"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
		},
		{
			name:  "mask with nested fields clears that fields only",
			paths: []string{"user.name", "photo.path", "photo.dimensions.width"},
			msg: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Path:    "photo path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 120,
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 1,
				},
				Photo: &testproto.Photo{
					PhotoId: 2,
					Dimensions: &testproto.Dimensions{
						Height: 120,
					},
				},
			},
		},
		{
			name:  "mask with oneof field clears that entire field only",
			paths: []string{"user"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_User{User: &testproto.User{
					UserId: 1,
					Name:   "user name",
				}},
			},
			want: &testproto.Event{
				EventId: 1,
			},
		},
		{
			name:  "mask with nested oneof fields clears listed fields only",
			paths: []string{"profile.photo.dimensions", "profile.user.user_id", "profile.login_timestamps"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{Profile: &testproto.Profile{
					User: &testproto.User{
						UserId: 1,
						Name:   "user name",
					},
					Photo: &testproto.Photo{
						PhotoId: 1,
						Path:    "photo path",
						Dimensions: &testproto.Dimensions{
							Width:  100,
							Height: 120,
						},
					},
					LoginTimestamps: []int64{1, 2, 3},
				}},
			},
			want: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{Profile: &testproto.Profile{
					User: &testproto.User{
						Name: "user name",
					},
					Photo: &testproto.Photo{
						PhotoId: 1,
						Path:    "photo path",
					},
				}},
			},
		},
		{
			name:  "mask with repeated nested fields clears the listed fields",
			paths: []string{"profile.gallery.photo_id", "profile.gallery.dimensions.height"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								Path: "path 1",
								Dimensions: &testproto.Dimensions{
									Width: 100,
								},
							},
							{
								Path: "path 2",
								Dimensions: &testproto.Dimensions{
									Width: 300,
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mask with repeated field clears the listed field only",
			paths: []string{"profile.gallery"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
						Gallery: []*testproto.Photo{
							{
								PhotoId: 1,
								Path:    "path 1",
								Dimensions: &testproto.Dimensions{
									Width:  100,
									Height: 200,
								},
							},
							{
								PhotoId: 2,
								Path:    "path 2",
								Dimensions: &testproto.Dimensions{
									Width:  300,
									Height: 400,
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Photo: &testproto.Photo{
							PhotoId: 4,
							Path:    "photo path",
						},
					},
				},
			},
		},
		{
			name:  "mask with map field prunes the listed field",
			paths: []string{"profile.attributes.a1", "profile.attributes.a2.tags.t2", "profile.attributes.aNonExistant"},
			msg: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a1": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a2": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
						},
					},
				},
			},
			want: &testproto.Event{
				EventId: 1,
				Changed: &testproto.Event_Profile{
					Profile: &testproto.Profile{
						Attributes: map[string]*testproto.Attribute{
							"a2": {
								Tags: map[string]string{
									"t1": "1",
									"t3": "3",
								},
							},
							"a3": {
								Tags: map[string]string{
									"t1": "1",
									"t2": "2",
									"t3": "3",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Prune(tt.msg, tt.paths)
			if !proto.Equal(tt.msg, tt.want) {
				t.Errorf("msg %v, want %v", tt.msg, tt.want)
			}
		})
	}
}

func TestOverwrite(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		src   proto.Message
		dest  proto.Message
		want  proto.Message
	}{
		{
			name: "overwrite scalar/message/map/list",
			paths: []string{
				"user.user_id", "photo", "login_timestamps", "attributes",
			},
			src: &testproto.Profile{
				User: &testproto.User{
					UserId: 567,
					Name:   "different-name",
				},
				Photo: &testproto.Photo{
					Path: "photo-path",
				},
				LoginTimestamps: []int64{1, 2, 3},
				Attributes: map[string]*testproto.Attribute{
					"src": {},
				},
			},
			dest: &testproto.Profile{
				User: &testproto.User{
					Name: "name",
				},
				LoginTimestamps: []int64{4},
				Attributes: map[string]*testproto.Attribute{
					"dest": {},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					UserId: 567,
					Name:   "name",
				},
				Photo: &testproto.Photo{
					Path: "photo-path",
				},
				LoginTimestamps: []int64{1, 2, 3},
				Attributes: map[string]*testproto.Attribute{
					"src": {},
				},
			},
		},
		{
			name:  "field inside nil message",
			paths: []string{"photo.path"},
			src: &testproto.Profile{
				Photo: &testproto.Photo{
					Path: "photo-path",
				},
			},
			dest: &testproto.Profile{
				Photo: nil,
			},
			want: &testproto.Profile{
				Photo: &testproto.Photo{
					Path: "photo-path",
				},
			},
		},
		{
			name:  "empty message/map/list fields",
			paths: []string{"user", "photo.photo_id", "attributes", "login_timestamps"},

			src: &testproto.Profile{
				User: nil, // Empty message
				Photo: &testproto.Photo{
					PhotoId: 0, // Empty scalar
				},
				Attributes:      make(map[string]*testproto.Attribute), // Empty map
				LoginTimestamps: make([]int64, 0),                      // Empty list
			},
			dest: &testproto.Profile{
				User: &testproto.User{
					Name: "name",
				},
				Photo: &testproto.Photo{
					PhotoId: 1234,
				},
				Attributes: map[string]*testproto.Attribute{
					"attribute": {
						Tags: map[string]string{
							"tag": "val",
						},
					},
				},
				LoginTimestamps: []int64{1, 2, 3},
				Gallery: []*testproto.Photo{
					{
						PhotoId: 567,
						Path:    "path",
					},
				},
			},
			want: &testproto.Profile{
				User: nil, // Empty message
				Photo: &testproto.Photo{
					PhotoId: 0, // Empty scalar
				},
				Attributes:      make(map[string]*testproto.Attribute), // Empty map
				LoginTimestamps: make([]int64, 0),                      // Empty list
				Gallery: []*testproto.Photo{
					{
						PhotoId: 567,
						Path:    "path",
					},
				},
			},
		},
		{
			name:  "overwrite map with message values",
			paths: []string{"attributes.src1.tags.key1", "attributes.src2"},
			src: &testproto.Profile{
				User: nil,
				Attributes: map[string]*testproto.Attribute{
					"src1": {
						Tags: map[string]string{"key1": "value1", "key2": "value2"},
					},
					"src2": {
						Tags: map[string]string{"key3": "value3"},
					},
				},
			},
			dest: &testproto.Profile{
				User: &testproto.User{
					Name: "name",
				},
				Attributes: map[string]*testproto.Attribute{
					"dest1": {
						Tags: map[string]string{"key4": "value4"},
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					Name: "name",
				},
				Attributes: map[string]*testproto.Attribute{
					"src1": {
						Tags: map[string]string{"key1": "value1"},
					},
					"src2": {
						Tags: map[string]string{"key3": "value3"},
					},
					"dest1": {
						Tags: map[string]string{"key4": "value4"},
					},
				},
			},
		},
		{
			name:  "overwrite repeated message fields",
			paths: []string{"gallery.path"},
			src: &testproto.Profile{
				User: &testproto.User{
					UserId: 567,
					Name:   "different-name",
				},
				Photo: &testproto.Photo{
					Path: "photo-path",
				},
				LoginTimestamps: []int64{1, 2, 3},
				Attributes: map[string]*testproto.Attribute{
					"src": {},
				},
				Gallery: []*testproto.Photo{
					{
						PhotoId: 123,
						Path:    "test-path-1",
						Dimensions: &testproto.Dimensions{
							Width:  345,
							Height: 456,
						},
					},
					{
						PhotoId: 234,
						Path:    "test-path-2",
						Dimensions: &testproto.Dimensions{
							Width:  3456,
							Height: 4567,
						},
					},
					{
						PhotoId: 345,
						Path:    "test-path-3",
						Dimensions: &testproto.Dimensions{
							Width:  34567,
							Height: 45678,
						},
					},
				},
			},
			dest: &testproto.Profile{
				User: &testproto.User{
					Name: "name",
				},
				Gallery: []*testproto.Photo{
					{
						PhotoId: 123,
						Path:    "test-path-7",
						Dimensions: &testproto.Dimensions{
							Width:  345,
							Height: 456,
						},
					},
					{
						PhotoId: 234,
						Path:    "test-path-6",
						Dimensions: &testproto.Dimensions{
							Width:  3456,
							Height: 4567,
						},
					},
					{
						PhotoId: 345,
						Path:    "test-path-5",
						Dimensions: &testproto.Dimensions{
							Width:  34567,
							Height: 45678,
						},
					},
					{
						PhotoId: 345,
						Path:    "test-path-4",
						Dimensions: &testproto.Dimensions{
							Width:  34567,
							Height: 45678,
						},
					},
				},
			},
			want: &testproto.Profile{
				User: &testproto.User{
					Name: "name",
				},
				Gallery: []*testproto.Photo{
					{
						PhotoId: 123,
						Path:    "test-path-1",
						Dimensions: &testproto.Dimensions{
							Width:  345,
							Height: 456,
						},
					},
					{
						PhotoId: 234,
						Path:    "test-path-2",
						Dimensions: &testproto.Dimensions{
							Width:  3456,
							Height: 4567,
						},
					},
					{
						PhotoId: 345,
						Path:    "test-path-3",
						Dimensions: &testproto.Dimensions{
							Width:  34567,
							Height: 45678,
						},
					},
				},
			},
		},
		{
			name:  "overwrite repeated message fields to empty list",
			paths: []string{"gallery.path"},
			src: &testproto.Profile{
				User: &testproto.User{
					UserId: 567,
					Name:   "different-name",
				},
				Photo: &testproto.Photo{
					Path: "photo-path",
				},
				LoginTimestamps: []int64{1, 2, 3},
				Attributes: map[string]*testproto.Attribute{
					"src": {},
				},
				Gallery: []*testproto.Photo{
					{
						PhotoId: 123,
						Path:    "test-path-1",
						Dimensions: &testproto.Dimensions{
							Width:  345,
							Height: 456,
						},
					},
					{
						PhotoId: 234,
						Path:    "test-path-2",
						Dimensions: &testproto.Dimensions{
							Width:  3456,
							Height: 4567,
						},
					},
					{
						PhotoId: 345,
						Path:    "test-path-3",
						Dimensions: &testproto.Dimensions{
							Width:  34567,
							Height: 45678,
						},
					},
				},
			},
			dest: &testproto.Profile{},
			want: &testproto.Profile{
				Gallery: []*testproto.Photo{
					{
						Path: "test-path-1",
					},
					{
						Path: "test-path-2",
					},
					{
						Path: "test-path-3",
					},
				},
			},
		},
		{
			name:  "overwrite optional fields with nil",
			paths: []string{"optional_string", "optional_int", "optional_photo", "optional_attr"},
			src: &testproto.Options{
				OptionalString: nil,
				OptionalInt:    nil,
				OptionalPhoto:  nil,
				OptionalAttr:   nil,
			},
			dest: &testproto.Options{
				OptionalString: func(s string) *string { return &s }("optional string"),
				OptionalInt:    func(n int32) *int32 { return &n }(10),
				OptionalPhoto: &testproto.Photo{
					PhotoId: 123,
					Path:    "test-path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 200,
					},
				},
				OptionalAttr: &testproto.Attribute{
					Tags: map[string]string{
						"a": "b",
						"c": "d",
					},
				},
			},
			want: &testproto.Options{
				OptionalString: nil,
				OptionalInt:    nil,
				OptionalPhoto:  nil,
				OptionalAttr:   nil,
			},
		},
		{
			name:  "overwrite empty optional field with a value",
			paths: []string{"optional_string", "optional_int", "optional_photo", "optional_attr"},
			src: &testproto.Options{
				OptionalString: func(s string) *string { return &s }("optional string"),
				OptionalInt:    func(n int32) *int32 { return &n }(10),
				OptionalPhoto: &testproto.Photo{
					PhotoId: 123,
					Path:    "test-path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 200,
					},
				},
				OptionalAttr: &testproto.Attribute{
					Tags: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			dest: &testproto.Options{},
			want: &testproto.Options{
				OptionalString: func(s string) *string { return &s }("optional string"),
				OptionalInt:    func(n int32) *int32 { return &n }(10),
				OptionalPhoto: &testproto.Photo{
					PhotoId: 123,
					Path:    "test-path",
					Dimensions: &testproto.Dimensions{
						Width:  100,
						Height: 200,
					},
				},
				OptionalAttr: &testproto.Attribute{
					Tags: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Overwrite(tt.src, tt.dest, tt.paths)
			if !proto.Equal(tt.dest, tt.want) {
				t.Errorf("dest %v, want %v", tt.dest, tt.want)
			}
		})
	}
}

func TestFromFieldNumbers(t *testing.T) {
	tests := []struct {
		name         string
		msg          proto.Message
		fieldNumbers []int
		want         []string
	}{
		{
			name:         "empty field numbers",
			msg:          &testproto.User{},
			fieldNumbers: []int{},
			want:         nil,
		},
		{
			name:         "single field number",
			msg:          &testproto.User{},
			fieldNumbers: []int{1},
			want:         []string{"user_id"},
		},
		{
			name:         "multiple field numbers",
			msg:          &testproto.User{},
			fieldNumbers: []int{1, 2},
			want:         []string{"user_id", "name"},
		},
		{
			name:         "field numbers with Profile",
			msg:          &testproto.Profile{},
			fieldNumbers: []int{1, 2, 3},
			want:         []string{"user", "photo", "login_timestamps"},
		},
		{
			name:         "non-existent field number",
			msg:          &testproto.User{},
			fieldNumbers: []int{999},
			want:         []string{},
		},
		{
			name:         "mixed valid and invalid field numbers",
			msg:          &testproto.User{},
			fieldNumbers: []int{1, 999, 2},
			want:         []string{"user_id", "name"},
		},
		{
			name:         "duplicate field numbers",
			msg:          &testproto.User{},
			fieldNumbers: []int{1, 1, 2},
			want:         []string{"user_id", "user_id", "name"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PathsFromFieldNumbers(tt.msg, tt.fieldNumbers...)
			slices.Sort(got)
			slices.Sort(tt.want)
			if !slices.Equal(got, tt.want) {
				t.Errorf("PathsFromFieldNumbers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		msg     proto.Message
		paths   []string
		wantErr bool
	}{
		{
			name: "empty mask",
			msg:  &testproto.Profile{},
		},
		{
			name: "happy path",
			msg:  &testproto.Profile{},
			paths: []string{
				"user",
				"photo.path",
				"login_timestamps",
				"gallery.photo_id",
				"gallery.dimensions.width",
				"attributes.tags",
			},
		},
		{
			name: "happy path with oneof",
			msg:  &testproto.Event{},
			paths: []string{
				"user",
				"photo.photo_id",
				"photo.dimensions.height",
				"details",
				"profile.login_timestamps",
				"profile.attributes.tags",
			},
		},
		{
			name: "happy path with optional fields",
			msg:  &testproto.Options{},
			paths: []string{
				"optional_string",
				"optional_photo.photo_id",
				"optional_photo.dimensions.height",
				"optional_attr.tags",
			},
		},
		{
			name:    "incorrect root field",
			msg:     &testproto.Profile{},
			paths:   []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "incorrect nested field",
			msg:     &testproto.Profile{},
			paths:   []string{"user.invalid"},
			wantErr: true,
		},
		{
			name:    "incorrect repeated field",
			msg:     &testproto.Profile{},
			paths:   []string{"gallery.invalid"},
			wantErr: true,
		},
		{
			name:    "incorrect nested repeated field",
			msg:     &testproto.Profile{},
			paths:   []string{"gallery.dimensions.invalid"},
			wantErr: true,
		},
		{
			name:    "incorrect map field",
			msg:     &testproto.Profile{},
			paths:   []string{"attributes.invalid"},
			wantErr: true,
		},
		{
			name:    "incorrect map field",
			msg:     &testproto.Profile{},
			paths:   []string{"attributes.invalid"},
			wantErr: true,
		},
		{
			name:    "incorrect field inside oneof",
			msg:     &testproto.Event{},
			paths:   []string{"user.gallery.invalid"},
			wantErr: true,
		},
		{
			name:    "incorrect repeated field inside oneof",
			msg:     &testproto.Event{},
			paths:   []string{"profile.gallery.invalid"},
			wantErr: true,
		},
		{
			name:    "incorrect map field inside oneof",
			msg:     &testproto.Event{},
			paths:   []string{"profile.attributes.invalid"},
			wantErr: true,
		},
		{
			name:    "incorrect optional field",
			msg:     &testproto.Options{},
			paths:   []string{"optional_photo.invalid"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.msg, tt.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("want error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func BenchmarkNestedMaskFromPaths(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NestedMaskFromPaths([]string{"aaa.bbb.c.d.e.f", "aa.b.cc.ddddddd", "e", "f", "g.h.i.j.k"})
	}
}
