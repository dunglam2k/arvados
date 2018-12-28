// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import * as React from "react";
import { compose } from 'redux';
import {
    withStyles, Dialog, DialogTitle, DialogContent, DialogActions,
    Button, StyleRulesCallback, WithStyles, Grid
} from '@material-ui/core';
import { WithDialogProps, withDialog } from "~/store/dialog/with-dialog";
import { COMPUTE_NODE_ATTRIBUTES_DIALOG } from '~/store/compute-nodes/compute-nodes-actions';
import { ArvadosTheme } from '~/common/custom-theme';
import { NodeResource, NodeProperties, NodeInfo } from '~/models/node';
import * as classnames from "classnames";

type CssRules = 'root' | 'grid';

const styles: StyleRulesCallback<CssRules> = (theme: ArvadosTheme) => ({
    root: {
        fontSize: '0.875rem',
        '& div:nth-child(odd):not(.nestedRoot)': {
            textAlign: 'right',
            color: theme.palette.grey["500"]
        },
        '& div:nth-child(even)': {
            overflowWrap: 'break-word'
        }
    },
    grid: {
        padding: '8px 0 0 0'
    } 
});

interface AttributesComputeNodeDialogDataProps {
    computeNode: NodeResource;
}

export const AttributesComputeNodeDialog = compose(
    withDialog(COMPUTE_NODE_ATTRIBUTES_DIALOG),
    withStyles(styles))(
        ({ open, closeDialog, data, classes }: WithDialogProps<AttributesComputeNodeDialogDataProps> & WithStyles<CssRules>) =>
            <Dialog open={open} onClose={closeDialog} fullWidth maxWidth='sm'>
                <DialogTitle>Attributes</DialogTitle>
                <DialogContent>
                    {data.computeNode && <div>
                        {renderPrimaryInfo(data.computeNode, classes)}
                        {renderInfo(data.computeNode.info, classes)}
                        {renderProperties(data.computeNode.properties, classes)}
                    </div>}
                </DialogContent>
                <DialogActions>
                    <Button
                        variant='text'
                        color='primary'
                        onClick={closeDialog}>
                        Close
                    </Button>
                </DialogActions>
            </Dialog>
    );

const renderPrimaryInfo = (computeNode: NodeResource, classes: any) => {
    const { uuid, ownerUuid, createdAt, modifiedAt, modifiedByClientUuid, modifiedByUserUuid } = computeNode;
    return (
        <Grid container direction="row" spacing={16} className={classes.root}>
            <Grid item xs={5}>UUID</Grid>
            <Grid item xs={7}>{uuid}</Grid>
            <Grid item xs={5}>Owner uuid</Grid>
            <Grid item xs={7}>{ownerUuid}</Grid>
            <Grid item xs={5}>Created at</Grid>
            <Grid item xs={7}>{createdAt}</Grid>
            <Grid item xs={5}>Modified at</Grid>
            <Grid item xs={7}>{modifiedAt}</Grid>
            <Grid item xs={5}>Modified by user uuid</Grid>
            <Grid item xs={7}>{modifiedByUserUuid}</Grid>
            <Grid item xs={5}>Modified by client uuid</Grid>
            <Grid item xs={7}>{modifiedByClientUuid || '(none)'}</Grid>
        </Grid>
    );
};

const renderInfo = (info: NodeInfo, classes: any) => {
    const { lastAction, pingSecret, ec2InstanceId, slurmState } = info;
    return (
        <Grid container direction="row" spacing={16} className={classnames([classes.root, classes.grid])}>
            <Grid item xs={5}>Info - Last action</Grid>
            <Grid item xs={7}>{lastAction || '(none)'}</Grid>
            <Grid item xs={5}>Info - Ping secret</Grid>
            <Grid item xs={7}>{pingSecret || '(none)'}</Grid>
            <Grid item xs={5}>Info - ec2 instance id</Grid>
            <Grid item xs={7}>{ec2InstanceId || '(none)'}</Grid>
            <Grid item xs={5}>Info - Slurm state</Grid>
            <Grid item xs={7}>{slurmState || '(none)'}</Grid>
        </Grid>
    );
};

const renderProperties = (properties: NodeProperties, classes: any) => {
    const { totalRamMb, totalCpuCores, totalScratchMb, cloudNode } = properties;
    return (
        <Grid container direction="row" spacing={16} className={classnames([classes.root, classes.grid])}>
            <Grid item xs={5}>Properties - Total ram mb</Grid>
            <Grid item xs={7}>{totalRamMb || '(none)'}</Grid>
            <Grid item xs={5}>Properties - Total scratch mb</Grid>
            <Grid item xs={7}>{totalScratchMb || '(none)'}</Grid>
            <Grid item xs={5}>Properties - Total cpu cores</Grid>
            <Grid item xs={7}>{totalCpuCores || '(none)'}</Grid>
            <Grid item xs={5}>Properties - Cloud node size </Grid>
            <Grid item xs={7}>{cloudNode ? cloudNode.size : '(none)'}</Grid>
            <Grid item xs={5}>Properties - Cloud node price</Grid>
            <Grid item xs={7}>{cloudNode ? cloudNode.price : '(none)'}</Grid>
        </Grid>
    );
};