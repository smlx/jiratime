#!/usr/bin/env sh

# This script can be used to inspect the raw worklog data returned by the Jira
# API.
#
# https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-worklogs/#api-rest-api-3-issue-issueidorkey-worklog-get

# e.g. ABC-123
ISSUE_ID=
# https://example.atlassian.net/_edge/tenant_info
CLOUD_ID=
# the email associated with the API token
EMAIL=
# https://id.atlassian.com/manage-profile/security/api-tokens
API_TOKEN=

curl -sSL --request GET \
  --url "https://api.atlassian.com/ex/jira/$CLOUD_ID/rest/api/3/issue/$ISSUE_ID/worklog" \
  --user "$EMAIL:$API_TOKEN" \
  --header 'Accept: application/json'
