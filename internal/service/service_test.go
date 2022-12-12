package service

import "testing"

func TestValidateServices(t *testing.T) {
	tt := []struct {
		name  string
		srvcs *services
		err   error
	}{
		{
			name:  "the functions returns err, if the services is nil",
			srvcs: nil,
			err:   errServicesIsNil,
		},
		{
			name:  "the functions returns err, if the services slice is zero length",
			srvcs: &services{},
			err:   errServicesSliceIsEmpty,
		},
		{
			name: "the functions returns err, if the not every service has the same prefix length",
			srvcs: &services{
				Services: []Service{
					{
						Prefix: "/foo/bar",
					},
					{
						Prefix: "/foo/baz/bar",
					},
				},
			},
			err: errservicesPrefixLength,
		},
		{
			name: "the functions not returning err, if every service is good",
			srvcs: &services{
				Services: []Service{
					{
						Prefix: "/foo/bar",
					},
					{
						Prefix: "/foo/baz",
					},
				},
			},
			err: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := validateServices(tc.srvcs)

			if gotErr != tc.err {
				t.Errorf("expected: %v; got: %v\n", tc.err, gotErr)
			}

		})

	}
}
