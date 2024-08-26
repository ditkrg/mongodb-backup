## Parameters

### values

| Name                         | Description                                | Value       |
| ---------------------------- | ------------------------------------------ | ----------- |
| `schedule`                   | Schedule for the cronjob                   | `nil`       |
| `concurrencyPolicy`          | Concurrency policy for the cronjob         | `Forbid`    |
| `failedJobsHistoryLimit`     | Number of failed jobs to keep              | `3`         |
| `successfulJobsHistoryLimit` | Number of successful jobs to keep          | `3`         |
| `activeDeadlineSeconds`      | Active deadline seconds for the cronjob    | `600`       |
| `ttlSecondsAfterFinished`    | TTL seconds after finished for the cronjob | `300`       |
| `backoffLimit`               | Backoff limit for the cronjob              | `3`         |
| `restartPolicy`              | Restart policy for the cronjob             | `OnFailure` |

### global

| Name                      | Description | Value |
| ------------------------- | ----------- | ----- |
| `global.image.fullPath`   |             | `nil` |
| `global.image.registry`   |             | `nil` |
| `global.image.repository` |             | `nil` |
| `global.image.tag`        |             | `nil` |
| `global.image.pullPolicy` |             | `nil` |

### image Image for the cronjob

| Name               | Description | Value                       |
| ------------------ | ----------- | --------------------------- |
| `image.fullPath`   |             | `nil`                       |
| `image.registry`   |             | `reg.dev.krd`               |
| `image.repository` |             | `common/mongodb-backup-cli` |
| `image.tag`        |             | `nil`                       |
| `image.pullPolicy` |             | `IfNotPresent`              |

### securityContext Security context for the cronjob (can replace the whole object)

| Name                                       | Description | Value     |
| ------------------------------------------ | ----------- | --------- |
| `securityContext.runAsUser`                |             | `1001`    |
| `securityContext.runAsGroup`               |             | `1001`    |
| `securityContext.runAsNonRoot`             |             | `true`    |
| `securityContext.allowPrivilegeEscalation` |             | `false`   |
| `securityContext.privileged`               |             | `false`   |
| `securityContext.capabilities.drop`        |             | `["ALL"]` |

### ephemeral Ephemeral volume for the cronjob (can replace the whole object)

| Name                                                           | Description | Value |
| -------------------------------------------------------------- | ----------- | ----- |
| `ephemeral.volumeClaimTemplate.spec.resources.limits.storage`  |             | `2Gi` |
| `ephemeral.volumeClaimTemplate.spec.resources.request.storage` |             | `2Gi` |

### resources Resources for the cronjob (can replace the whole object)

| Name                      | Description | Value   |
| ------------------------- | ----------- | ------- |
| `resources.requests.cpu`  |             | `100m`  |
| `resources.limits.memory` |             | `200Mi` |

### jobs list of cronjobs to create

| Name                                               | Description                                                                     | Value   |
| -------------------------------------------------- | ------------------------------------------------------------------------------- | ------- |
| `jobs.job-name.schedule`                           | Schedule for the cronjob (Required)                                             | `""`    |
| `jobs.job-name.args`                               | Arguments for the cronjob (Required)                                            | `[]`    |
| `jobs.job-name.configs`                            | Configs for the cronjob, key value pairs                                        | `nil`   |
| `jobs.job-name.concurrencyPolicy`                  | Concurrency policy for the cronjob (Optional, defaults to global value)         | `nil`   |
| `jobs.job-name.failedJobsHistoryLimit`             | Number of failed jobs to keep (Optional, defaults to global value)              | `nil`   |
| `jobs.job-name.successfulJobsHistoryLimit`         | Number of successful jobs to keep (Optional, defaults to global value)          | `nil`   |
| `jobs.job-name.activeDeadlineSeconds`              | Active deadline seconds for the cronjob (Optional, defaults to global value)    | `nil`   |
| `jobs.job-name.ttlSecondsAfterFinished`            | TTL seconds after finished for the cronjob (Optional, defaults to global value) | `nil`   |
| `jobs.job-name.backoffLimit`                       | Backoff limit for the cronjob (Optional, defaults to global value)              | `nil`   |
| `jobs.job-name.restartPolicy`                      | Restart policy for the cronjob (Optional, defaults to global value)             | `""`    |
| `jobs.job-name.securityContext`                    | Security context for the cronjob (Optional, defaults to global value)           | `{}`    |
| `jobs.job-name.ephemeral`                          | Ephemeral volume for the cronjob (Optional, defaults to global value)           | `{}`    |
| `jobs.job-name.resources`                          | Resources for the cronjob (Optional, defaults to global value)                  | `{}`    |
| `jobs.job-name.env`                                | Environment variables for the cronjob                                           | `[]`    |
| `jobs.job-name.envFrom`                            | Environment variables for the cronjob                                           | `[]`    |
| `jobs.job-name.sealedSecrets.enabled`              | Enable sealed secrets                                                           | `false` |
| `jobs.job-name.sealedSecrets.name`                 | Name of the sealed secret (Optional, defaults to the <job name>-sealed)         | `nil`   |
| `jobs.job-name.sealedSecrets.encryptedData`        | Encrypted data for the sealed secret                                            | `nil`   |
| `jobs.job-name.infisical.enabled`                  | Enable infisical                                                                | `false` |
| `jobs.job-name.infisical.sealedSecret.name`        | Name of the infisical sealed secret (Defaults to <job name>-infisical-token)    | `nil`   |
| `jobs.job-name.infisical.sealedSecret.namespace`   | Namespace for the infisical sealed secret (Defaults to the release namespace)   | `nil`   |
| `jobs.job-name.infisical.sealedSecret.token`       | Token for the infisical sealed secret                                           | `nil`   |
| `jobs.job-name.infisical.secretsScope.secretsPath` | Secrets path for the infisical                                                  | `""`    |
| `jobs.job-name.infisical.secretsScope.envSlug`     | Env slug for the infisical                                                      | `""`    |
| `jobs.job-name.infisical.managedSecret.name`       | Name of the managed secret (Defaults to <job name>-managed)                     | `nil`   |
| `jobs.job-name.infisical.managedSecret.namespace`  | Namespace for the managed secret (Defaults to the release namespace)            | `nil`   |
