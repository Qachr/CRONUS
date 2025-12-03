package IpfsLink

import (
	// "context"
	// "encoding/json"
	// "errors"
	// 	"fmt"
	// 	"io"
	// 	"io/fs"
	// 	"io/ioutil"
	// 	"os"

	"io"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"strconv"

	// 	"strconv"
	// 	"sync"
	// 	"time"

	// 	iface "github.com/ipfs/boxo/coreiface"

	// 	"github.com/ipfs/kubo/config"
	// 	"github.com/ipfs/kubo/core"
	// 	"github.com/ipfs/kubo/core/bootstrap"
	// 	"github.com/ipfs/kubo/core/coreapi"

	// 	"github.com/ipfs/kubo/plugin/loader"
	// 	"github.com/ipfs/kubo/repo/fsrepo"

	// 	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	// "github.com/ipfs/go-cid"
	// pubsub "github.com/libp2p/go-libp2p-pubsub"
	// "github.com/libp2p/go-libp2p/core/host"
	// "github.com/libp2p/go-libp2p/core/peer"
	// "github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	// 	"github.com/ipfs/go-cid"

	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"

	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	coreapi "github.com/ipfs/kubo/core/coreapi" // experimental API interface
	iface "github.com/ipfs/kubo/core/coreiface"

	"github.com/ipfs/kubo/plugin/loader"
	_ "github.com/ipfs/kubo/plugin/loader" // ensure built-in plugins are loaded
	"github.com/ipfs/kubo/repo/fsrepo"
	"github.com/libp2p/go-libp2p/core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	libp2pIFPS "github.com/ipfs/kubo/core/node/libp2p"
	// "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "pubsub-chat-example"

// printErr is like fmt.Printf, but writes to stderr.
func printErr(m string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, m, args...)
}

// defaultNick generates a nickname based on the $USER environment variable and
// the last 8 chars of a peer ID.
func defaultNick(p peer.ID) string {
	return fmt.Sprintf("%s-%s", os.Getenv("USER"), shortID(p))
}

// shortID returns the last 8 chars of a base58-encoded peer id.
func shortID(p peer.ID) string {
	pretty := p.ShortString()
	return pretty[len(pretty)-8:]
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.Addrs[0])
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.ShortString(), err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}

type MultiAddressesJson struct {
	PeerID      string   `json:"peer_id"`
	AddressList []string `json:"listened_addresses"`
}

type IpfsLink struct {
	Cancel          context.CancelFunc
	Ctx             context.Context
	IpfsCore        iface.CoreAPI
	IpfsNode        *core.IpfsNode
	Topics          []*pubsub.Topic
	Hst             host.Host
	GossipSub       *pubsub.PubSub
	Cr              *Client
	ParalelRetrieve bool
}

func InitNode(peerName string, bootstrapPeer string, ipfsBootstrap []byte, swarmKey bool, parallelRetrieve bool) (*IpfsLink, error) {
	ct, cancl := context.WithCancel(context.Background())

	// Spawn a local peer using a temporary path, for testing purposes
	// var idBootstrap peer.AddrInfo
	var ipfsA iface.CoreAPI
	var nodeA *core.IpfsNode
	var err error

	fmt.Printf("Bootstrap peer : %s\n Bootstrap Byte : %s\n", bootstrapPeer, ipfsBootstrap)
	if bootstrapPeer != "" {
		bootstrapConfig := ReadPeerInfo(ipfsBootstrap)

		fmt.Println("calling spawn ephermeral\n !!!!!!\n!!!!!!\n!!!!!!\n!!!!!!")
		ipfsA, nodeA, err = spawnEphemeral(ct, bootstrapConfig.AddressList, swarmKey)
	} else {
		fmt.Println("calling spawn ephermeral\n !!!!!!\n!!!!!!\n!!!!!!\n!!!!!!")
		ipfsA, nodeA, err = spawnEphemeral(ct, nil, swarmKey)

	}

	if err != nil {
		panic(fmt.Errorf("failed to spawn peer node: %s", err))
	}
	h := InitClient(peerName, bootstrapPeer)
	ipfs := IpfsLink{
		Cancel:          cancl,
		Ctx:             ct,
		IpfsCore:        ipfsA,
		IpfsNode:        nodeA,
		Hst:             nodeA.PeerHost,
		GossipSub:       h.Ps,
		Cr:              h,
		ParalelRetrieve: parallelRetrieve,
	}
	if bootstrapPeer != "" {
		connectToPeer(h.Host, string(ipfsBootstrap))
	}

	//fmt.Println(ipfs.IpfsNode.Peerstore.PeerInfo(ipfs.IpfsNode.PeerHost.ID()))
	return &ipfs, err
}

