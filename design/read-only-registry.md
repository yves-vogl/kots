# Support for Read-Only Registries

## Goals

- Configure kotsadm to use application images from a private registry
- Support using read-only registry during initial installation
- Support using read-only registry on application updates
- Support using read-only registry when settings are modified on the Registry Settings page in Admin Console
- Support readon-only registries in airgap and on-line modes

## Non Goals

- Delivering images to the read-only registry.
- Changing registry settings in airgap mode

## Background

Some production enviroments restrict image registry use to read-only mode.  This is an environment wide restriction.  Images are side-loaded into the registry from a trusted internal source.  Right now kotsadm will always push images to the private registry, making it impossible to use in such an environment.

## Detailed Design

### Install

A new argument will be added to `kots install` command to flag the registry as a read-only registry.

```
kubectl kots install --kotsadm-registry my.private.registry/myapp \
  --registry-username pulluser \
  --registry-password pullpassword \
  --registry-is-readonly \
  app-slug
```

When the above command is used, the following will happen:

1. All kotsadm pods will be rewritten to pull images from the private registry. This is existing functionality.
1. Provided registry credentials will be saved in `.dockerconfigjson` secret in the app's namespace.
1. During license upload in Admin Console UI, registry settings will be populated with these values.
1. After the application is installed, Registry Settings page will be populated with these settings.

Note: this changes current functionality where registry information specified in the command line arguments is not saved and is not applied to the application pods.

### Update

Online install will download updates, but will not push images to the private registry as long as the read-only flag is set.

Aigapped installs will accept airgap bundle, but will not push images to the private registry as long as the read-only flag is set.

Stretch: Replicated could provide imageless airgap bundles.

### Registry Settings page

1. Registry settings page will have a checkbox named Read only.
1. In aigapped environments this is always disabled
   - if read-only set, all fields can be modiefied
   - if read-only is not set, only user name and password can be modified (this matches current behavior)
1. When Registry Settings are saved:
   - if read-only is true, a new release is created, but images are not pushed to the registry.
   - if read-only is false, current functionality is unchanged.

## Limitations

TBD

## Assumptions

TBD

## Testing

TBD

## Alternatives Considered



## Security Considerations

- No open security considerations
