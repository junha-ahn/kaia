// Modifications Copyright 2024 The Kaia Authors
// Modifications Copyright 2018 The klaytn Authors
// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from node/config.go (2018/06/04).
// Modified and improved for the klaytn development.
// Modified and improved for the Kaia development.

package node

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kaiachain/kaia/accounts"
	"github.com/kaiachain/kaia/accounts/keystore"
	"github.com/kaiachain/kaia/common"
	"github.com/kaiachain/kaia/crypto"
	"github.com/kaiachain/kaia/crypto/bls"
	"github.com/kaiachain/kaia/log"
	"github.com/kaiachain/kaia/networks/p2p"
	"github.com/kaiachain/kaia/networks/p2p/discover"
	"github.com/kaiachain/kaia/networks/rpc"
	"github.com/kaiachain/kaia/storage/database"
)

const (
	datadirPrivateKey      = "nodekey"            // Path within the datadir to the node's private key
	DatadirBlsSecretKey    = "bls-nodekey"        // Path within the datadir to the node's bls secret key
	datadirDefaultKeyStore = "keystore"           // Path within the datadir to the keystore
	datadirStaticNodes     = "static-nodes.json"  // Path within the datadir to the static node list
	datadirTrustedNodes    = "trusted-nodes.json" // Path within the datadir to the trusted node list
	datadirNodeDatabase    = "nodes"              // Path within the datadir to store the node infos
)