func WritePeerInfo(sys IpfsLink, file string) {
	// Get peerID from IPFS Key
	ipfsKey, err := sys.IpfsCore.Key().Self(context.Background())
	if err != nil {
		panic(err)
	}
	peerID := ipfsKey.ID().String()

	announcedAddressesList, err := sys.IpfsCore.Swarm().LocalAddrs(context.Background())
	if err != nil {
		panic(err)
	}

	// Store the announced addresses
	myPeerLocalInfo := MultiAddressesJson{
		PeerID:      peerID,
		AddressList: []string{},
	}
	for _, addr := range announcedAddressesList {
		myPeerLocalInfo.AddressList = append(myPeerLocalInfo.AddressList, addr.String()+"/p2p/"+peerID)
	}

	myPeerlocalInfoBytes, _ := json.Marshal(&myPeerLocalInfo)
	err = ioutil.WriteFile(file, myPeerlocalInfoBytes, 0644)

}

func ReadPeerInfo(data []byte) MultiAddressesJson {
	var addresses MultiAddressesJson

	err := json.Unmarshal(data, &addresses)
	if err != nil {
		panic(fmt.Errorf("could not unmarshal multiaddress file : %v", err.Error()))
	}

	return addresses

}

var loadPluginsOnce sync.Once

func setupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

// Get preferred outbound ip of this machine
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

var LoopBackAddresses = []string{
	"/ip4/127.0.0.1/ipcidr/8",
	"/ip6/::1/ipcidr/128",
}

const PORT = 0

