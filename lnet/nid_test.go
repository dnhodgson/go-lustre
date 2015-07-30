package lnet_test

import (
	"encoding/json"
	"testing"

	"github.intel.com/hpdd/ce-tools/lib/tu"
	"github.intel.com/hpdd/ce-tools/resources/lustre/lnet"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNidFunctions(t *testing.T) {
	Convey("NidFromString() should attempt to parse a string into a Nid", t, func() {
		var tests = []struct {
			in  string
			out string
			err string
		}{
			{
				in:  `127.0.0.1@tcp`,
				out: `127.0.0.1@tcp0`,
			},
			{
				in:  `127.0.0.2@tcp42`,
				out: `127.0.0.2@tcp42`,
			},
			{
				in:  `10.0.1.10@o2ib42`,
				out: `10.0.1.10@o2ib42`,
			},
			{
				in:  `101@gni`,
				err: `Unsupported LND: gni`,
			},
			{
				in:  `101`,
				err: `Cannot parse NID from "101"`,
			},
			{
				in:  `@tcp`,
				err: `"" is not a valid IP address`,
			},
		}

		for _, tc := range tests {
			Convey(tc.in, func() {
				n, err := lnet.NidFromString(tc.in)
				So(tu.Err2str(err), ShouldEqual, tc.err)

				if n != nil {
					So(n.String(), ShouldEqual, tc.out)
				}
			})
		}
	})

	Convey("SupportedDrivers() should return a list of driver names", t, func() {
		So(lnet.SupportedDrivers(), ShouldNotBeEmpty)
	})
}

func TestMarshalNid(t *testing.T) {
	var tests = []struct {
		in  string
		out string
		err string
	}{
		{
			in:  `127.0.0.1@tcp0`,
			out: `"127.0.0.1@tcp0"`,
		},
		{
			in:  `127.0.0.2@tcp42`,
			out: `"127.0.0.2@tcp42"`,
		},
		{
			in:  `10.0.1.10@o2ib42`,
			out: `"10.0.1.10@o2ib42"`,
		},
	}

	Convey("Marshalling to JSON should return string ", t, func() {
		for _, tc := range tests {
			Convey(tc.in, func() {
				n, err := lnet.NidFromString(tc.in)
				if err != nil {
					t.Fatal(err)
				}
				j, err := json.Marshal(n)
				So(tu.Err2str(err), ShouldEqual, tc.err)

				if j != nil {
					So(string(j), ShouldEqual, tc.out)
				}
			})
		}
	})

	Convey("Unmarshalling from JSON should return nid ", t, func() {
		for _, tc := range tests {
			Convey(tc.out, func() {
				var nid lnet.Nid
				err := json.Unmarshal([]byte(tc.out), &nid)
				So(tu.Err2str(err), ShouldEqual, tc.err)

				if err == nil {
					So(nid.String(), ShouldEqual, tc.in)
				}
			})
		}
	})
}