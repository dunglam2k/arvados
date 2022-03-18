// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import React, { useState } from 'react';
import {
    StyleRulesCallback,
    WithStyles,
    withStyles,
    Card,
    CardHeader,
    IconButton,
    CardContent,
    Tooltip,
    Grid,
    Typography,
} from '@material-ui/core';
import { ArvadosTheme } from 'common/custom-theme';
import {
    CloseIcon,
    CollectionIcon,
    LogIcon,
    MaximizeIcon,
    WordWrapIcon
} from 'components/icon/icon';
import { Process } from 'store/processes/process';
import { MPVPanelProps } from 'components/multi-panel-view/multi-panel-view';
import {
    FilterOption,
    ProcessLogForm
} from 'views/process-panel/process-log-form';
import { ProcessLogCodeSnippet } from 'views/process-panel/process-log-code-snippet';
import { DefaultView } from 'components/default-view/default-view';
import { CodeSnippetDataProps } from 'components/code-snippet/code-snippet';

type CssRules = 'card' | 'content' | 'title' | 'iconHeader' | 'header' | 'root' | 'logViewer' | 'logViewerContainer';

const styles: StyleRulesCallback<CssRules> = (theme: ArvadosTheme) => ({
    card: {
        height: '100%'
    },
    header: {
        paddingTop: theme.spacing.unit,
        paddingBottom: theme.spacing.unit,
    },
    content: {
        padding: theme.spacing.unit * 0,
        height: '100%',
    },
    logViewer: {
        height: '100%',
    },
    logViewerContainer: {
        height: '100%',
    },
    title: {
        overflow: 'hidden',
        paddingTop: theme.spacing.unit * 0.5
    },
    iconHeader: {
        fontSize: '1.875rem',
        color: theme.customs.colors.green700
    },
    root: {
        height: '100%',
    },
});

export interface ProcessLogsCardDataProps {
    process: Process;
    selectedFilter: FilterOption;
    filters: FilterOption[];
}

export interface ProcessLogsCardActionProps {
    onLogFilterChange: (filter: FilterOption) => void;
    navigateToLog: (uuid: string) => void;
}

type ProcessLogsCardProps = ProcessLogsCardDataProps
    & ProcessLogsCardActionProps
    & CodeSnippetDataProps
    & WithStyles<CssRules>
    & MPVPanelProps;

export const ProcessLogsCard = withStyles(styles)(
    ({ classes, process, filters, selectedFilter, lines, onLogFilterChange, navigateToLog,
        doHidePanel, doMaximizePanel, panelMaximized, panelName }: ProcessLogsCardProps) => {
        const [wordWrapToggle, setWordWrapToggle] = useState<boolean>(true);
        return <Grid item className={classes.root} xs={12}>
            <Card className={classes.card}>
                <CardHeader className={classes.header}
                    avatar={<LogIcon className={classes.iconHeader} />}
                    action={<Grid container direction='row' alignItems='center'>
                        <Grid item>
                            <ProcessLogForm selectedFilter={selectedFilter}
                                filters={filters} onChange={onLogFilterChange} />
                        </Grid>
                        <Grid item>
                            <Tooltip title="Toggle word wrapping" disableFocusListener>
                                <IconButton onClick={() => setWordWrapToggle(!wordWrapToggle)}>
                                    <WordWrapIcon />
                                </IconButton>
                            </Tooltip>
                        </Grid>
                        <Grid item>
                            <Tooltip title="Go to Log collection" disableFocusListener>
                                <IconButton onClick={() => navigateToLog(process.containerRequest.logUuid!)}>
                                    <CollectionIcon />
                                </IconButton>
                            </Tooltip>
                        </Grid>
                        { doMaximizePanel && !panelMaximized &&
                        <Tooltip title={`Maximize ${panelName || 'panel'}`} disableFocusListener>
                            <IconButton onClick={doMaximizePanel}><MaximizeIcon /></IconButton>
                        </Tooltip> }
                        { doHidePanel && <Grid item>
                            <Tooltip title={`Close ${panelName || 'panel'}`} disableFocusListener>
                                <IconButton onClick={doHidePanel}><CloseIcon /></IconButton>
                            </Tooltip>
                        </Grid> }
                    </Grid>}
                    title={
                        <Typography noWrap variant='h6' className={classes.title}>
                            Logs
                        </Typography>}
                />
                <CardContent className={classes.content}>
                    {lines.length > 0
                        ? < Grid
                            className={classes.logViewerContainer}
                            container
                            spacing={24}
                            direction='column'>
                            <Grid className={classes.logViewer} item xs>
                                <ProcessLogCodeSnippet wordWrap={wordWrapToggle} lines={lines} />
                            </Grid>
                        </Grid>
                        : <DefaultView
                            icon={LogIcon}
                            messages={['No logs yet']} />
                    }
                </CardContent>
            </Card>
        </Grid >
});

