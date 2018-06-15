// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

export interface DataItem {
    name: string;
    type: string;
    owner: string;
    lastModified: string;
    fileSize?: number;
    status?: string;
}