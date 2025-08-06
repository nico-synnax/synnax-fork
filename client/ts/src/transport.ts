// Copyright 2025 Synnax Labs, Inc.
//
// Use of this software is governed by the Business Source License included in the file
// licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with the Business Source
// License, use of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt.

import {
  HTTPClient,
  type Middleware,
  type UnaryClient,
  unaryWithBreaker,
  WebSocketClient,
} from "@synnaxlabs/freighter";
import { type breaker } from "@synnaxlabs/x";
import { binary } from "@synnaxlabs/x/binary";
import { type URL } from "@synnaxlabs/x/url";

export class Transport {
  readonly url: URL;
  readonly unary: UnaryClient;
  private readonly http: HTTPClient;
  readonly stream: WebSocketClient;
  readonly secure: boolean;
  private readonly breakerCfg: breaker.Config;

  constructor(url: URL, breakerCfg: breaker.Config = {}, secure: boolean = false) {
    this.secure = secure;
    this.url = url.child("/api/v1");
    this.breakerCfg = breakerCfg;
    const codec = new binary.JSONCodec();
    this.http = new HTTPClient(this.url, codec, this.secure);
    this.unary = unaryWithBreaker(this.http, this.breakerCfg);
    this.stream = new WebSocketClient(this.url, codec, this.secure);
  }

  withDecoder(decoder: binary.Codec): UnaryClient {
    const c = this.http.withDecoder(decoder);
    return unaryWithBreaker(c, this.breakerCfg);
  }

  use(...middleware: Middleware[]): void {
    this.unary.use(...middleware);
    this.stream.use(...middleware);
  }
}
