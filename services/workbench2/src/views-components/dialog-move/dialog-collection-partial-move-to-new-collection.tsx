// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import React from "react";
import { memoize } from "lodash/fp";
import { FormDialog } from 'components/form-dialog/form-dialog';
import { CollectionNameField, CollectionDescriptionField, CollectionProjectPickerField } from 'views-components/form-fields/collection-form-fields';
import { WithDialogProps } from 'store/dialog/with-dialog';
import { InjectedFormProps } from 'redux-form';
import { CollectionPartialMoveToNewCollectionFormData } from "store/collections/collection-partial-move-actions";
import { PickerIdProp } from "store/tree-picker/picker-id";

type DialogCollectionPartialMoveProps = WithDialogProps<string> & InjectedFormProps<CollectionPartialMoveToNewCollectionFormData>;

export const DialogCollectionPartialMoveToNewCollection = (props: DialogCollectionPartialMoveProps & PickerIdProp) =>
    <FormDialog
        dialogTitle='Move to new collection'
        formFields={CollectionPartialMoveFields(props.pickerId)}
        submitLabel='Create collection'
        {...props}
    />;

const CollectionPartialMoveFields = memoize(
    (pickerId: string) =>
        () =>
            <>
                <CollectionNameField />
                <CollectionDescriptionField />
                <CollectionProjectPickerField {...{ pickerId }} />
            </>);
