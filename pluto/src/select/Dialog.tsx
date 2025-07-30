// Copyright 2025 Synnax Labs, Inc.
//
// Use of this software is governed by the Business Source License included in the file
// licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with the Business Source
// License, use of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt.

import "@/select/Dialog.css";

import { type record, type status } from "@synnaxlabs/x";
import { memo, type ReactElement, useMemo } from "react";

import { CSS } from "@/css";
import { Dialog as CoreDialog } from "@/dialog";
import { List } from "@/list";
import { SearchInput, type SearchInputProps } from "@/select/SearchInput";
import { Status } from "@/status";

export interface DialogProps<K extends record.Key>
  extends Omit<CoreDialog.DialogProps, "children">,
    SearchInputProps,
    Pick<List.ItemsProps<K>, "emptyContent" | "children"> {
  status?: status.Status;
}

const DefaultEmptyContent = () => (
  <Status.Text.Centered variant="disabled">No results</Status.Text.Centered>
);

export const Core = memo(
  <K extends record.Key>({
    onSearch,
    children,
    emptyContent,
    searchPlaceholder,
    status,
    actions,
    ...rest
  }: DialogProps<K>) => {
    emptyContent = useMemo(() => {
      if (status != null && status.variant !== "success")
        return (
          <Status.Text.Centered
            variant={status?.variant}
            description={status?.description}
          >
            {status?.message}
          </Status.Text.Centered>
        );
      if (typeof emptyContent === "string")
        return (
          <Status.Text.Centered variant="disabled">{emptyContent}</Status.Text.Centered>
        );
      if (emptyContent == null) return <DefaultEmptyContent />;
      return emptyContent;
    }, [status?.key, emptyContent]);
    return (
      <CoreDialog.Dialog {...rest} className={CSS.BE("select", "dialog")}>
        {onSearch != null && (
          <SearchInput
            dialogVariant="floating"
            onSearch={onSearch}
            searchPlaceholder={searchPlaceholder}
            actions={actions}
          />
        )}
        <List.Items emptyContent={emptyContent} bordered borderShade={6} grow>
          {children}
        </List.Items>
      </CoreDialog.Dialog>
    );
  },
);
Core.displayName = "Select.Dialog";
export const Dialog = Core as <K extends record.Key>(
  props: DialogProps<K>,
) => ReactElement;
