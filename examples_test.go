package fmutils_test

import (
	"fmt"
	"regexp"

	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/proto"

	"github.com/mennanov/fmutils"
	"github.com/mennanov/fmutils/testproto"
)

var reSpaces = regexp.MustCompile(`\s+`)

// ExampleFilter_update_request illustrates an API endpoint that updates an existing entity.
// The request to that endpoint provides a field mask that should be used to update the entity.
func ExampleFilter_update_request() {
	// Assuming the profile entity is loaded from a database.
	profile := &testproto.Profile{
		User: &testproto.User{
			UserId: 64,
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
		LoginTimestamps: []int64{1, 2, 3},
	}
	// An API request from an API user.
	updateProfileRequest := &testproto.UpdateProfileRequest{
		Profile: &testproto.Profile{
			User: &testproto.User{
				UserId: 65, // not listed in the field mask, so won't be updated.
				Name:   "new user name",
			},
			Photo: &testproto.Photo{
				PhotoId: 3, // not listed in the field mask, so won't be updated.
				Path:    "new photo path",
				Dimensions: &testproto.Dimensions{
					Width: 50,
				},
			},
			LoginTimestamps: []int64{4, 5}},
		Fieldmask: &field_mask.FieldMask{
			Paths: []string{"user.name", "photo.path", "photo.dimensions.width", "login_timestamps"}},
	}
	// Normalize and validate the field mask before using it.
	updateProfileRequest.Fieldmask.Normalize()
	if !updateProfileRequest.Fieldmask.IsValid(profile) {
		// Return an error.
		panic("invalid field mask")
	}
	// Redact the request according to the provided field mask.
	fmutils.Filter(updateProfileRequest.GetProfile(), updateProfileRequest.Fieldmask.GetPaths())
	// Now that the request is vetted we can merge it with the profile entity.
	proto.Merge(profile, updateProfileRequest.GetProfile())
	// The profile can now be saved in a database.
	fmt.Println(reSpaces.ReplaceAllString(profile.String(), " "))
	// Output: user:{user_id:64 name:"new user name"} photo:{photo_id:2 path:"new photo path" dimensions:{width:50 height:120}} login_timestamps:1 login_timestamps:2 login_timestamps:3 login_timestamps:4 login_timestamps:5
}

// ExampleFilter_reuse_mask illustrates how a single NestedMask instance can be used to process multiple proto messages.
func ExampleFilter_reuse_mask() {
	users := []*testproto.User{
		{
			UserId: 1,
			Name:   "name 1",
		},
		{
			UserId: 2,
			Name:   "name 2",
		},
	}
	// Create a mask only once and reuse it.
	mask := fmutils.NestedMaskFromPaths([]string{"name"})
	for _, user := range users {
		mask.Filter(user)
	}
	fmt.Println(users)
	// Output: [name:"name 1" name:"name 2"]
}

// ExamplePathsFromFieldNumbers illustrates how to convert protobuf field numbers to field paths.
// This is useful when you have field numbers from the protobuf schema and need to convert them
// to field paths for use with field masks.
func ExamplePathsFromFieldNumbers() {
	user := &testproto.User{}

	// Convert field numbers to field paths.
	// Field 1 is "user_id", field 2 is "name"
	paths := fmutils.PathsFromFieldNumbers(user, 1, 2)
	fmt.Println("Field numbers:", []int{1, 2})
	fmt.Println("Paths:", paths)

	// Output:
	// Field numbers: [1 2]
	// Paths: [user_id name]
}
