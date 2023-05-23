// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import { ContextMenuAction, ContextMenuActionSet } from "views-components/context-menu/context-menu-action-set";
import { collectionPanelFilesAction, openMultipleFilesRemoveDialog } from "store/collection-panel/collection-panel-files/collection-panel-files-actions";
import {
    openCollectionPartialCopyMultipleToNewCollectionDialog,
    openCollectionPartialCopyMultipleToExistingCollectionDialog,
    openCollectionPartialCopyToSeparateCollectionsDialog
} from 'store/collections/collection-partial-copy-actions';
import { openCollectionPartialMoveMultipleToExistingCollectionDialog, openCollectionPartialMoveMultipleToNewCollectionDialog, openCollectionPartialMoveToSeparateCollectionsDialog } from "store/collections/collection-partial-move-actions";

const copyActions: ContextMenuAction[] = [
    {
        name: "Copy selected into new collection",
        execute: dispatch => {
            dispatch<any>(openCollectionPartialCopyMultipleToNewCollectionDialog());
        }
    },
    {
        name: "Copy selected into existing collection",
        execute: dispatch => {
            dispatch<any>(openCollectionPartialCopyMultipleToExistingCollectionDialog());
        }
    },
];

const copyActionsMultiple: ContextMenuAction[] = [
    ...copyActions,
    {
        name: "Copy selected into separate collections",
        execute: dispatch => {
            dispatch<any>(openCollectionPartialCopyToSeparateCollectionsDialog());
        }
    }
];

const moveActions: ContextMenuAction[] = [
    {
        name: "Move selected into new collection",
        execute: dispatch => {
            dispatch<any>(openCollectionPartialMoveMultipleToNewCollectionDialog());
        }
    },
    {
        name: "Move selected into existing collection",
        execute: dispatch => {
            dispatch<any>(openCollectionPartialMoveMultipleToExistingCollectionDialog());
        }
    },
];

const moveActionsMultiple: ContextMenuAction[] = [
    ...moveActions,
    {
        name: "Move selected into separate collections",
        execute: dispatch => {
            dispatch<any>(openCollectionPartialMoveToSeparateCollectionsDialog());
        }
    }
];

const selectActions: ContextMenuAction[] = [
    {
        name: "Select all",
        execute: dispatch => {
            dispatch(collectionPanelFilesAction.SELECT_ALL_COLLECTION_FILES());
        }
    },
    {
        name: "Unselect all",
        execute: dispatch => {
            dispatch(collectionPanelFilesAction.UNSELECT_ALL_COLLECTION_FILES());
        }
    },
];

const removeAction: ContextMenuAction = {
    name: "Remove selected",
    execute: dispatch => {
        dispatch(openMultipleFilesRemoveDialog());
    }
};

// These action sets are used on the multi-select actions button.
export const readOnlyCollectionFilesActionSet: ContextMenuActionSet = [
    selectActions,
    copyActions,
];

export const readOnlyCollectionFilesMultipleActionSet: ContextMenuActionSet = [
    selectActions,
    copyActionsMultiple,
];

export const collectionFilesActionSet: ContextMenuActionSet = readOnlyCollectionFilesActionSet.concat([[
    removeAction,
    ...moveActions
]]);

export const collectionFilesMultipleActionSet: ContextMenuActionSet = readOnlyCollectionFilesMultipleActionSet.concat([[
    removeAction,
    ...moveActionsMultiple
]]);
