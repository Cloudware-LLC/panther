package pantherlog

/**
 * Panther is a Cloud-Native SIEM for the Modern Security Team.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"encoding/json"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"

	"github.com/panther-labs/panther/internal/log_analysis/log_processor/pantherlog/null"
	"github.com/panther-labs/panther/pkg/box"
)

type testStringer struct {
	Foo string
}

func (t *testStringer) String() string {
	return t.Foo
}
func (t *testStringer) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Foo)
}

var (
	// Register our own random value kinds
	kindFoo  = ValueKind(time.Now().UnixNano())
	kindBar  = kindFoo + 1
	kindBaz  = kindFoo + 2
	kindQux  = kindFoo + 3
	kindQuux = kindFoo + 4
)

func init() {
	MustRegisterScanner("foo", kindFoo, kindFoo)
	MustRegisterScanner("bar", kindBar, kindBar)
	MustRegisterScanner("baz", kindBaz, kindBaz)
	MustRegisterScanner("qux", kindQux, kindQux)
	MustRegisterScanner("quux", kindQuux, kindQuux)
}

func TestScanValueEncodersExt(t *testing.T) {
	ext := scanValueEncodersExt{}
	api := jsoniter.Config{}.Froze()
	api.RegisterExtension(&ext)

	// Check all possible string types
	type T struct {
		Foo  *testStringer `json:"foo" panther:"foo"`
		Bar  testStringer  `json:"bar" panther:"bar"`
		Baz  string        `json:"baz" panther:"baz"`
		Qux  *string       `json:"qux" panther:"qux"`
		Quux null.String   `json:"quux" panther:"quux"`
	}

	v := T{
		Foo: &testStringer{
			Foo: "ok",
		},
		Bar: testStringer{
			Foo: "ok",
		},
		Baz:  "ok",
		Qux:  box.String("ok"),
		Quux: null.FromString("ok"),
	}

	values := ValueBuffer{}
	stream := api.BorrowStream(nil)
	stream.Attachment = &values
	stream.WriteVal(&v)
	require.Equal(t, []string{"ok"}, values.Values(kindFoo))
	require.Equal(t, []string{"ok"}, values.Values(kindBar))
	require.Equal(t, []string{"ok"}, values.Values(kindBaz))
	require.Equal(t, []string{"ok"}, values.Values(kindQux))
	require.Equal(t, []string{"ok"}, values.Values(kindQuux))
	actual := string(stream.Buffer())
	require.Equal(t, `{"foo":"ok","bar":"ok","baz":"ok","qux":"ok","quux":"ok"}`, actual)
}