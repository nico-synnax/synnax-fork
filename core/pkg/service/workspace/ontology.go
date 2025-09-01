// Copyright 2025 Synnax Labs, Inc.
//
// Use of this software is governed by the Business Source License included in the file
// licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with the Business Source
// License, use of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt.

package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/synnaxlabs/synnax/pkg/distribution/ontology"
	xchange "github.com/synnaxlabs/x/change"
	"github.com/synnaxlabs/x/gorp"
	"github.com/synnaxlabs/x/iter"
	"github.com/synnaxlabs/x/observe"
	"github.com/synnaxlabs/x/zyn"
)

const OntologyType ontology.Type = "workspace"

func OntologyID(k uuid.UUID) ontology.ID {
	return ontology.ID{Type: OntologyType, Key: k.String()}
}

func OntologyIDs(keys []uuid.UUID) (ids []ontology.ID) {
	return lo.Map(keys, func(k uuid.UUID, _ int) ontology.ID { return OntologyID(k) })
}

func OntologyIDsFromWorkspaces(workspaces []Workspace) (ids []ontology.ID) {
	return lo.Map(workspaces, func(w Workspace, _ int) ontology.ID {
		return OntologyID(w.Key)
	})
}

func KeysFromOntologyIDs(ids []ontology.ID) ([]uuid.UUID, error) {
	keys := make([]uuid.UUID, len(ids))
	var err error
	for i, id := range ids {
		keys[i], err = uuid.Parse(id.Key)
		if err != nil {
			return nil, err
		}
	}
	return keys, nil
}

var schema = zyn.Object(map[string]zyn.Schema{
	"key":  zyn.UUID(),
	"name": zyn.String(),
})

func newResource(ws Workspace) ontology.Resource {
	return ontology.NewResource(schema, OntologyID(ws.Key), ws.Name, ws)
}

type change = xchange.Change[uuid.UUID, Workspace]

func (s *Service) Type() ontology.Type { return OntologyType }

// Schema implements ontology.Service.
func (s *Service) Schema() zyn.Schema { return schema }

// RetrieveResource implements ontology.Service.
func (s *Service) RetrieveResource(
	ctx context.Context,
	key string,
	tx gorp.Tx,
) (ontology.Resource, error) {
	k := uuid.MustParse(key)
	var ws Workspace
	if err := s.NewRetrieve().WhereKeys(k).Entry(&ws).Exec(ctx, tx); err != nil {
		return ontology.Resource{}, err
	}
	return newResource(ws), nil
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
	handleChange := func(
		ctx context.Context,
		reader gorp.TxReader[uuid.UUID, Workspace],
	) {
		f(
			ctx,
			iter.NexterTranslator[change, ontology.Change]{
				Wrap:      reader,
				Translate: translateChange,
			},
		)
	}
	return gorp.Observe[uuid.UUID, Workspace](s.DB).OnChange(handleChange)
}

// OpenNexter implements ontology.Service.
func (s *Service) OpenNexter() (iter.NexterCloser[ontology.Resource], error) {
	n, err := gorp.WrapReader[uuid.UUID, Workspace](s.DB).OpenNexter()
	if err != nil {
		return nil, err
	}
	return iter.NexterCloserTranslator[Workspace, ontology.Resource]{
		Wrap:      n,
		Translate: newResource,
	}, nil
}
