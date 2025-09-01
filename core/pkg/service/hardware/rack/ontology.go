// Copyright 2025 Synnax Labs, Inc.
//
// Use of this software is governed by the Business Source License included in the file
// licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with the Business Source
// License, use of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt.

package rack

import (
	"context"
	"strconv"

	"github.com/samber/lo"
	"github.com/synnaxlabs/synnax/pkg/distribution/ontology"
	xchange "github.com/synnaxlabs/x/change"
	"github.com/synnaxlabs/x/gorp"
	"github.com/synnaxlabs/x/iter"
	"github.com/synnaxlabs/x/observe"
	"github.com/synnaxlabs/x/zyn"
)

const OntologyType ontology.Type = "rack"

func OntologyID(k Key) ontology.ID {
	return ontology.ID{Type: OntologyType, Key: k.String()}
}

func OntologyIDs(keys []Key) []ontology.ID {
	return lo.Map(keys, func(k Key, _ int) ontology.ID { return OntologyID(k) })
}

func OntologyIDsFromRacks(racks []Rack) []ontology.ID {
	return lo.Map(racks, func(r Rack, _ int) ontology.ID { return OntologyID(r.Key) })
}

func KeysFromOntologyIDs(ids []ontology.ID) ([]Key, error) {
	keys := make([]Key, len(ids))
	for i, id := range ids {
		k, err := strconv.Atoi(id.Key)
		if err != nil {
			return nil, err
		}
		keys[i] = Key(k)
	}
	return keys, nil
}

var schema = zyn.Object(map[string]zyn.Schema{
	"key":  zyn.Uint32().Coerce(),
	"name": zyn.String(),
})

func newResource(r Rack) ontology.Resource {
	return ontology.NewResource(schema, OntologyID(r.Key), r.Name, r)
}

type change = xchange.Change[Key, Rack]

func (s *Service) Type() ontology.Type { return OntologyType }

// Schema implements ontology.Service.
func (s *Service) Schema() zyn.Schema { return schema }

// RetrieveResource implements ontology.Service.
func (s *Service) RetrieveResource(
	ctx context.Context,
	key string,
	tx gorp.Tx,
) (ontology.Resource, error) {
	k, err := strconv.Atoi(key)
	if err != nil {
		return ontology.Resource{}, err
	}
	var r Rack
	if err = s.NewRetrieve().WhereKeys(Key(k)).Entry(&r).Exec(ctx, tx); err != nil {
		return ontology.Resource{}, err
	}
	return newResource(r), nil
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
	handleChange := func(ctx context.Context, reader gorp.TxReader[Key, Rack]) {
		f(
			ctx,
			iter.NexterTranslator[change, ontology.Change]{
				Wrap:      reader,
				Translate: translateChange,
			},
		)
	}
	return gorp.Observe[Key, Rack](s.DB).OnChange(handleChange)
}

// OpenNexter implements ontology.Service.
func (s *Service) OpenNexter() (iter.NexterCloser[ontology.Resource], error) {
	n, err := gorp.WrapReader[Key, Rack](s.DB).OpenNexter()
	if err != nil {
		return nil, err
	}
	return iter.NexterCloserTranslator[Rack, ontology.Resource]{
		Wrap:      n,
		Translate: newResource,
	}, nil
}
