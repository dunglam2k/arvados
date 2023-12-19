// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import React from 'react';
import { StyleRulesCallback, WithStyles, withStyles, Card, CardContent, Typography, Grid } from '@material-ui/core';
import { ArvadosTheme } from 'common/custom-theme';
import { ResourceIcon } from 'components/icon/icon';
import { RootState } from 'store/store';
import { connect } from 'react-redux';
import { ClusterConfigJSON } from 'common/config';
import { NotFoundView } from 'views/not-found-panel/not-found-panel';
import { formatCWLResourceSize, formatCost, formatFileSize } from 'common/formatters';
import { DetailsAttribute } from 'components/details-attribute/details-attribute';
import { DefaultCodeSnippet } from 'components/default-code-snippet/default-code-snippet';

type CssRules = 'root' | 'infoBox' | 'instanceType';

const styles: StyleRulesCallback<CssRules> = (theme: ArvadosTheme) => ({
    root: {
       width: "calc(100% + 20px)",
       margin: "0 -10px",
       overflow: 'auto'
    },
    infoBox: {
        padding: "0 10px 10px",
    },
    instanceType: {
        padding: "10px",
    },
});

type InstanceTypesPanelConnectedProps = {config: ClusterConfigJSON};

type InstanceTypesPanelRootProps = InstanceTypesPanelConnectedProps & WithStyles<CssRules>;

const mapStateToProps = ({auth}: RootState): InstanceTypesPanelConnectedProps => ({
    config: auth.config.clusterConfig,
});

export const InstanceTypesPanel = withStyles(styles)(connect(mapStateToProps)(
    ({ config, classes }: InstanceTypesPanelRootProps) => {

        const instances = config.InstanceTypes || {};

        return <Grid className={classes.root} container direction="row">
            <Grid className={classes.infoBox} item xs={12}>
                <Card>
                    <CardContent>
                        <Typography variant="body2">
                            These are the cloud compute instance types
                            configured for this cluster. The core count and
                            maximum RAM request correspond to the greatest
                            values you can put in the CWL Workflow
                            ResourceRequest{" "}
                            <DefaultCodeSnippet
                                inline
                                lines={["minCores"]}
                            />{" "}
                            and{" "}
                            <DefaultCodeSnippet inline lines={["minRAM"]} />{" "}
                            and still be scheduled on that instance type.
                        </Typography>
                    </CardContent>
                </Card>
            </Grid>
            {Object.keys(instances).length > 0 ?
                Object.keys(instances)
                    .sort((a, b) => {
                        const typeA = instances[a];
                        const typeB = instances[b];

                        if (typeA.Price !== typeB.Price) {
                            return typeA.Price - typeB.Price;
                        } else {
                            return typeA.ProviderType.localeCompare(typeB.ProviderType);
                        }
                    }).map((instanceKey) => {
                        const instanceType = instances[instanceKey];
                        const diskRequest = instanceType.IncludedScratch;
                        const ramRequest = instanceType.RAM - config.Containers.ReserveExtraRAM;

                        return <Grid data-cy={instanceKey} className={classes.instanceType} item sm={6} xs={12} key={instanceKey}>
                            <Card>
                                <CardContent>
                                    <Typography variant="h6">
                                        {instanceKey}
                                    </Typography>
                                    <Grid item xs={12}>
                                        <DetailsAttribute label="Provider type" value={instanceType.ProviderType} />
                                    </Grid>
                                    <Grid item xs={12}>
                                        <DetailsAttribute label="Price" value={instanceType.Price} />
                                    </Grid>
                                    <Grid item xs={12}>
                                        <DetailsAttribute label="Cores" value={instanceType.VCPUs} />
                                    </Grid>
                                    <Grid item xs={12}>
                                        <DetailsAttribute label="Max RAM request" value={`${formatCWLResourceSize(ramRequest)} (${formatFileSize(ramRequest)})`} />
                                    </Grid>
                                    <Grid item xs={12}>
                                        <DetailsAttribute label="Max disk request" value={`${formatCWLResourceSize(diskRequest)} (${formatFileSize(diskRequest)})`} />
                                    </Grid>
                                    <Grid item xs={12}>
                                        <DetailsAttribute label="Preemptible" value={instanceType.Preemptible.toString()} />
                                    </Grid>
                                    {instanceType.CUDA && instanceType.CUDA.DeviceCount > 0 ?
                                        <>
                                            <Grid item xs={12}>
                                                <DetailsAttribute label="CUDA GPUs" value={instanceType.CUDA.DeviceCount} />
                                            </Grid>
                                            <Grid item xs={12}>
                                                <DetailsAttribute label="Hardware capability" value={instanceType.CUDA.HardwareCapability} />
                                            </Grid>
                                            <Grid item xs={12}>
                                                <DetailsAttribute label="Driver version" value={instanceType.CUDA.DriverVersion} />
                                            </Grid>
                                        </> : <></>
                                    }
                                </CardContent>
                            </Card>
                        </Grid>;
                    }) :
                <NotFoundView
                    icon={ResourceIcon}
                    messages={["No instances found"]}
                />
            }
        </Grid>;
    }
));
