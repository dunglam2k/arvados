// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import { ContextMenuActionSet } from "views-components/context-menu/context-menu-action-set";
import { RenameIcon, AdvancedIcon, RemoveIcon, AttributesIcon } from "components/icon/icon";
import { openAdvancedTabDialog } from "store/advanced-tab/advanced-tab";
import { openGroupAttributes, openRemoveGroupDialog } from "store/groups-panel/groups-panel-actions";
import { openProjectUpdateDialog } from "store/projects/project-update-actions";

export const groupActionSet: ContextMenuActionSet = [[{
    name: "Rename",
    icon: RenameIcon,
    execute: (dispatch, resource) => {
        dispatch<any>(openProjectUpdateDialog(resource));
    }
}, {
    name: "Attributes",
    icon: AttributesIcon,
    execute: (dispatch, { uuid }) => {
        dispatch<any>(openGroupAttributes(uuid));
    }
}, {
    name: "Advanced",
    icon: AdvancedIcon,
    execute: (dispatch, resource) => {
        dispatch<any>(openAdvancedTabDialog(resource.uuid));
    }
}, {
    name: "Remove",
    icon: RemoveIcon,
    execute: (dispatch, { uuid }) => {
        dispatch<any>(openRemoveGroupDialog(uuid));
    }
}]];
