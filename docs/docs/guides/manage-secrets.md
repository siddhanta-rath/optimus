---
id: secret
title: Manage Secrets
---

During job execution, specific credentials are needed to access required resources, for example, BigQuery credential 
for BQ to BQ tasks. Users are able to register secrets on their own, manage it, and use it in tasks and hooks. 
Please go through [concepts](../concepts/overview.md) to know more about it.

## Registering secret with Optimus

Register a secret by run the following command:
```shell
$ optimus secret create someSecret someSecretValue
````

By default, Optimus will encode the secret value. However, to register secret that has been encoded, run the following 
command instead:
```shell
$ optimus secret create someSecret encodedSecretValue --base64
````

There is also a flexibility to register using an existing secret file, instead of providing the secret value in the 
command.
```shell
$ optimus secret create someSecret --file=/path/to/secret
```

Please note that registering a secret which already exists will result in error. Modifying an existing secret can be 
done using the Update command.

## Updating a secret

Update an already existing secret in Optimus can be done by using the following command:

```shell
$ optimus secret create someSecret someSecretValue
```

To provide a secret which has already been encoded, use the following option:
```shell
$ optimus secret create someSecret encodedSecretValue --base64
```

To provide the secret through a file instead of passing it as command:
```shell
$ optimus secret create someSecret --file=/path/to/secret
```

The update command can only update the secret which has already been registered by the user. Trying to update a secret 
which does not exist will result in error.
