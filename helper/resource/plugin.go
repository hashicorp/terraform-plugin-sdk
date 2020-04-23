package resource

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-plugin-sdk/acctest"
	grpcplugin "github.com/hashicorp/terraform-plugin-sdk/internal/helper/plugin"
	proto "github.com/hashicorp/terraform-plugin-sdk/internal/tfplugin5"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	tftest "github.com/hashicorp/terraform-plugin-test"
)

func runProviderCommand(f func() error, wd *tftest.WorkingDir, opts *plugin.ServeOpts) error {
	// offer an opt-out that runs tests in separate provider processes
	// this will behave just like prod
	if os.Getenv("TF_TEST_PROVIDERS_OOP") != "" {
		logToFile("not running provider server inline")
		return f()
	}

	logToFile("starting runProviderCommand:\n" + string(debug.Stack()))

	// by default, run tests in the same process as the test runner
	// using the reattach behavior in Terraform. This ensures we get
	// test coverage and enables the use of delve as a debugger.

	// we need to create our own listener for the plugin server (the
	// provider gRPC server) to listen on, so we know its address. By
	// default, plugin.Serve creates one and listens on it, but we
	// won't know the address if it does that, and we need to know the
	// address so we coordinate Terraform connecting to it. So we create
	// our own and make plugin.Serve use it.
	//
	// go-plugin will close this Listener when it's done with it. We
	// don't need to, and in fact should not, close it ourselves. We're
	// giving it to go-plugin to own.
	logToFile("creating listener")
	listener, err := serverListener()
	if err != nil {
		return err
	}
	opts.Listener = listener
	logToFile("created listener")

	// we're going to run the test orchestration code in a goroutine,
	// because plugin.Serve expects to be running on the main goroutine,
	// and uses os.Exit, so we're just going to mirror production here.
	// Something has to run in a goroutine, may as well be the orchestrator
	// code path. We make a channel of errors to return the error from it
	// when it's done and to keep track of when it's done. The channel will
	// be closed when the orchestration code is done running.
	done := make(chan error)
	go func() {
		defer close(done)
		// we build up a big reattach string here. This is given to
		// Terraform's reattach environment variable so Terraform knows
		// how to connect to the process running the server and
		// how to connect to the server itself.

		// the provider name is technically supposed to be specified
		// in the format returned by addrs.Provider.GetDisplay(), but
		// 1. I'm not importing the entire addrs package for this and
		// 2. we only get the provider name here. Fortunately, when
		// only a provider name is specified in a provider block--which
		// is how the config file we generate does things--Terraform
		// just automatically assumes it's in the hashicorp namespace
		// and the default registry.terraform.io host, so we can just
		// construct the output of GetDisplay() ourselves, based on the
		// provider name. GetDisplay() omits the default host, so for
		// our purposes this will always be hashicorp/PROVIDER_NAME.
		providerName := wd.GetHelper().GetPluginName()

		// providerName gets returned as terraform-provider-foo, and we
		// need just foo. So let's fix that.
		providerName = strings.TrimPrefix(providerName, "terraform-provider-")

		// We need to tell the provider which version of the Terraform
		// protocol to serve. Usually this is negotiated with Terraform
		// during the handshake that sets the server up, but because
		// we're manually setting the server up, it's on us to do.
		// Because the SDK only supports 0.12+ of Terraform at the
		// moment, we can just set this to 5 (the latest version of the
		// protocol) and call it a day. But if and when we get a version
		// 6, we're going to have to figure something out.
		protoVersion := 5 // TODO: make this configurable?

		// similarly, we need to tell the provider whether the protocol
		// should be served over netrpc (0.11 and before) or gRPC (0.12
		// and after). Fortunately, the SDK only supports gRPC, so we
		// can just hardcode this for the moment. This probably won't
		// change any time in the near future, but it would still,
		// in theory, be nice to make this configurable from the test
		// code for providers.
		protoType := "grpc" // TODO: make this configurable?

		reattachStr := fmt.Sprintf("hashicorp/%s=%d|%s|%s|%s|%d",
			providerName,
			protoVersion,
			listener.Addr().Network(),
			listener.Addr().String(),
			protoType,
			os.Getpid(),
		)
		wd.Setenv("TF_PROVIDER_REATTACH", reattachStr)
		logToFile("set TF_PROVIDER_REATTACH to " + reattachStr)

		// By default, Terraform kills the provider process when it
		// finishes running a command. But the provider process is the
		// test process in this case, and we want to run a bunch of
		// commands all against the same provider process (the test
		// process). So we need to tell Terraform to shut down the gRPC
		// server instead of killing the process, which fortunately
		// there's (now) an environment variable for.
		wd.Setenv("TF_PROVIDER_SOFT_STOP", "1")

		logToFile("running terraform command")
		err := f()
		logToFile("ran terraform command")

		// once we've run the Terraform command, let's remove the
		// reattach information from the WorkingDir's environment. The
		// WorkingDir will persist until the next call, but the server
		// in the reattach info doesn't exist anymore at this point, so
		// the reattach info is no longer valid. In theory it should be
		// overwritten in the next call, but just to avoid any
		// confusing bug reports, let's just unset the environment
		// variable altogether.
		wd.Unsetenv("TF_PROVIDER_REATTACH")
		logToFile("finished with orchestrator code")
		if err != nil {
			logToFile("got error from orchestrator code: " + err.Error())
			client, e := goplugin.NewClient(&goplugin.ClientConfig{
				Reattach: &goplugin.ReattachConfig{
					Protocol: goplugin.Protocol(protoType),
					Addr:     listener.Addr(),
					Pid:      os.Getpid(),
				},
			}).Client()
			if e != nil {
				panic(e)
			}
			e = client.Close()
			if e != nil {
				panic(e)
			}
			logToFile("closed provider server in the face of the error")
		}

		done <- err
		logToFile("sent error to done")
	}()

	// back in our main goroutine, we need to start up the gRPC server for
	// the plugin, so Terraform can connect to it. We're going to start up
	// a server for every Terraform command, so global mutable state gets
	// reset between commands, just like in prod.

	// We need to signal which plugin protocol versions Terraform wants us
	// to use. Normally we get this for free from the handshake, but
	// because we're launching the server ourselves, it's on us to set it.
	// We hardcode 5 here to signal Terraform only supports version 5 of
	// the protocol, which is all the SDK supports at the moment. We'll
	// need to revisit this if and when version 6 of the protocol comes
	// out.
	os.Setenv("PLUGIN_PROTOCOL_VERSIONS", "5")

	// go-plugin uses a magic environment variable (called a magic cookie)
	// to detect when a process is running as a plugin. plugin.Serve checks
	// for this magic cookie when starting the server, and quits out if the
	// magic cookie isn't found, as the provider process is only supported
	// as a plugin. So we need to set that magic cookie value on our
	// process so plugin.Serve sees it, so it's tricked into thinking we're
	// running as a plugin.
	os.Setenv(plugin.Handshake.MagicCookieKey, plugin.Handshake.MagicCookieValue)

	// plugin.Serve will block on this goroutine until Terraform is done.
	// When the Terraform command completes, it will send a GRPC request to
	// the server which will shut down the server and this function will
	// return. This means that we don't need to clean up the server;
	// Terraform does that for us.
	logToFile("starting provider")
	plugin.Serve(opts)
	logToFile("provider returned")

	// we wait here to block on Terraform completing. The plugin.Serve
	// function blocks until Terraform decides the plugin's work is
	// complete, but this receive will let us block until Terraform reports
	// that its own work is complete. This ensures both the client and
	// server are finished before progressing. It also surfaces any errors
	// that Terraform may have returned.
	err = <-done

	// return any error returned from the orchestration code running
	// Terraform commands
	return err
}

