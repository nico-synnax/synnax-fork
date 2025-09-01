// Copyright 2025 Synnax Labs, Inc.
//
// Use of this software is governed by the Business Source License included in the file
// licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with the Business Source
// License, use of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt.

package task

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

const OntologyType ontology.Type = "task"

func OntologyID(k Key) ontology.ID {
	return ontology.ID{Type: OntologyType, Key: k.String()}
}

func OntologyIDs(keys []Key) []ontology.ID {
	return lo.Map(keys, func(item Key, _ int) ontology.ID { return OntologyID(item) })
}

func OntologyIDsFromTasks(tasks []Task) []ontology.ID {
	return lo.Map(tasks, func(t Task, _ int) ontology.ID { return OntologyID(t.Key) })
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
	"key":      zyn.Uint64().Coerce(),
	"name":     zyn.String(),
	"type":     zyn.String(),
	"snapshot": zyn.Bool(),
})

func newResource(t Task) ontology.Resource {
	return ontology.NewResource(schema, OntologyID(t.Key), t.Name, t)
}

type change = xchange.Change[Key, Task]

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
	var t Task
	if err = s.NewRetrieve().WhereKeys(Key(k)).Entry(&t).Exec(ctx, tx); err != nil {
		return ontology.Resource{}, err
	}
	return newResource(t), nil
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
	handleChange := func(ctx context.Context, reader gorp.TxReader[Key, Task]) {
		f(
			ctx,
			iter.NexterTranslator[change, ontology.Change]{
				Wrap:      reader,
				Translate: translateChange,
			},
		)
	}
	return gorp.Observe[Key, Task](s.cfg.DB).OnChange(handleChange)
}

// OpenNexter implements ontology.Service.
func (s *Service) OpenNexter() (iter.NexterCloser[ontology.Resource], error) {
	n, err := gorp.WrapReader[Key, Task](s.cfg.DB).OpenNexter()
	if err != nil {
		return nil, err
	}
	return iter.NexterCloserTranslator[Task, ontology.Resource]{
		Wrap:      n,
		Translate: newResource,
	}, nil
}
