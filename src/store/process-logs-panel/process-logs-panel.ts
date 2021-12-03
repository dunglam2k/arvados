// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import { RootState } from '../store';
import { matchProcessLogRoute, matchProcessRoute } from 'routes/routes';

export interface ProcessLogsPanel {
    filters: string[];
    selectedFilter: string;
    logs: ProcessLogs;
}

export interface ProcessLogs {
    [logType: string]: string[];
}

export const getProcessPanelLogs = ({ selectedFilter, logs }: ProcessLogsPanel) => {
    return logs[selectedFilter];
};

export const getProcessLogsPanelCurrentUuid = ({ router }: RootState) => {
    const pathname = router.location ? router.location.pathname : '';
    const match = matchProcessLogRoute(pathname) || matchProcessRoute(pathname);
    return match ? match.params.id : undefined;
};
