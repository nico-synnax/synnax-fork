// Copyright 2025 Synnax Labs, Inc.
//
// Use of this software is governed by the Business Source License included in the file
// licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with the Business Source
// License, use of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt.

package lineplot

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

const ontologyType ontology.Type = "lineplot"

// OntologyID returns unique identifier for the line plot within the ontology.
func OntologyID(key uuid.UUID) ontology.ID {
	return ontology.ID{Type: ontologyType, Key: key.String()}
}

// OntologyIDs returns unique identifiers for the line plots within the ontology.
func OntologyIDs(keys []uuid.UUID) []ontology.ID {
	return lo.Map(keys, func(id uuid.UUID, _ int) ontology.ID { return OntologyID(id) })
}

// OntologyIDsFromLinePlots returns the ontology IDs of the line plots.
func OntologyIDsFromLinePlots(lineplots []LinePlot) []ontology.ID {
	return lo.Map(lineplots, func(lp LinePlot, _ int) ontology.ID {
		return OntologyID(lp.Key)
	})
}

var schema = zyn.Object(map[string]zyn.Schema{"key": zyn.UUID(), "name": zyn.String()})

func newResource(lineplot LinePlot) ontology.Resource {
	return ontology.NewResource(
		schema,
		OntologyID(lineplot.Key),
		lineplot.Name,
		lineplot,
	)
}

var _ ontology.Service = (*Service)(nil)

type change = xchange.Change[uuid.UUID, LinePlot]

func (s *Service) Type() ontology.Type { return ontologyType }

// Schema implements ontology.Service.
func (s *Service) Schema() zyn.Schema { return schema }

// RetrieveResource implements ontology.Service.
func (s *Service) RetrieveResource(
	ctx context.Context,
	key string,
	tx gorp.Tx,
) (ontology.Resource, error) {
	k := uuid.MustParse(key)
	var lp LinePlot
	if err := s.NewRetrieve().WhereKeys(k).Entry(&lp).Exec(ctx, tx); err != nil {
		return ontology.Resource{}, err
	}
	return newResource(lp), nil
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
		reader gorp.TxReader[uuid.UUID, LinePlot],
	) {
		f(
			ctx,
			iter.NexterTranslator[change, ontology.Change]{
				Wrap:      reader,
				Translate: translateChange,
			},
		)
	}
	return gorp.Observe[uuid.UUID, LinePlot](s.DB).OnChange(handleChange)
}

// OpenNexter implements ontology.Service.
func (s *Service) OpenNexter() (iter.NexterCloser[ontology.Resource], error) {
	n, err := gorp.WrapReader[uuid.UUID, LinePlot](s.DB).OpenNexter()
	if err != nil {
		return nil, err
	}
	return iter.NexterCloserTranslator[LinePlot, ontology.Resource]{
		Wrap:      n,
		Translate: newResource,
	}, nil
}