// Config represents a small collection of configuration values to fine tune the
// P2P network layer of a protocol stack. These values can be further extended by
// all registered services.
type Config struct {
	// Name sets the instance name of the node. It must not contain the / character and is
	// used in the devp2p node identifier. The instance name of Kaia is one among "kcn", "ken", and "kpn".
	// If no value is specified, the basename of the current executable is used.
	Name string `toml:"-"`

	// UserIdent, if set, is used as an additional component in the devp2p node identifier.
	UserIdent string `toml:",omitempty"`

	// Version should be set to the version number of the program. It is used
	// in the devp2p node identifier.
	Version string `toml:"-"`

	// key-value database type [LevelDB, RocksDB, BadgerDB, MemoryDB, DynamoDB]
	DBType database.DBType

	// DataDir is the file system folder the node should use for any data storage
	// requirements. The configured data directory will not be directly shared with
	// registered services, instead those can use utility methods to create/access
	// databases or flat files. This enables ephemeral nodes which can fully reside
	// in memory.
	DataDir string

	// ChainDataDir is separate directory path for chaindata. This is introduced in order to
	// serve API from multiple processes to share chaindata with others.
	ChainDataDir string

	// Configuration of peer-to-peer networking.
	// Includes the ECDSA NodeKey
	P2P p2p.Config

	// KeyStoreDir is the file system folder that contains private keys. The directory can
	// be specified as a relative path, in which case it is resolved relative to the
	// current directory.
	//
	// If KeyStoreDir is empty, the default location is the "keystore" subdirectory of
	// DataDir. If DataDir is unspecified and KeyStoreDir is empty, an ephemeral directory
	// is created by New and destroyed when the node is stopped.
	KeyStoreDir string `toml:",omitempty"`

	// BlsKey is the BLS secret key for Randao consensus.
	// If BlsKey is empty, the node will look for the default location which is the
	// "bls-nodekey" file in DataDir. If the "bls-nodekey" does not exist, then BlsKey
	// is derived from the ECDSA NodeKey.
	BlsKey bls.SecretKey `toml:"-"` // ignored by toml marshaller

	// UseLightweightKDF lowers the memory and CPU requirements of the key store
	// scrypt KDF at the expense of security.
	UseLightweightKDF bool `toml:",omitempty"`

	// IPCPath is the requested location to place the IPC endpoint. If the path is
	// a simple file name, it is placed inside the data directory (or on the root
	// pipe path on Windows), whereas if it's a resolvable path name (absolute or
	// relative), then that specific path is enforced. An empty path disables IPC.
	IPCPath string `toml:",omitempty"`

	// HTTP module type is http server module type (fasthttp and http)
	HTTPServerType string `toml:",omitempty"`

	// HTTPHost is the host interface on which to start the HTTP RPC server. If this
	// field is empty, no HTTP API endpoint will be started.
	HTTPHost string `toml:",omitempty"`

	// HTTPPort is the TCP port number on which to start the HTTP RPC server. The
	// default zero value is/ valid and will pick a port number randomly (useful
	// for ephemeral nodes).
	HTTPPort int `toml:",omitempty"`

	// HTTPCors is the Cross-Origin Resource Sharing header to send to requesting
	// clients. Please be aware that CORS is a browser enforced security, it's fully
	// useless for custom HTTP clients.
	HTTPCors []string `toml:",omitempty"`

	// HTTPVirtualHosts is the list of virtual hostnames which are allowed on incoming requests.
	// This is by default {'localhost'}. Using this prevents attacks like
	// DNS rebinding, which bypasses SOP by simply masquerading as being within the same
	// origin. These attacks do not utilize CORS, since they are not cross-domain.
	// By explicitly checking the Host-header, the server will not allow requests
	// made against the server with a malicious host domain.
	// Requests using ip address directly are not affected
	HTTPVirtualHosts []string `toml:",omitempty"`

	// HTTPModules is a list of API modules to expose via the HTTP RPC interface.
	// If the module list is empty, all RPC API endpoints designated public will be
	// exposed.
	HTTPModules []string `toml:",omitempty"`

	// HTTPTimeouts allows for customization of the timeout values used by the HTTP RPC
	// interface.
	HTTPTimeouts rpc.HTTPTimeouts

	// WSHost is the host interface on which to start the websocket RPC server. If
	// this field is empty, no websocket API endpoint will be started.
	WSHost string `toml:",omitempty"`

	// WSPort is the TCP port number on which to start the websocket RPC server. The
	// default zero value is/ valid and will pick a port number randomly (useful for
	// ephemeral nodes).
	WSPort int `toml:",omitempty"`

	// WSOrigins is the list of domain to accept websocket requests from. Please be
	// aware that the server can only act upon the HTTP request the client sends and
	// cannot verify the validity of the request header.
	WSOrigins []string `toml:",omitempty"`

	// WSModules is a list of API modules to expose via the websocket RPC interface.
	// If the module list is empty, all RPC API endpoints designated public will be
	// exposed.
	WSModules []string `toml:",omitempty"`

	// WSExposeAll exposes all API modules via the WebSocket RPC interface rather
	// than just the public ones.
	//
	// *WARNING* Only set this if the node is running in a trusted network, exposing
	// private APIs to untrusted users is a major security risk.
	WSExposeAll bool `toml:",omitempty"`

	// GRPCHost is the host interface on which to start the gRPC server. If
	// this field is empty, no gRPC API endpoint will be started.
	GRPCHost string `toml:",omitempty"`

	// GRPCPort is the TCP port number on which to start the gRPC server. The
	// default zero value is valid and will pick a port number randomly (useful for
	// ephemeral nodes).
	GRPCPort int `toml:",omitempty"`

	// UpstreamArchiveEN is an archive mode EN endpoint
	UpstreamArchiveEN string

	// Ntp server:port to check the synchronization when booting the node
	NtpRemoteServer string `toml:",omitempty"`

	// Disable option for unsafe debug APIs
	DisableUnsafeDebug bool `toml:",omitempty"`

	// Logger is a custom logger to use with the p2p.Server.
	Logger log.Logger `toml:",omitempty"`
}

