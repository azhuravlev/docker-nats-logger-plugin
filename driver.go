package docker_nats_logger_plugin

import (
  "bytes"
  "fmt"
  "github.com/docker/docker/daemon/logger"
  "github.com/docker/docker/daemon/logger/loggerutils"
  "github.com/nats-io/nats.go"
  "github.com/sirupsen/logrus"
  "strconv"
)

const driverName = "natsLogger"

type natsLogger struct {
  nec       *nats.EncodedConn
  subject   string
  logFields map[string]interface{}
}

func (n *natsLogger) Name() string {
  return driverName
}

// Log takes a message and Publish it with the NATS client.
func (n *natsLogger) Log(logMsg *logger.Message) error {
  fields := make(map[string]interface{})
  for k, v := range n.logFields {
    fields[k] = v
  }
  fields["timestamp"] = logMsg.Timestamp.UTC()
  fields["line"] = string(logMsg.Line)
  fields["source"] = logMsg.Source

  return n.nec.Publish(n.subject, fields)
}

// Close terminates the connection to NATS,
// flushing any pending events before disconnecting.
func (n *natsLogger) Close() error {
  n.nec.Close()
  return nil
}

func init() {
  if err := logger.RegisterLogDriver(driverName, New); err != nil {
    logrus.Fatal(err)
  }
  if err := logger.RegisterLogOptValidator(driverName, ValidateLogOpt); err != nil {
    logrus.Fatal(err)
  }
}

// New creates a nats connection using the custom configuration values
// and container metadata passed in on the context
func New(info logger.Info) (logger.Logger, error) {
  var natsUrl string
  // Never stop reconnecting by default
  var maxReconnects = nats.MaxReconnects(-1)

  // --log-opt nats-servers="nats://127.0.0.1:4222,nats://127.0.0.1:4223"
  if v, ok := info.Config["nats-servers"]; ok {
    natsUrl = v
  } else {
    natsUrl = nats.DefaultURL
  }

  // --log-opt nats-max-reconnect=-1
  if v, ok := info.Config["nats-max-reconnect"]; ok {
    i, err := strconv.Atoi(v)
    if err != nil {
      return nil, err
    }
    maxReconnects = nats.MaxReconnects(i)
  }

  // Use standardized tag for the events and default subject
  tag, err := loggerutils.ParseLogTag(info, loggerutils.DefaultTemplate)
  if err != nil {
    return nil, err
  }

  // Subject under which log entries will be published, defaults
  // to using tag as the subject name.
  var subject string
  if v, ok := info.Config["nats-subject"]; ok {
    subject = v
  } else {
    subject = tag
  }

  connName := nats.Name(tag)

  // Create a single connection per container
  nc, err := nats.Connect(natsUrl, maxReconnects, connName)
  if err != nil {
    return nil, err
  }

  c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
  if err != nil {
    return nil, err
  }

  logrus.WithField("container", info.ContainerID).Infof("nats-logger: connected to %q", nc.ConnectedUrl())

  // Connection - related handlers
  nc.SetDisconnectErrHandler(func(c *nats.Conn, err error) {
    logrus.WithField("container", info.ContainerID).Warnf("nats-logger: disconnected")
  })

  nc.SetReconnectHandler(func(c *nats.Conn) {
    logrus.WithField("container", info.ContainerID).Warnf("nats-logger: reconnected to %q", c.ConnectedUrl())
  })

  nc.SetClosedHandler(func(c *nats.Conn) {
    logrus.WithField("container", info.ContainerID).Warnf("nats-logger: connection closed")
  })

  // Include hostname info in the record message
  hostname, err := info.Hostname()
  if err != nil {
    return nil, err
  }

  // Remove trailing slash from container name
  containerName := bytes.TrimLeft([]byte(info.ContainerName), "/")

  fields := make(map[string]interface{})
  fields["container_id"] = info.ContainerID
  fields["container_name"] = string(containerName)
  fields["image_id"] = info.ContainerImageID
  fields["image_name"] = info.ContainerImageName
  fields["hostname"] = hostname
  fields["tag"] = tag

  extra, err := info.ExtraAttributes(nil)
  if err != nil {
    return nil, err
  }
  for k, v := range extra {
    fields[k] = v
  }

  return &natsLogger{
    nec:       c,
    subject:   subject,
    logFields: fields,
  }, nil
}

// ValidateLogOpt looks for nats logger related custom options.
func ValidateLogOpt(cfg map[string]string) error {
  for key := range cfg {
    switch key {
    case "env":
    case "labels":
    case "nats-max-reconnect":
    case "nats-servers":
    case "nats-subject":
    case "tag":
    default:
      return fmt.Errorf("unknown log opt %q for nats log driver", key)
    }
  }

  return nil
}
