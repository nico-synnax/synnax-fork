// Copyright 2025 Synnax Labs, Inc.
//
// Use of this software is governed by the Business Source License included in the file
// licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with the Business Source
// License, use of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt.

import { type PageNavNode } from "@/components/PageNav/PageNav";

export const cppClientNav: PageNavNode = {
  key: "cpp-client",
  name: "C++ Client",
  children: [
    {
      key: "/reference/cpp-client/get-started",
      href: "/reference/cpp-client/get-started",
      name: "Get Started",
    },
    {
      key: "/reference/cpp-client/channels",
      href: "/reference/cpp-client/channels",
      name: "Channels",
    },
    {
      key: "/reference/cpp-client/read-data",
      href: "/reference/cpp-client/read-data",
      name: "Read Data",
    },
    {
      key: "/reference/cpp-client/write-data",
      href: "/reference/cpp-client/write-data",
      name: "Write Data",
    },
    {
      key: "/reference/cpp-client/stream-data",
      href: "/reference/cpp-client/stream-data",
      name: "Stream Data",
    },
  ],
};
