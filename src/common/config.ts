// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

import Axios from "axios";

export const CONFIG_URL = process.env.REACT_APP_ARVADOS_CONFIG_URL || "/config.json";

export interface Config {
    apiHost: string;
    keepWebHost: string;
}

export const fetchConfig = () => {
    return Axios
        .get<ConfigJSON>(CONFIG_URL + "?nocache=" + (new Date()).getTime())
        .then(response => response.data)
        .catch(() => Promise.resolve(getDefaultConfig()))
        .then(mapConfig);
};

interface ConfigJSON {
    API_HOST: string;
    KEEP_WEB_HOST: string;
}

const mapConfig = (config: ConfigJSON): Config => ({
    apiHost: addProtocol(config.API_HOST),
    keepWebHost: addProtocol(config.KEEP_WEB_HOST)
});

const getDefaultConfig = (): ConfigJSON => ({
    API_HOST: process.env.REACT_APP_ARVADOS_API_HOST || "",
    KEEP_WEB_HOST: process.env.REACT_APP_ARVADOS_KEEP_WEB_HOST || ""
});

const addProtocol = (url: string) => `${window.location.protocol}//${url}`;
