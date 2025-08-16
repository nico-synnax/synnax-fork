// Copyright 2025 Synnax Labs, Inc.
//
// Use of this software is governed by the Business Source License included in the file
// licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with the Business Source
// License, use of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt.

import { framer, type Synnax } from "@synnaxlabs/client";
import { type AsyncDestructor, DataType, unique } from "@synnaxlabs/x";
import type z from "zod";

import { type Store, type StoreConfig } from "@/flux/aether/store";
import { type Status } from "@/status";

/**
 * Sorts channel names to ensure deletions are processed before other changes.
 * This ensures that modifications to things like relationships (delete followed by create)
 * are processed in the correct order.
 *
 * @param a - First channel name
 * @param b - Second channel name
 * @returns Sort order (-1, 0, or 1)
 */
const channelNameSort = (a: string, b: string) => {
  const aHasDelete = a.includes("delete");
  const bHasDelete = b.includes("delete");
  if (aHasDelete && !bHasDelete) return -1;
  if (!aHasDelete && bHasDelete) return 1;
  return 0;
};

/**
 * Arguments for opening a flux streamer.
 *
 * @template ScopedStore - The type of the store
 */
export interface StreamerArgs<ScopedStore extends Store> {
  /** Function to handle errors that occur during streaming */
  handleError: Status.ErrorHandler;
  /** Configuration defining store structure and listeners */
  storeConfig: StoreConfig<ScopedStore>;
  /** Synnax client instance for API access */
  client: Synnax;
  /** Function to open a frame streamer */
  openStreamer: framer.StreamOpener;
  /** The store instance to update with streamed data */
  store: ScopedStore;
}

/**
 * Opens a hardened streamer that listens to configured channels and invokes
 * the appropriate listeners when data changes.
 *
 * @template ScopedStore - The type of the store
 * @param args - Configuration for the streamer
 * @returns A destructor function to close the streamer
 */
export const openStreamer = async <ScopedStore extends Store>({
  openStreamer: streamOpener,
  storeConfig,
  handleError,
  client,
  store,
}: StreamerArgs<ScopedStore>): Promise<AsyncDestructor> => {
  const configValues = Object.values(storeConfig);
  const channels = unique.unique(
    configValues.flatMap(({ listeners }) => listeners.map(({ channel }) => channel)),
  );
  const listenersForChannel = (name: string) =>
    configValues.flatMap(({ listeners }) =>
      listeners.filter(({ channel }) => channel === name),
    );
  const hardenedStreamer = await framer.HardenedStreamer.open(streamOpener, channels);
  const observableStreamer = new framer.ObservableStreamer(hardenedStreamer);
  const handleChange = (frame: framer.Frame) => {
    const namesInFrame = [...frame.uniqueNames];
    namesInFrame.sort(channelNameSort);
    namesInFrame.forEach((name) => {
      const series = frame.get(name);
      listenersForChannel(name).forEach(({ onChange, schema }) => {
        handleError(async () => {
          let parsed: z.output<typeof schema>[];
          if (!series.dataType.equals(DataType.JSON))
            parsed = Array.from(series).map((s) => schema.parse(s));
          else parsed = series.parseJSON(schema);
          for (const changed of parsed) await onChange({ changed, client, store });
        }, "Failed to handle streamer change");
      });
    });
  };
  observableStreamer.onChange(handleChange);
  return observableStreamer.close.bind(observableStreamer);
};