// IPCEndpoint resolves an IPC endpoint based on a configured value, taking into
// account the set data folders as well as the designated platform we're currently
// running on.
func (c *Config) IPCEndpoint() string {
	// Short circuit if IPC has not been enabled
	if c.IPCPath == "" {
		return ""
	}
	// On windows we can only use plain top-level pipes
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(c.IPCPath, `\\.\pipe\`) {
			return c.IPCPath
		}
		return `\\.\pipe\` + c.IPCPath
	}
	// Resolve names into the data directory full paths otherwise
	if filepath.Base(c.IPCPath) == c.IPCPath {
		if c.DataDir == "" {
			return filepath.Join(os.TempDir(), c.IPCPath)
		}
		return filepath.Join(c.DataDir, c.IPCPath)
	}
	return c.IPCPath
}

// NodeDB returns the path to the discovery node database.
func (c *Config) NodeDB() string {
	if c.DataDir == "" {
		return "" // ephemeral
	}
	return c.ResolvePath(datadirNodeDatabase)
}

func DefaultIPCEndpoint(clientIdentifier string) string {
	if clientIdentifier == "" {
		clientIdentifier = strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if clientIdentifier == "" {
			panic("empty executable name")
		}
	}
	config := &Config{DataDir: DefaultDataDir(), IPCPath: clientIdentifier + ".ipc"}
	return config.IPCEndpoint()
}

// HTTPEndpoint resolves an HTTP endpoint based on the configured host interface
// and port parameters.
func (c *Config) HTTPEndpoint() string {
	if c.HTTPHost == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}

// DefaultHTTPEndpoint returns the HTTP endpoint used by default.
func DefaultHTTPEndpoint() string {
	config := &Config{HTTPHost: DefaultHTTPHost, HTTPPort: DefaultHTTPPort}
	return config.HTTPEndpoint()
}

// WSEndpoint resolves a websocket endpoint based on the configured host interface
// and port parameters.
func (c *Config) WSEndpoint() string {
	if c.WSHost == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", c.WSHost, c.WSPort)
}

// DefaultWSEndpoint returns the websocket endpoint used by default.
func DefaultWSEndpoint() string {
	config := &Config{WSHost: DefaultWSHost, WSPort: DefaultWSPort}
	return config.WSEndpoint()
}

// GRPCEndpoint resolves a gRPC endpoint based on the configured host interface
// and port parameters.
func (c *Config) GRPCEndpoint() string {
	if c.GRPCHost == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", c.GRPCHost, c.GRPCPort)
}

// DefaultGRPCEndpoint returns the gRPC endpoint used by default.
func DefaultGRPCEndpoint() string {
	config := &Config{GRPCHost: DefaultGRPCHost, GRPCPort: DefaultGRPCPort}
	return config.GRPCEndpoint()
}

// NodeName returns the devp2p node identifier.
func (c *Config) NodeName() string {
	name := c.name()
	// Backwards compatibility: previous versions used title-cased "Klaytn", keep that.
	if name == "klay" || name == "klay-testnet" {
		name = "Klaytn"
	}
	if c.UserIdent != "" {
		name += "/" + c.UserIdent
	}
	if c.Version != "" {
		name += "/" + c.Version
	}
	name += "/" + runtime.GOOS + "-" + runtime.GOARCH
	name += "/" + runtime.Version()
	return name
}

func (c *Config) name() string {
	if c.Name == "" {
		progname := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if progname == "" {
			panic("empty executable name, set Config.Name")
		}
		return progname
	}
	return c.Name
}

var isKaiaResource = map[string]bool{
	"chaindata":          true,
	"nodes":              true,
	"nodekey":            true,
	"bls-nodekey":        true,
	"static-nodes.json":  true,
	"trusted-nodes.json": true,
}

// ResolvePath resolves path in the instance directory.
func (c *Config) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if c.DataDir == "" {
		return ""
	}
	if c.name() == "klay" && isKaiaResource[path] {
		oldpath := ""
		if c.Name == "klay" {
			oldpath = filepath.Join(c.DataDir, path)
		}
		if c.ChainDataDir != "" && path == "chaindata" {
			return filepath.Join(c.ChainDataDir, path)
		}
		if oldpath != "" && common.FileExist(oldpath) {
			// TODO: print warning
			return oldpath
		}
	}
	return filepath.Join(c.instanceDir(), path)
}

func (c *Config) instanceDir() string {
	if c.DataDir == "" {
		return ""
	}
	return filepath.Join(c.DataDir, c.name())
}

func (c *Config) HttpServerType() string {
	if c.HTTPServerType == "" {
		return "http"
	}
	return c.HTTPServerType
}

// NodeKey retrieves the currently configured private key of the node, checking
// first any manually set key, falling back to the one found in the configured
// data folder. If no key can be found, a new one is generated.
func (c *Config) NodeKey() *ecdsa.PrivateKey {
	// Use any specifically configured key.
	if c.P2P.PrivateKey != nil {
		return c.P2P.PrivateKey
	}

	keyfile := c.ResolvePath(datadirPrivateKey)
	if key, err := crypto.LoadECDSA(keyfile); err == nil {
		return key
	}
	// No persistent key found, generate and store a new one.
	key, err := crypto.GenerateKey()
	if err != nil {
		logger.Crit("Failed to generate node key", "err", err)
	}
	instanceDir := filepath.Join(c.DataDir, c.name())
	if err := os.MkdirAll(instanceDir, 0o700); err != nil {
		logger.Crit("Failed to make dir to persist node key", "err", err)
	}
	keyfile = filepath.Join(instanceDir, datadirPrivateKey)
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		logger.Crit("Failed to persist node key", "err", err)
	}
	logger.Warn("Generated nodekey")
	return key
}

// BlsNodeKey retrieves the currently configured BLS secret key key of the node,
// check first any manually set key, falling back to the one found in the configured
// data folder. If no key can be found, derive from the NodeKey.
func (c *Config) BlsNodeKey() bls.SecretKey {
	// Manually set via flags --bls-nodekey or --bls-nodekeyhex
	if c.BlsKey != nil {
		return c.BlsKey
	}

	// Load from default location under datadir
	path := c.ResolvePath(DatadirBlsSecretKey)
	if key, err := bls.LoadKey(path); err == nil {
		return key
	}

	// No persistent key found, derive from NodeKey and store it
	key, err := bls.GenerateKey(crypto.FromECDSA(c.NodeKey()))
	if err != nil {
		logger.Crit("Failed to derive bls-nodekey from nodekey", "err", err)
	}
	instanceDir := filepath.Join(c.DataDir, c.name())
	if err := os.MkdirAll(instanceDir, 0o700); err != nil {
		logger.Crit("Failed to make dir to persist bls node key", "err", err)
	}
	keyfile := c.ResolvePath(DatadirBlsSecretKey)
	if err := bls.SaveKey(keyfile, key); err != nil {
		logger.Crit("Failed to persist bls node key", "err", err)
	}
	logger.Warn("Derived bls-nodekey from nodekey")
	return key
}

// StaticNodes returns a list of node enode URLs configured as static nodes.
func (c *Config) StaticNodes() []*discover.Node {
	return c.parsePersistentNodes(c.ResolvePath(datadirStaticNodes))
}

// TrustedNodes returns a list of node enode URLs configured as trusted nodes.
func (c *Config) TrustedNodes() []*discover.Node {
	return c.parsePersistentNodes(c.ResolvePath(datadirTrustedNodes))
}

// parsePersistentNodes parses a list of discovery node URLs loaded from a .json
// file from within the data directory.
func (c *Config) parsePersistentNodes(path string) []*discover.Node {
	// Short circuit if no node config is present
	if c.DataDir == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	// Load the nodes from the config file.
	var nodelist []string
	if err := common.LoadJSON(path, &nodelist); err != nil {
		logger.Error(fmt.Sprintf("Can't load node file %s: %v", path, err))
		return nil
	}
	// Interpret the list as a discovery node array
	var nodes []*discover.Node
	for _, url := range nodelist {
		if url == "" {
			continue
		}
		node, err := discover.ParseNode(url)
		if err != nil {
			logger.Error(fmt.Sprintf("Node URL %s: %v\n", url, err))
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// AccountConfig determines the settings for scrypt and keydirectory
func (c *Config) AccountConfig() (int, int, string, error) {
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP
	if c.UseLightweightKDF {
		scryptN = keystore.LightScryptN
		scryptP = keystore.LightScryptP
	}

	var (
		keydir string
		err    error
	)
	switch {
	case filepath.IsAbs(c.KeyStoreDir):
		keydir = c.KeyStoreDir
	case c.DataDir != "":
		if c.KeyStoreDir == "" {
			keydir = filepath.Join(c.DataDir, datadirDefaultKeyStore)
		} else {
			keydir, err = filepath.Abs(c.KeyStoreDir)
		}
	case c.KeyStoreDir != "":
		keydir, err = filepath.Abs(c.KeyStoreDir)
	}
	return scryptN, scryptP, keydir, err
}

func makeAccountManager(conf *Config) (*accounts.Manager, string, error) {
	scryptN, scryptP, keydir, err := conf.AccountConfig()
	var ephemeral string
	if keydir == "" {
		// There is no datadir.
		keydir, err = os.MkdirTemp("", "kaia-keystore")
		ephemeral = keydir
	}

	if err != nil {
		return nil, "", err
	}
	if err := os.MkdirAll(keydir, 0o700); err != nil {
		return nil, "", err
	}
	// Assemble the account manager and supported backends
	backends := []accounts.Backend{
		keystore.NewKeyStore(keydir, scryptN, scryptP),
	}
	return accounts.NewManager(backends...), ephemeral, nil
}
