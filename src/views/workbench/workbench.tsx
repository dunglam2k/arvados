// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import * as React from 'react';

import { StyleRulesCallback, Theme, WithStyles, withStyles } from '@material-ui/core/styles';
import Drawer from '@material-ui/core/Drawer';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import { connect, DispatchProp } from "react-redux";
import ProjectList from "../../components/project-list/project-list";
import { Route, Switch } from "react-router";
import { Link } from "react-router-dom";
import Button from "@material-ui/core/Button/Button";
import authActions from "../../store/auth/auth-action";
import IconButton from "@material-ui/core/IconButton/IconButton";
import Menu from "@material-ui/core/Menu/Menu";
import MenuItem from "@material-ui/core/MenuItem/MenuItem";
import { AccountCircle } from "@material-ui/icons";
import { User } from "../../models/user";
import Grid from "@material-ui/core/Grid/Grid";
import { RootState } from "../../store/store";
import MainAppBar, { MainAppBarActionProps, MainAppBarMenuItems, MainAppBarMenuItem } from '../../components/main-app-bar/main-app-bar';
import { Breadcrumb } from '../../components/breadcrumbs/breadcrumbs';
import { push } from 'react-router-redux';
import projectActions from "../../store/project/project-action";
import ProjectTree from '../../components/project-tree/project-tree';
import { TreeItem } from "../../components/tree/tree";
import { Project } from "../../models/project";
import { projectService } from '../../services/services';
import { findTreeBranch } from '../../store/project/project-reducer';

const drawerWidth = 240;

type CssRules = 'root' | 'appBar' | 'drawerPaper' | 'content' | 'toolbar';

const styles: StyleRulesCallback<CssRules> = (theme: Theme) => ({
    root: {
        flexGrow: 1,
        zIndex: 1,
        overflow: 'hidden',
        position: 'relative',
        display: 'flex',
        width: '100vw',
        height: '100vh'
    },
    appBar: {
        zIndex: theme.zIndex.drawer + 1,
        backgroundColor: '#692498',
        position: "absolute",
        width: "100%"
    },
    drawerPaper: {
        position: 'relative',
        width: drawerWidth,
    },
    content: {
        flexGrow: 1,
        backgroundColor: theme.palette.background.default,
        padding: theme.spacing.unit * 3,
        height: '100%',
        minWidth: 0,
    },
    toolbar: theme.mixins.toolbar
});

interface WorkbenchDataProps {
    projects: Array<TreeItem<Project>>;
    user?: User;
}

interface WorkbenchActionProps {
}

type WorkbenchProps = WorkbenchDataProps & WorkbenchActionProps & DispatchProp & WithStyles<CssRules>;

interface NavBreadcrumb extends Breadcrumb {
    itemId: string;
}

interface NavMenuItem extends MainAppBarMenuItem {
    action: () => void;
}

interface WorkbenchState {
    anchorEl: any;
    breadcrumbs: NavBreadcrumb[];
    searchText: string;
    menuItems: {
        accountMenu: NavMenuItem[],
        helpMenu: NavMenuItem[],
        anonymousMenu: NavMenuItem[]
    };
}

class Workbench extends React.Component<WorkbenchProps, WorkbenchState> {
    state = {
        anchorEl: null,
        searchText: "",
        breadcrumbs: [],
        menuItems: {
            accountMenu: [
                {
                    label: "Logout",
                    action: () => this.props.dispatch(authActions.LOGOUT())
                },
                {
                    label: "My account",
                    action: () => this.props.dispatch(push("/my-account"))
                }
            ],
            helpMenu: [
                {
                    label: "Help",
                    action: () => this.props.dispatch(push("/help"))
                }
            ],
            anonymousMenu: [
                {
                    label: "Sign in",
                    action: () => this.props.dispatch(authActions.LOGIN())
                }
            ]
        }
    };


    mainAppBarActions: MainAppBarActionProps = {
        onBreadcrumbClick: ({ itemId }: NavBreadcrumb) => {
            this.toggleProjectTreeItem(itemId);
        },
        onSearch: searchText => {
            this.setState({ searchText });
            this.props.dispatch(push(`/search?q=${searchText}`));
        },
        onMenuItemClick: (menuItem: NavMenuItem) => menuItem.action()
    };

    toggleProjectTreeItem = (itemId: string) => {
        const branch = findTreeBranch(this.props.projects, itemId);
        this.setState({
            breadcrumbs: branch.map(item => ({
                label: item.data.name,
                itemId: item.data.uuid
            }))
        });
        this.props.dispatch<any>(projectService.getProjectList(itemId)).then(() => {
            this.props.dispatch(projectActions.TOGGLE_PROJECT_TREE_ITEM(itemId));
        });
    }

    render() {
        const { classes, user } = this.props;
        return (
            <div className={classes.root}>
                <div className={classes.appBar}>
                    <MainAppBar
                        breadcrumbs={this.state.breadcrumbs}
                        searchText={this.state.searchText}
                        user={this.props.user}
                        menuItems={this.state.menuItems}
                        {...this.mainAppBarActions}
                    />
                </div>
                {user &&
                    <Drawer
                        variant="permanent"
                        classes={{
                            paper: classes.drawerPaper,
                        }}>
                        <div className={classes.toolbar} />
                        <ProjectTree
                            projects={this.props.projects}
                            toggleProjectTreeItem={this.toggleProjectTreeItem} />
                    </Drawer>}
                <main className={classes.content}>
                    <div className={classes.toolbar} />
                    <div className={classes.toolbar} />
                    <Switch>
                        <Route path="/project/:name" component={ProjectList} />
                    </Switch>
                </main>
            </div>
        );
    }
}

export default connect<WorkbenchDataProps>(
    (state: RootState) => ({
        projects: state.projects,
        user: state.auth.user
    })
)(
    withStyles(styles)(Workbench)
);
