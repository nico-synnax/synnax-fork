// Copyright 2025 Synnax Labs, Inc.
//
// Use of this software is governed by the Business Source License included in the file
// licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with the Business Source
// License, use of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt.

package device

import (
	"context"

	"github.com/samber/lo"
	"github.com/synnaxlabs/synnax/pkg/distribution/ontology"
	xchange "github.com/synnaxlabs/x/change"
	"github.com/synnaxlabs/x/gorp"
	"github.com/synnaxlabs/x/iter"
	"github.com/synnaxlabs/x/observe"
	"github.com/synnaxlabs/x/zyn"
)

const OntologyType ontology.Type = "device"

func OntologyID(k string) ontology.ID { return ontology.ID{Type: OntologyType, Key: k} }

func OntologyIDsFromDevices(devices []Device) []ontology.ID {
	return lo.Map(devices, func(d Device, _ int) ontology.ID {
		return OntologyID(d.Key)
	})
}

func OntologyIDs(keys []string) []ontology.ID {
	return lo.Map(keys, func(key string, _ int) ontology.ID { return OntologyID(key) })
}

func KeysFromOntologyIDs(ids []ontology.ID) []string {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = id.Key
	}
	return keys
}

var schema = zyn.Object(map[string]zyn.Schema{
	"key":        zyn.String(),
	"name":       zyn.String(),
	"make":       zyn.String(),
	"model":      zyn.String(),
	"configured": zyn.Bool(),
	"location":   zyn.String(),
	"rack":       zyn.Uint32().Coerce(),
})

func newResource(d Device) ontology.Resource {
	return ontology.NewResource(schema, OntologyID(d.Key), d.Name, d)
}

var _ ontology.Service = (*Service)(nil)

type change = xchange.Change[string, Device]

func (s *Service) Type() ontology.Type { return OntologyType }

// Schema implements ontology.Service.
func (s *Service) Schema() zyn.Schema { return schema }

// RetrieveResource implements ontology.Service.
func (s *Service) RetrieveResource(
	ctx context.Context,
	key string,
	tx gorp.Tx,
) (ontology.Resource, error) {
	var d Device
	if err := s.NewRetrieve().WhereKeys(key).Entry(&d).Exec(ctx, tx); err != nil {
		return ontology.Resource{}, err
	}
	return newResource(d), nil
}

func translateChange(c change) ontology.Change {
	return ontology.Change{
		Variant: c.Variant,
		Key:     OntologyID(c.Key),
		Value:   newResource(c.Value),
	}
}

// OnChange implements ontology.Service.
func (s *Service) OnChange(
	f func(context.Context, iter.Nexter[ontology.Change]),
) observe.Disconnect {
	handleChange := func(ctx context.Context, reader gorp.TxReader[string, Device]) {
		f(
			ctx,
			iter.NexterTranslator[change, ontology.Change]{
				Wrap:      reader,
				Translate: translateChange,
			},
		)
	}
	return gorp.Observe[string, Device](s.DB).OnChange(handleChange)
}

// OpenNexter implements ontology.Service.
func (s *Service) OpenNexter() (iter.NexterCloser[ontology.Resource], error) {
	n, err := gorp.WrapReader[string, Device](s.DB).OpenNexter()
	if err != nil {
		return nil, err
	}
	return iter.NexterCloserTranslator[Device, ontology.Resource]{
		Wrap:      n,
		Translate: newResource,
	}, nil
}
