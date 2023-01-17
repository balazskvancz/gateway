package service

import "testing"

func TestValidateServices(t *testing.T) {
	tt := []struct {
		name  string
		srvcs services
		err   error
	}{
		{
			name:  "the functions returns err, if the services is nil",
			srvcs: nil,
			err:   errServicesIsNil,
		},
		{
			name:  "the functions returns err, if the services slice is zero length",
			srvcs: &[]Service{},
			err:   errServicesSliceIsEmpty,
		},
		{
			name: "the functions returns err, if the service has zero length prefix",
			srvcs: &[]Service{
				{
					Prefix: "",
				},
				{
					Prefix: "/foo/baz/bar",
				},
			},
			err: errServicesPrefixLength,
		},
		{
			name: "the functions returns err, if the not every service has the same prefix length",
			srvcs: &[]Service{
				{
					Prefix: "/foo/bar",
				},
				{
					Prefix: "/foo/baz/bar",
				},
			},
			err: errServicesSamePrefixLength,
		},
		{
			name: "the functions not returning err, if every service is good",
			srvcs: &[]Service{
				{
					Prefix: "/foo/bar",
				},
				{
					Prefix: "/foo/baz",
				},
			},
			err: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := ValidateServices(tc.srvcs)

			if gotErr != tc.err {
				t.Errorf("expected: %v; got: %v\n", tc.err, gotErr)
			}
		})
	}
}
