// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import { ContextMenuActionSet } from "../context-menu-action-set";
import { RenameIcon, DownloadIcon, RemoveIcon } from "../../../components/icon/icon";
import { openRemoveDialog } from "../../remove-dialog/remove-dialog";
import { openRenameFileDialog } from "../../rename-file-dialog/rename-file-dialog";


export const collectionFilesItemActionSet: ContextMenuActionSet = [[{
    name: "Rename",
    icon: RenameIcon,
    execute: (dispatch, resource) => {
        dispatch(openRenameFileDialog(resource.name));
    }
},{
    name: "Download",
    icon: DownloadIcon,
    execute: (dispatch, resource) => {
        return;
    }
},{
    name: "Remove",
    icon: RemoveIcon,
    execute: (dispatch, resource) => {
        dispatch(openRemoveDialog('selected file'));
    }
}]];
