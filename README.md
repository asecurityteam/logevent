<a id="markdown-logevent---a-structured-event-logger-abstraction" name="logevent---a-structured-event-logger-abstraction"></a>
# logevent - A structured event logger abstraction
[![GoDoc](https://godoc.org/github.com/asecurityteam/logevent?status.svg)](https://godoc.org/github.com/asecurityteam/logevent)
<!-- TOC -->

- [logevent - A structured event logger abstraction](#logevent---a-structured-event-logger-abstraction)
    - [Usage](#usage)
        - [Defining Events](#defining-events)
        - [Logging Events](#logging-events)
        - [Adding Adapters](#adding-adapters)
    - [Contributing](#contributing)
        - [License](#license)
        - [Contributing Agreement](#contributing-agreement)

<!-- /TOC -->

<a id="markdown-usage" name="usage"></a>
## Usage

<a id="markdown-defining-events" name="defining-events"></a>
### Defining Events

There are no log events defined within this project. Instead, developers
are free to create and maintain their own events structures and schema as
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

<a id="markdown-logging-events" name="logging-events"></a>
### Logging Events

```golang
logevent.FromContext(ctx).Info(myEvent{}) // Log the event
logevent.FromContext(ctx).SetField("key", "value")
logevent.FromContext(ctx).Warn("uh oh") // Fall back to string logging if not an event.
var newCtx = logevent.NewContext(context.Background(), logger.Copy())
```

<a id="markdown-adding-adapters" name="adding-adapters"></a>
### Adding Adapters

<a id="markdown-contributing" name="contributing"></a>
## Contributing

<a id="markdown-license" name="license"></a>
### License

This project is licensed under Apache 2.0. See LICENSE.txt for details.

<a id="markdown-contributing-agreement" name="contributing-agreement"></a>
### Contributing Agreement

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
