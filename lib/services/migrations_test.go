/*
Copyright 2015 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package services

import (
	"encoding/json"
	"time"

	"github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/utils"

	"github.com/davecgh/go-spew/spew"
	"github.com/kylelemons/godebug/diff"
	. "gopkg.in/check.v1"
)

type MigrationsSuite struct {
}

var _ = Suite(&MigrationsSuite{})

func (s *MigrationsSuite) SetUpSuite(c *C) {
	utils.InitLoggerForTests()
}

func (s *MigrationsSuite) TestMigrateServers(c *C) {
	in := &ServerV0{
		Kind:      KindNode,
		ID:        "id1",
		Addr:      "127.0.0.1:22",
		Hostname:  "localhost",
		Labels:    map[string]string{"a": "b", "c": "d"},
		CmdLabels: map[string]CommandLabelV0{"o": CommandLabelV0{Command: []string{"ls", "-l"}, Period: time.Second}},
	}

	out := in.V1()
	c.Assert(out, DeepEquals, &ServerV1{
		Kind:    KindNode,
		Version: V1,
		Metadata: Metadata{
			Name:      in.ID,
			Labels:    in.Labels,
			Namespace: defaults.Namespace,
		},
		Spec: ServerSpecV1{
			Addr:      in.Addr,
			Hostname:  in.Hostname,
			CmdLabels: map[string]CommandLabelV1{"o": CommandLabelV1{Command: []string{"ls", "-l"}, Period: NewDuration(time.Second)}},
		},
	})

	data, err := json.Marshal(in)
	c.Assert(err, IsNil)
	out2, err := GetServerMarshaler().UnmarshalServer(data, KindNode)
	c.Assert(err, IsNil)
	c.Assert(out2, DeepEquals, out)
}

func (s *MigrationsSuite) TestMigrateUsers(c *C) {
	in := &UserV0{
		Name:           "alice",
		AllowedLogins:  []string{"admin", "centos"},
		OIDCIdentities: []OIDCIdentity{{Email: "alice@example.com", ConnectorID: "example"}},
		Status: LoginStatus{
			IsLocked:    true,
			LockedTime:  time.Date(2015, 12, 10, 1, 1, 3, 0, time.UTC),
			LockExpires: time.Date(2015, 12, 10, 1, 2, 3, 0, time.UTC),
		},
		Expires: time.Date(2016, 12, 10, 1, 2, 3, 0, time.UTC),
		CreatedBy: CreatedBy{
			Time:      time.Date(2013, 12, 10, 1, 2, 3, 0, time.UTC),
			Connector: &ConnectorRef{ID: "example"},
		},
	}

	out := in.V1()
	expected := &UserV1{
		Kind:    KindUser,
		Version: V1,
		Metadata: Metadata{
			Name:      in.Name,
			Namespace: defaults.Namespace,
		},
		Spec: UserSpecV1{
			OIDCIdentities: in.OIDCIdentities,
			Status:         in.Status,
			Expires:        in.Expires,
			CreatedBy:      in.CreatedBy,
		},
		rawObject: *in,
	}
	c.Assert(out.rawObject, DeepEquals, *in)
	c.Assert(out.Metadata, DeepEquals, expected.Metadata)
	c.Assert(out.Spec, DeepEquals, expected.Spec)
	c.Assert(out, DeepEquals, expected)

	data, err := json.Marshal(in)
	c.Assert(err, IsNil)
	out2, err := GetUserMarshaler().UnmarshalUser(data)
	c.Assert(err, IsNil)
	c.Assert(out2.GetRawObject(), DeepEquals, *in)
	c.Assert(out2, DeepEquals, expected)

	data, err = json.Marshal(expected)
	c.Assert(err, IsNil)
	out3, err := GetUserMarshaler().UnmarshalUser(data)
	c.Assert(err, IsNil)

	d := &spew.ConfigState{Indent: " ", DisableMethods: true}
	c.Assert(out3, DeepEquals, expected, Commentf("%v", diff.Diff(d.Sdump(out3), d.Sdump(expected))))
}

func (s *MigrationsSuite) TestMigrateReverseTunnels(c *C) {
	in := &ReverseTunnelV0{
		DomainName: "example.com",
		DialAddrs:  []string{"127.0.0.1:3245", "127.0.0.1:3450"},
	}

	out := in.V1()
	expected := &ReverseTunnelV1{
		Kind:    KindReverseTunnel,
		Version: V1,
		Metadata: Metadata{
			Name:      in.DomainName,
			Namespace: defaults.Namespace,
		},
		Spec: ReverseTunnelSpecV1{
			DialAddrs:   in.DialAddrs,
			ClusterName: in.DomainName,
		},
	}
	c.Assert(out, DeepEquals, expected)

	data, err := json.Marshal(in)
	c.Assert(err, IsNil)
	out2, err := GetReverseTunnelMarshaler().UnmarshalReverseTunnel(data)
	c.Assert(err, IsNil)
	c.Assert(out2, DeepEquals, expected)

	data, err = json.Marshal(expected)
	c.Assert(err, IsNil)
	out3, err := GetReverseTunnelMarshaler().UnmarshalReverseTunnel(data)
	c.Assert(err, IsNil)
	c.Assert(out3, DeepEquals, expected)
}

func (s *MigrationsSuite) TestMigrateCertAuthorities(c *C) {
	in := &CertAuthorityV0{
		Type:          UserCA,
		DomainName:    "example.com",
		CheckingKeys:  [][]byte{[]byte("checking key")},
		SigningKeys:   [][]byte{[]byte("signing key")},
		AllowedLogins: []string{"root", "admin"},
	}

	out := in.V1()
	expected := &CertAuthorityV1{
		Kind:    KindCertAuthority,
		Version: V1,
		Metadata: Metadata{
			Name:      in.DomainName,
			Namespace: defaults.Namespace,
		},
		Spec: CertAuthoritySpecV1{
			ClusterName:  in.DomainName,
			Type:         in.Type,
			CheckingKeys: in.CheckingKeys,
			SigningKeys:  in.SigningKeys,
		},
		rawObject: *in,
	}
	c.Assert(out, DeepEquals, expected)

	data, err := json.Marshal(in)
	c.Assert(err, IsNil)
	out2, err := GetCertAuthorityMarshaler().UnmarshalCertAuthority(data)
	c.Assert(err, IsNil)
	c.Assert(out2, DeepEquals, expected)
}