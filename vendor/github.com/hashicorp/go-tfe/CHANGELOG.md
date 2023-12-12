# v1.3.0

## Enhancements
* Adds support for Microsoft Teams notification configuration by @JarrettSpiker [#398](https://github.com/hashicorp/go-tfe/pull/389)
* Add support for Audit Trail API by @sebasslash [#407](https://github.com/hashicorp/go-tfe/pull/407)
* Adds Private Registry Provider, Provider Version, and Provider Platform APIs support by @joekarl and @annawinkler [#313](https://github.com/hashicorp/go-tfe/pull/313)
* Adds List Registry Modules endpoint by @chroju [#385](https://github.com/hashicorp/go-tfe/pull/385)
* Adds `WebhookURL` field to `VCSRepo` struct by @kgns [#413](https://github.com/hashicorp/go-tfe/pull/413)
* Adds `Category` field to `VariableUpdateOptions` struct by @jtyr [#397](https://github.com/hashicorp/go-tfe/pull/397)
* Adds `TriggerPatterns` to `Workspace` by @matejrisek [#400](https://github.com/hashicorp/go-tfe/pull/400)
* [beta] Adds `ExtState` field to `StateVersionCreateOptions` by @brandonc [#416](https://github.com/hashicorp/go-tfe/pull/416)

# v1.2.0

## Enhancements
* Adds support for reading current state version outputs to StateVersionOutputs, which can be useful for reading outputs when users don't have the necessary permissions to read the entire state by @brandonc [#370](https://github.com/hashicorp/go-tfe/pull/370)
* Adds Variable Set methods for `ApplyToWorkspaces` and `RemoveFromWorkspaces` by @byronwolfman [#375](https://github.com/hashicorp/go-tfe/pull/375)
* Adds `Names` query param field to `TeamListOptions` by @sebasslash [#393](https://github.com/hashicorp/go-tfe/pull/393)
* Adds `Emails` query param field to `OrganizationMembershipListOptions` by @sebasslash [#393](https://github.com/hashicorp/go-tfe/pull/393)
* Adds Run Tasks API support by @glennsarti [#381](https://github.com/hashicorp/go-tfe/pull/381), [#382](https://github.com/hashicorp/go-tfe/pull/382) and [#383](https://github.com/hashicorp/go-tfe/pull/383)


## Bug fixes
* Fixes ignored comment when performing apply, discard, cancel, and force-cancel run actions [#388](https://github.com/hashicorp/go-tfe/pull/388)

# v1.1.0

## Enhancements

* Add Variable Set API support by @rexredinger [#305](https://github.com/hashicorp/go-tfe/pull/305)
* Add Comments API support by @alex-ikse [#355](https://github.com/hashicorp/go-tfe/pull/355)
* Add beta support for SSOTeamID to `Team`, `TeamCreateOptions`, `TeamUpdateOptions` by @xlgmokha [#364](https://github.com/hashicorp/go-tfe/pull/364)

# v1.0.0

## Breaking Changes
* Renamed methods named Generate to Create for `AgentTokens`, `OrganizationTokens`, `TeamTokens`, `UserTokens` by @sebasslash [#327](https://github.com/hashicorp/go-tfe/pull/327)
* Methods that express an action on a relationship have been prefixed with a verb, e.g `Current()` is now `ReadCurrent()` by @sebasslash [#327](https://github.com/hashicorp/go-tfe/pull/327)
* All list option structs are now pointers @uturunku1 [#309](https://github.com/hashicorp/go-tfe/pull/309)
* All errors have been refactored into constants in `errors.go` @uturunku1 [#310](https://github.com/hashicorp/go-tfe/pull/310)
* The `ID` field in Create/Update option structs has been renamed to `Type` in accordance with the JSON:API spec by @omarismail, @uturunku1 [#190](https://github.com/hashicorp/go-tfe/pull/190), [#323](https://github.com/hashicorp/go-tfe/pull/323), [#332](https://github.com/hashicorp/go-tfe/pull/332)
* Nested URL params (consisting of an organization, module and provider name) used to identify a `RegistryModule` have been refactored into a struct `RegistryModuleID` by @sebasslash [#337](https://github.com/hashicorp/go-tfe/pull/337)


## Enhancements
* Added missing include fields for `AdminRuns`, `AgentPools`, `ConfigurationVersions`, `OAuthClients`, `Organizations`, `PolicyChecks`, `PolicySets`, `Policies` and `RunTriggers` by @uturunku1 [#339](https://github.com/hashicorp/go-tfe/pull/339)
* Cleanup documentation and improve consistency by @uturunku1 [#331](https://github.com/hashicorp/go-tfe/pull/331)
* Add more linters to our CI pipeline by @sebasslash [#326](https://github.com/hashicorp/go-tfe/pull/326)
* Resolve `TFE_HOSTNAME` as fallback for `TFE_ADDRESS` by @sebasslash [#340](https://github.com/hashicorp/go-tfe/pull/326)
* Adds a `fetching` status to `RunStatus` and adds the `Archive` method to the ConfigurationVersions interface by @mpminardi [#338](https://github.com/hashicorp/go-tfe/pull/338)
* Added a `Download` method to the `ConfigurationVersions` interface by @tylerwolf [#358](https://github.com/hashicorp/go-tfe/pull/358)
* API Coverage documentation by @laurenolivia [#334](https://github.com/hashicorp/go-tfe/pull/334)

## Bug Fixes
* Fixed invalid memory address error when `AdminSMTPSettingsUpdateOptions.Auth` field is empty and accessed by @uturunku1 [#335](https://github.com/hashicorp/go-tfe/pull/335)
