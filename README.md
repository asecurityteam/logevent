# logevent - A structured event logger abstraction #

## Usage ##

### Defining Events ###

There are no log events defined within this project. Instead, service developers
are free to create and maintain their own events structures and schemas as
preferred. Instead, this project relies on the tag annotation feature, like
JSON, to render an event to the log stream.

Each event a service emits must be defined as a struct. Each attribute of the
struct must be annotated with an `logevent` tag that identifies the desired
field name as it appears in the logs. Each struct must contain a `Message`
attribute. Optionally, fields may include a `default` tag to auto-populate
field values. For example, an informative event describing an invalid ASAP token
might be defined like this:

```go
type InvalidJWT struct {
  Issuer string `logevent:"issuer"`
  Reason string `logevent:"reason"`
  Message string `logevent:"message,default=invalid-jwt"`
}
```

### Adding Adapters ###

The default implementation of logevent leverages xlog but any logger may be
added using the `NewFromContextFunc` helper. This method will consume any
`LogFunc` and return a method for extracting a `Logger` from a given context.

### Logging Events ###

With a set of events defined, the `logevent.FromContext` tools can be used to
pull a Logger instance from the context and then log an event. For example, the
above event could be emitted like this:

```go
var token = goasap.FromContext(ctx)
var issuer, _ = token.Issuer()
var logger = logevent.FromContext(ctx)
logger.Info(InvalidJWT{Issuer: issuer, Reason:"timestamp expired"})
```

The following would appear in the log line:

```json
{"level": "info", "message": "invalid-asap-token", "reason": "timestamp expired", "issuer": "api-gateway"}
```

## Contributing ##

### License ###

This project is licensed under Apache 2.0. See LICENSE.txt for details.

### Contributing Agreement ###

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