// defaultPluginServeOpts builds ths *plugin.ServeOpts that you usually want to
// use when running runProviderCommand. It just sets the ProviderFunc to return
// the provider under test.
func defaultPluginServeOpts(wd *tftest.WorkingDir, providers map[string]terraform.ResourceProvider) *plugin.ServeOpts {
	return &plugin.ServeOpts{
		ProviderFunc: acctest.TestProviderFunc,
		GRPCProviderFunc: func() proto.ProviderServer {
			return grpcplugin.NewGRPCProviderServerShim(acctest.TestProviderFunc())
		},
	}
}

// shamelessly copied from go-plugin's server.go file
func serverListener() (net.Listener, error) {
	if runtime.GOOS == "windows" {
		return serverListener_tcp()
	}

	return serverListener_unix()
}

// shamelessly copied from go-plugin's server.go file
func serverListener_tcp() (net.Listener, error) {
	envMinPort := os.Getenv("PLUGIN_MIN_PORT")
	envMaxPort := os.Getenv("PLUGIN_MAX_PORT")

	var minPort, maxPort int64
	var err error

	switch {
	case len(envMinPort) == 0:
		minPort = 0
	default:
		minPort, err = strconv.ParseInt(envMinPort, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Couldn't get value from PLUGIN_MIN_PORT: %v", err)
		}
	}

	switch {
	case len(envMaxPort) == 0:
		maxPort = 0
	default:
		maxPort, err = strconv.ParseInt(envMaxPort, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Couldn't get value from PLUGIN_MAX_PORT: %v", err)
		}
	}

	if minPort > maxPort {
		return nil, fmt.Errorf("ENV_MIN_PORT value of %d is greater than PLUGIN_MAX_PORT value of %d", minPort, maxPort)
	}

	for port := minPort; port <= maxPort; port++ {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		listener, err := net.Listen("tcp", address)
		if err == nil {
			return listener, nil
		}
	}

	return nil, errors.New("Couldn't bind plugin TCP listener")
}

// shamelessly copied from go-plugin's server.go file
func serverListener_unix() (net.Listener, error) {
	tf, err := ioutil.TempFile("", "plugin")
	if err != nil {
		return nil, err
	}
	path := tf.Name()

	// Close the file and remove it because it has to not exist for
	// the domain socket.
	if err := tf.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		return nil, err
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	// Wrap the listener in rmListener so that the Unix domain socket file
	// is removed on close.
	return &rmListener{
		Listener: l,
		Path:     path,
	}, nil
}

// shamelessly copied from go-plugin's server.go file
// rmListener is an implementation of net.Listener that forwards most
// calls to the listener but also removes a file as part of the close. We
// use this to cleanup the unix domain socket on close.
type rmListener struct {
	net.Listener
	Path string
}

// shamelessly copied from go-plugin's server.go file
func (l *rmListener) Close() error {
	// Close the listener itself
	if err := l.Listener.Close(); err != nil {
		return err
	}

	// Remove the file
	return os.Remove(l.Path)
}
