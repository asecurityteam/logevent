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

With a set of events defined, the `logevent.FromContext` tool can be used to
pull a `Logger` instance from the context and then log an event. For example,
the above event could be emitted like this:

```go
logger.Info(UserOverLimit{
  UserID: "user1",
  TenantID: "tenant1",
  AttemptsOverLimit: 10,
})
```

The following would appear in the log line:

```json
{"level": "info", "message": "user-over-limit", "user_id": "user1", "tenant_id": "tenant1", "attempts_over_limit": 10}
```

The types of data are preserved in the JSON output and the annotated names are
used as the attributes names.

### Adding Adapters ###

The default `logevent.FromContext` tool assumes a given context contains an
instance of a logger from the `github.co/rs/xlog` package. This is the typical
logger used by Stride and the default choice for the package. However, virtually
any logging implementation can be plugged in by providing a `LogFunc` with
a signature of:

```golang
type LogFunc func(ctx context.Context, level LogLevel, message string, annotations map[string]interface{})
```

Any logger that can be composed into this format can be added as an option for
the event logging features. Calling `NewFromContextFunc` with a valid `LogFunc`
will return the equivalent of `logevent.FromContext` can can be used.


## Contributing ##

### License ###

This project is licensed under Apache 2.0. See LICENSE.txt for details.

### Contributing Agreement ###

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