func createTempRepo(BootstrapMultiAddrList []string) (string, error) {
	repoPath, err := os.MkdirTemp("", "ipfs-shell")

	// Create a default config with a new identity
	cfg, err := config.Init(io.Discard, 2048)
	if err != nil {
		return "", fmt.Errorf("failed to init config: %w", err)
	}

	// Set both ipv4 and ipv6 addresses.
	cfg.Addresses.Swarm = []string{
		fmt.Sprintf("/ip4/%s/tcp/%d", GetOutboundIP(), PORT),
		fmt.Sprintf("/ip4/%s/udp/%d/quic-v1", GetOutboundIP(), PORT),
		fmt.Sprintf("/ip4/%s/udp/%d/quic-v1/webtransport", GetOutboundIP(), PORT),
		// fmt.Sprintf("/ip6/::/tcp/%d", peerConfig.Port),
		// fmt.Sprintf("/ip6/::/udp/%d/quic-v1", peerConfig.Port),
		// fmt.Sprintf("/ip6/::/udp/%d/quic-v1/webtransport", peerConfig.Port),
	}

	//1 to ...
	// Validate peer addresses
	for _, addr := range BootstrapMultiAddrList {
		if _, err := peer.AddrInfoFromString(addr); err != nil {
			return "", fmt.Errorf("invalid bootstrap addr %s: %w", addr, err)
		}
	}

	cfg.Bootstrap = []string{}

	cfg.Discovery.MDNS.Enabled = false
	cfg.AutoNAT = config.AutoNATConfig{
		ServiceMode: config.AutoNATServiceEnabled,
	}
	// ... to 2

	cfg.Swarm = config.SwarmConfig{
		AddrFilters: LoopBackAddresses,
		RelayService: config.RelayService{
			Enabled: config.True,
		},
	}
	cfg.Addresses = config.Addresses{
		Swarm: []string{
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", PORT)},
		NoAnnounce: LoopBackAddresses,
	}
	cfg.AutoTLS.Enabled = config.False
	cfg.Swarm.Transports.Network.Websocket = config.False // No webocket in Private network

	cfg.Addresses.Gateway = config.Strings{"/ip4/0.0.0.0/tcp/8080"}
	cfg.Addresses.API = config.Strings{"/ip4/0.0.0.0/tcp/5001"}

	cfg.Datastore = config.DefaultDatastoreConfig()
	dataStoreFilePath := filepath.Join(repoPath, "datastore_spec")
	datastoreContent := map[string]interface{}{
		"mounts": []interface{}{
			map[string]interface{}{
				"mountpoint": "/blocks",
				"path":       "blocks",
				"shardFunc":  "/repo/flatfs/shard/v1/next-to-last/2",
				"type":       "flatfs",
			},
			map[string]interface{}{
				"mountpoint": "/",
				"path":       "datastore",
				"type":       "levelds",
			},
		},
		"type": "mount",
	}

	datastoreContentBytes, err := json.Marshal(datastoreContent)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(dataStoreFilePath, []byte(datastoreContentBytes), 0644); err != nil {
		panic(err)
	}

	myRepositoryVersion := []byte("14")
	if err = os.WriteFile(filepath.Join(repoPath, "version"), myRepositoryVersion, 0644); err != nil {
		panic(err)
	}

	// initRepo initializes a repo at the given path if it does not already exist.
	plugins, err := loader.NewPluginLoader("")
	if err != nil {
		return "", fmt.Errorf("failed to create plgin loader: %w", err)
	}
	if err := plugins.Initialize(); err != nil {
		return "", fmt.Errorf("failed to initialise plugins loader: %w", err)
	}
	if err := plugins.Inject(); err != nil {
		return "", fmt.Errorf("failed to inject plugins: %w", err)
	}

	if err := fsrepo.Init(repoPath, cfg); err != nil {
		return "", fmt.Errorf("failed to init fsrepo: %w", err)
	}

	return repoPath, nil
}

/// ------ Spawning the node

// Creates an IPFS node and returns its coreAPI
func createNode(ctx context.Context, repoPath string) (*core.IpfsNode, error) {

	// Open the repo
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2pIFPS.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		// Routing: libp2p.DHTClientOption, // This option sets the node to be a client DHT node (only fetching records)
		Repo:      repo,
		Permanent: true,
	}

	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Node in repo: %s", err)
	}
	return node, nil

}

// Spawns a node to be used just for this run (i.e. creates a tmp repo)
func spawnEphemeral(ctx context.Context, btstrap []string, swarmKey bool) (iface.CoreAPI, *core.IpfsNode, error) {
	// Create a Temporary Repo containing all configuration
	repoPath, err := createTempRepo(btstrap)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp repo: %s", err)
	}

	// Create an IPFS node
	printErr("repository : %s\n", repoPath)
	if swarmKey {
		os.WriteFile(repoPath+"/swarm.key", []byte("/key/swarm/psk/1.0.0/\n/base16/\nedd99a84bbdd5c9cfc06bcc039d219b1000885ecba26901c02e7c8792bfaaa70"), 0o600)
	}

	node, err := createNode(ctx, repoPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Node: %s", err)
	}

	api, err := coreapi.NewCoreAPI(node)

	if swarmKey {
		node.PNetFingerprint = []byte("4c7dc2a2735a84b4b11ff5b39225aa771cea1abd3acf9b98708a25f286df851c")
	}
	// Connect the node to the other private network nodes
	if btstrap != nil {
		fmt.Println("trying to going in btstrap loop")
		for _, addr := range btstrap {
			fmt.Println("trying to peer.AddrInfoFromString")
			addresses, err := peer.AddrInfoFromString(addr)
			if err != nil {
				panic(fmt.Errorf("addr from P2PADDR : %v \n", err.Error()))
			}
			fmt.Printf("providing : %s", addresses)

			fmt.Printf("Bootstrap peer ID : %s\n Address total : %s\n", addresses.ID.String(), addresses.String())
			err = api.Swarm().Connect(context.Background(), *addresses)
			if err != nil {
				panic(fmt.Errorf("ERROR Connect  : %v \n(or) %s\n", err.Error(), err.Error()))
			}

		}
	}

	return api, node, err
}

