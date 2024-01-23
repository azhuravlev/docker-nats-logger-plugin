package docker_nats_logger_plugin

import (
  "fmt"
  "github.com/docker/go-plugins-helpers/sdk"
  "os"
)

func fatal(format string, vs ...interface{}) {
  fmt.Fprintf(os.Stderr, format, vs...)
  os.Exit(1)
}

func main() {
  sdkHandler := sdk.NewHandler(`{"Implements": ["LoggingDriver"]}`)

  sdkHandler.HandleFunc("/LogDriver.StartLogging", startLoggingHandler())
  sdkHandler.HandleFunc("/LogDriver.StopLogging", stopLoggingHandler())
  sdkHandler.HandleFunc("/LogDriver.Capabilities", reportCaps())

  err := sdkHandler.ServeUnix("natslogSocket", 0)
  if err != nil {
    fatal("Error in socket handler: %s", err)
  }
}
