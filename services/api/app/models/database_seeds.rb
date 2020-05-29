# Copyright (C) The Arvados Authors. All rights reserved.
#
# SPDX-License-Identifier: AGPL-3.0

require 'update_permissions'

class DatabaseSeeds
  extend CurrentApiClient
  def self.install
    system_user
    system_group
    all_users_group
    anonymous_group
    anonymous_group_read_permission
    anonymous_user
    refresh_permissions
    refresh_trashed
  end
end