func AddIPFS(ipfs *IpfsLink, message []byte) (blocks.Block, error) {

	// wrap reader in Unixfs Add
	// fileAdder := ipfs.IpfsCore.Unixfs().Add

	hash, _ := mh.Sum(message, mh.SHA2_256, -1)

	c := cid.NewCidV1(cid.Raw, hash)

	blk, err := blocks.NewBlockWithCid(message, c)
	if err != nil {
		return nil, fmt.Errorf("newblock issue : %w", err)
	}

	err = ipfs.IpfsNode.Blockstore.Put(context.Background(), blk)
	if err != nil {
		return nil, fmt.Errorf("Blockstore Add: %w", err)
	}

	err = ipfs.IpfsNode.Routing.Provide(context.Background(), c, true)
	if err != nil {
		return nil, fmt.Errorf("Blockstore Add: %w", err)
	}

	var b blocks.Block = blk
	return b, nil

}

func connectToPeer(h host.Host, addrStr string) error {
	pi, err := peer.AddrInfoFromString(addrStr)
	if err != nil {
		return fmt.Errorf("AddrInfoFromString: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := h.Connect(ctx, *pi); err != nil {
		return fmt.Errorf("host.Connect(%s): %w", addrStr, err)
	}
	return nil
}

type CID struct{ str string }

func GetIPFS(ipfs *IpfsLink, cids []cid.Cid) ([]blocks.Block, time.Duration, time.Duration, error) {
	// Search for providersr

	ti := time.Now()

	for _, c := range cids {
		fmt.Println("Looking up providers for CID:", c)

		ctx, _ := context.WithTimeout(context.Background(), time.Second*100)

		// FindProvidersAsync returns a channel of peer.AddrInfo
		provCh := ipfs.IpfsNode.Routing.FindProvidersAsync(ctx, c, 100)

		// Consume the channel
		for p := range provCh {
			fmt.Printf("  Provider: %s, addrs: %v\n", p.ID, p.Addrs)
		}

	}

	timeSeekroviders := time.Since(ti)

	ti = time.Now()
	// retrieving the files
	var out []blocks.Block

	blocks := ipfs.IpfsNode.Blocks.GetBlocks(context.Background(), cids)

	for b := range blocks {
		out = append(out, b)
	}

	timeRetrieveFile := time.Since(ti)

	var err error
	var file *os.File
	if len(out) > 0 {

		file, err = os.OpenFile("node1/time/timeConcurrentRetrieve.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			panic(fmt.Errorf("could not Close Debug File in IPFSLink:: GetIPFS\nerror:%s", err))
		}

		file.WriteString("" +
			"===============================New Batch of Cid To Retrieve===============================\n")
	}

	file.WriteString("Got all the cids asked\n")
	if len(cids) > 0 {
		file.WriteString("\n" +
			"Nb of Cids: " + strconv.Itoa(len(cids)) + "\n" +
			"Time To seek: " + strconv.FormatInt(timeSeekroviders.Milliseconds(), 10) + " ms\n" + "Time To retrieve: " + strconv.FormatInt(timeRetrieveFile.Milliseconds(), 10) + " ms\n" +
			"=================================The end of CID retrieval=================================\n" +
			"\n" +
			"\n" +
			"\n")
		err = file.Close()
		if err != nil {
			panic(fmt.Errorf("could not Close Debug File in IPFSLink:: GetIPFS\nerror:%s", err))
		}
	} else {

		file.WriteString("\n" +
			fmt.Sprintf("Even if no CID Where downloaded, len(cids):%d", len(cids)) +
			"=================================The end of CID retrieval=================================\n")
	}

	return out, timeSeekroviders, timeRetrieveFile, nil
}

func PubIPFS(ipfs *IpfsLink, msg []byte) {
	ipfs.Cr.Publish(msg)
}
