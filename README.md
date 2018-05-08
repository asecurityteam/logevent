# logevent - A structured event logger abstraction #

## Usage ##

### Defining Events ###

There are no log events defined within this project. Instead, developers
are free to create and maintain their own events structures and schemas as
preferred. This project relies on the tag annotation feature, like JSON, to
render an event to the log stream.

Each event emitted must be defined as a struct. Each attribute of the
struct must be annotated with a `logevent` tag that identifies the desired
field name as it appears in the logs. Each struct must contain a `Message`
attribute. Optionally, fields may include a `default` tag to auto-populate
field values. For example, an event describing an a rate-limited call might
appear as:

```golang
type UserOverLimit struct {
  UserID string `logevent:"user_id"`
  TenantID string `logevent:"tenant_id"`
  Message string `logevent:"message,default=user-over-limit"`
  AttemptsOverLimit int `logevent:"attempts_over_limit"`
}
```

### Logging Events ###

### Adding Adapters ###

## Contributing ##

### License ###

This project is licensed under Apache 2.0. See LICENSE.txt for details.

### Contributing Agreement ###

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
